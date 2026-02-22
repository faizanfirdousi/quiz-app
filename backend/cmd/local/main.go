package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"kahootclone/internal/auth"
	"kahootclone/internal/cache"
	"kahootclone/internal/config"
	"kahootclone/internal/db"
	"kahootclone/internal/game"
	"kahootclone/internal/models"
	"kahootclone/internal/observability"
)

var (
	cfg         *config.Config
	dbClient    *db.Client
	redisClient *cache.RedisClient
	validator   *auth.CognitoValidator
	gameEngine  *game.Engine
	broadcaster *game.Broadcaster
	hub         *game.Hub
	upgrader    = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true }, // Allow all origins in local dev
	}
)

func main() {
	cfg = config.Load()
	observability.InitLogger(cfg.LogLevel, cfg.Env)
	observability.InitTracer(cfg.Env)

	slog.Info("starting KahootClone local server", "env", cfg.Env, "port", cfg.Port)

	// Initialize DynamoDB
	var err error
	dbClient, err = db.NewClient(context.Background(), cfg)
	if err != nil {
		slog.Error("failed to initialize DynamoDB client", "error", err.Error())
		os.Exit(1)
	}

	// Initialize Redis
	redisClient, err = cache.NewRedisClient(context.Background(), cfg)
	if err != nil {
		slog.Error("failed to initialize Redis client", "error", err.Error())
		os.Exit(1)
	}
	defer redisClient.Close()

	// Initialize Cognito validator (async — don't block startup)
	validator = auth.NewCognitoValidator(cfg.CognitoRegion, cfg.CognitoUserPoolID, cfg.CognitoClientID)
	validator.InitAsync()
	slog.Info("Cognito JWKS fetch started in background")

	// Initialize game engine
	hub = game.NewHub()
	broadcaster = game.NewBroadcaster(dbClient, cfg.Env)
	broadcaster.SetHub(hub)
	gameEngine = game.NewEngine(dbClient, redisClient, broadcaster)

	// Setup routes
	mux := http.NewServeMux()

	// Health check (no auth)
	mux.HandleFunc("GET /health", handleHealth)

	// WebSocket endpoint (auth via query param)
	mux.HandleFunc("/ws", handleWebSocket)

	// REST API routes (with auth middleware)
	authMiddleware := auth.Middleware(validator)

	mux.Handle("POST /api/quizzes", authMiddleware(http.HandlerFunc(handleCreateQuiz)))
	mux.Handle("GET /api/quizzes/{quizId}", authMiddleware(http.HandlerFunc(handleGetQuiz)))
	mux.Handle("POST /api/sessions", authMiddleware(http.HandlerFunc(handleCreateSession)))
	mux.Handle("POST /api/sessions/{sessionId}/join", authMiddleware(http.HandlerFunc(handleJoinSession)))
	mux.Handle("GET /api/sessions/{sessionId}/leaderboard", authMiddleware(http.HandlerFunc(handleGetLeaderboard)))

	// Note: CORS preflight is handled by corsMiddleware, no need for explicit OPTIONS route
	// Wrap with logging middleware
	handler := loggingMiddleware(corsMiddleware(mux))

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		slog.Info("shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	slog.Info("server listening", "addr", ":"+cfg.Port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("server error", "error", err.Error())
		os.Exit(1)
	}
}

// --- Health ---

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]interface{}{
		"status": "ok",
		"env":    cfg.Env,
	})
}

// --- WebSocket ---

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Authenticate via query param
	token := r.URL.Query().Get("token")
	sessionID := r.URL.Query().Get("sessionId")
	role := r.URL.Query().Get("role")

	if sessionID == "" {
		http.Error(w, "sessionId query param required", http.StatusBadRequest)
		return
	}

	var userID string
	if token != "" {
		claims, err := validator.ValidateToken(r.Context(), token)
		if err != nil {
			slog.Warn("WS auth failed", "error", err.Error())
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		userID = claims.UserID
	} else {
		userID = "anon-" + uuid.New().String()[:8]
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WS upgrade failed", "error", err.Error())
		return
	}

	connectionID := uuid.New().String()

	// Register in hub
	hub.Register(connectionID, sessionID, conn)

	// Register in DynamoDB
	playerRole := models.PlayerRolePlayer
	if role == "HOST" {
		playerRole = models.PlayerRoleHost
	}
	player := &models.Player{
		SessionID:    sessionID,
		ConnectionID: connectionID,
		UserID:       userID,
		Role:         playerRole,
		ConnectedAt:  time.Now().UTC(),
	}
	if err := dbClient.PutConnection(r.Context(), player); err != nil {
		slog.Error("failed to register WS connection in DB", "error", err.Error())
	}

	slog.Info("WS connected", "connectionId", connectionID, "sessionId", sessionID, "userId", userID)

	// Read messages in a goroutine
	go func() {
		defer func() {
			hub.Unregister(connectionID)
			_ = dbClient.DeleteConnection(context.Background(), sessionID, connectionID)
			conn.Close()
			slog.Info("WS disconnected", "connectionId", connectionID)
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					slog.Warn("WS unexpected close", "connectionId", connectionID, "error", err.Error())
				}
				break
			}

			ctx := observability.WithRequestID(context.Background(), uuid.New().String())
			ctx = observability.WithUserID(ctx, userID)
			ctx = observability.WithSessionID(ctx, sessionID)

			if handleErr := gameEngine.HandleMessage(ctx, connectionID, message); handleErr != nil {
				slog.Error("WS message error", "connectionId", connectionID, "error", handleErr.Error())
				// Send error back to client
				errPayload := models.WSOutbound{
					Type: models.WSTypeError,
					Payload: models.ErrorPayload{
						Code:    "INTERNAL_ERROR",
						Message: handleErr.Error(),
					},
				}
				data, _ := json.Marshal(errPayload)
				conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	}()
}

// --- REST Handlers ---

func handleCreateQuiz(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	claims := auth.GetClaims(r.Context())

	var req struct {
		Title       string            `json:"title"`
		Description string            `json:"description"`
		Questions   []models.Question `json:"questions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "VALIDATION_ERROR", "Invalid request body", requestID)
		return
	}
	if req.Title == "" {
		writeError(w, 400, "VALIDATION_ERROR", "Title is required", requestID)
		return
	}
	if len(req.Questions) == 0 {
		writeError(w, 400, "VALIDATION_ERROR", "At least one question is required", requestID)
		return
	}
	if len(req.Questions) > 100 {
		writeError(w, 400, "VALIDATION_ERROR", "Maximum 100 questions per quiz", requestID)
		return
	}

	for i := range req.Questions {
		if req.Questions[i].QuestionID == "" {
			req.Questions[i].QuestionID = uuid.New().String()
		}
		for j := range req.Questions[i].Options {
			if req.Questions[i].Options[j].ID == "" {
				req.Questions[i].Options[j].ID = uuid.New().String()
			}
		}
	}

	now := time.Now().UTC()
	quiz := &models.Quiz{
		QuizID:      uuid.New().String(),
		HostUserID:  claims.UserID,
		Title:       req.Title,
		Description: req.Description,
		Questions:   req.Questions,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := dbClient.CreateQuiz(r.Context(), quiz); err != nil {
		slog.Error("failed to create quiz", "error", err.Error())
		writeError(w, 500, "INTERNAL_ERROR", "Failed to create quiz", requestID)
		return
	}

	writeSuccess(w, 201, quiz, requestID)
}

func handleGetQuiz(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	claims := auth.GetClaims(r.Context())
	quizID := r.PathValue("quizId")

	quiz, err := dbClient.GetQuiz(r.Context(), quizID)
	if err != nil {
		writeError(w, 500, "INTERNAL_ERROR", "Failed to retrieve quiz", requestID)
		return
	}
	if quiz == nil {
		writeError(w, 404, "NOT_FOUND", "Quiz not found", requestID)
		return
	}
	if quiz.HostUserID != claims.UserID {
		writeError(w, 403, "FORBIDDEN", "You don't have access to this quiz", requestID)
		return
	}

	writeSuccess(w, 200, quiz, requestID)
}

func handleCreateSession(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	claims := auth.GetClaims(r.Context())

	var req struct {
		QuizID string `json:"quizId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "VALIDATION_ERROR", "Invalid request body", requestID)
		return
	}
	if req.QuizID == "" {
		writeError(w, 400, "VALIDATION_ERROR", "Quiz ID is required", requestID)
		return
	}

	quiz, err := dbClient.GetQuiz(r.Context(), req.QuizID)
	if err != nil || quiz == nil {
		writeError(w, 404, "NOT_FOUND", "Quiz not found", requestID)
		return
	}
	if quiz.HostUserID != claims.UserID {
		writeError(w, 403, "FORBIDDEN", "You don't own this quiz", requestID)
		return
	}

	pin, err := generateUniquePIN(r.Context())
	if err != nil {
		writeError(w, 500, "INTERNAL_ERROR", "Failed to generate PIN", requestID)
		return
	}

	session := &models.Session{
		SessionID:            uuid.New().String(),
		PIN:                  pin,
		QuizID:               req.QuizID,
		HostUserID:           claims.UserID,
		Status:               models.SessionStatusLobby,
		CurrentQuestionIndex: 0,
		CreatedAt:            time.Now().UTC(),
	}

	if err := dbClient.CreateSession(r.Context(), session); err != nil {
		writeError(w, 500, "INTERNAL_ERROR", "Failed to create session", requestID)
		return
	}

	writeSuccess(w, 201, session, requestID)
}

func handleJoinSession(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	claims := auth.GetClaims(r.Context())
	sessionID := r.PathValue("sessionId")

	var req struct {
		Nickname string `json:"nickname"`
		PIN      string `json:"pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "VALIDATION_ERROR", "Invalid request body", requestID)
		return
	}
	if req.Nickname == "" || len(req.Nickname) > 20 {
		writeError(w, 400, "VALIDATION_ERROR", "Nickname must be 1-20 characters", requestID)
		return
	}

	// Look up by PIN if sessionId not provided
	if sessionID == "" && req.PIN != "" {
		session, err := dbClient.GetSessionByPIN(r.Context(), req.PIN)
		if err != nil || session == nil {
			writeError(w, 404, "NOT_FOUND", "No session found with this PIN", requestID)
			return
		}
		sessionID = session.SessionID
	}

	session, err := dbClient.GetSession(r.Context(), sessionID)
	if err != nil || session == nil {
		writeError(w, 404, "NOT_FOUND", "Session not found", requestID)
		return
	}
	if session.Status != models.SessionStatusLobby {
		writeError(w, 409, "GAME_ALREADY_STARTED", "This game has already started", requestID)
		return
	}

	count, _ := dbClient.GetPlayerCountBySession(r.Context(), sessionID)
	if count >= 2000 {
		writeError(w, 409, "SESSION_FULL", "Session is full (max 2000 players)", requestID)
		return
	}

	_ = redisClient.UpsertScore(r.Context(), sessionID, claims.UserID, 0)
	_ = redisClient.SetNickname(r.Context(), sessionID, claims.UserID, req.Nickname)

	writeSuccess(w, 200, map[string]interface{}{
		"sessionId": sessionID,
		"pin":       session.PIN,
		"nickname":  req.Nickname,
	}, requestID)
}

func handleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	sessionID := r.PathValue("sessionId")

	leaderboard, err := redisClient.GetTopN(r.Context(), sessionID, 100)
	if err != nil {
		slog.Warn("failed to get leaderboard", "error", err.Error())
	}

	writeSuccess(w, 200, map[string]interface{}{
		"sessionId":   sessionID,
		"leaderboard": leaderboard,
	}, requestID)
}

func handleCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(204)
}

// --- Helpers ---

func generateUniquePIN(ctx context.Context) (string, error) {
	for attempt := 0; attempt < 10; attempt++ {
		pin := fmt.Sprintf("%06d", rand.Intn(1000000))
		existing, err := dbClient.GetSessionByPIN(ctx, pin)
		if err != nil {
			return "", err
		}
		if existing == nil || existing.Status == models.SessionStatusFinished {
			return pin, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique PIN")
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeSuccess(w http.ResponseWriter, status int, data interface{}, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"data":      data,
		"requestId": requestID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func writeError(w http.ResponseWriter, status int, code, message, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   false,
		"error":     map[string]string{"code": code, "message": message},
		"requestId": requestID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := uuid.New().String()

		// Wrap response writer to capture status code
		wrapped := &statusWriter{ResponseWriter: w, status: 200}

		ctx := observability.WithRequestID(r.Context(), requestID)
		r = r.WithContext(ctx)

		next.ServeHTTP(wrapped, r)

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.status,
			"latency", time.Since(start).String(),
			"requestId", requestID,
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
