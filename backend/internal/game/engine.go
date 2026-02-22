package game

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"kahootclone/internal/cache"
	"kahootclone/internal/db"
	"kahootclone/internal/models"
	"kahootclone/internal/observability"
)

// GameState represents the current phase of a game.
type GameState string

const (
	StateLobby          GameState = "LOBBY"
	StateActive         GameState = "ACTIVE"
	StateQuestionOpen   GameState = "QUESTION_OPEN"
	StateQuestionClosed GameState = "QUESTION_CLOSED"
	StateLeaderboard    GameState = "LEADERBOARD"
	StateFinished       GameState = "FINISHED"
)

// Engine manages the game state machine.
type Engine struct {
	DB          *db.Client
	Cache       *cache.RedisClient
	Broadcaster *Broadcaster
}

// NewEngine creates a new game engine.
func NewEngine(dbClient *db.Client, cacheClient *cache.RedisClient, broadcaster *Broadcaster) *Engine {
	return &Engine{
		DB:          dbClient,
		Cache:       cacheClient,
		Broadcaster: broadcaster,
	}
}

// HandleJoinSession processes a player joining a session via WebSocket.
func (e *Engine) HandleJoinSession(ctx context.Context, connectionID string, payload models.JoinSessionPayload) error {
	observability.Info(ctx, "player joining session",
		"sessionId", payload.SessionID,
		"nickname", payload.Nickname,
		"connectionId", connectionID,
	)

	// Validate nickname length
	if len(payload.Nickname) == 0 || len(payload.Nickname) > 20 {
		return fmt.Errorf("nickname must be between 1 and 20 characters")
	}

	// Get session
	session, err := e.DB.GetSession(ctx, payload.SessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return fmt.Errorf("session not found")
	}
	if session.Status != models.SessionStatusLobby {
		return fmt.Errorf("game already started")
	}

	// Check player count
	count, err := e.DB.GetPlayerCountBySession(ctx, payload.SessionID)
	if err != nil {
		return fmt.Errorf("failed to get player count: %w", err)
	}
	if count >= 2000 {
		return fmt.Errorf("session is full (max 2000 players)")
	}

	// Extract userId from context (set by auth middleware or WS auth)
	claims := getClaimsFromContext(ctx)
	userID := "anonymous"
	if claims != nil {
		userID = claims.UserID
	}

	// Register connection
	player := &models.Player{
		SessionID:    payload.SessionID,
		ConnectionID: connectionID,
		UserID:       userID,
		Nickname:     payload.Nickname,
		Role:         models.PlayerRolePlayer,
		ConnectedAt:  time.Now().UTC(),
	}
	if err := e.DB.PutConnection(ctx, player); err != nil {
		return fmt.Errorf("failed to register connection: %w", err)
	}

	// Initialize score in leaderboard
	if err := e.Cache.UpsertScore(ctx, payload.SessionID, userID, 0); err != nil {
		slog.Warn("failed to initialize score in Redis", "error", err.Error())
	}
	if err := e.Cache.SetNickname(ctx, payload.SessionID, userID, payload.Nickname); err != nil {
		slog.Warn("failed to set nickname in Redis", "error", err.Error())
	}

	// Broadcast player joined
	newCount := count + 1
	return e.Broadcaster.BroadcastToSession(ctx, payload.SessionID, models.WSOutbound{
		Type: models.WSTypePlayerJoined,
		Payload: models.PlayerJoinedPayload{
			Nickname:    payload.Nickname,
			PlayerCount: newCount,
		},
	})
}

// HandleStartGame processes the host starting the game.
func (e *Engine) HandleStartGame(ctx context.Context, connectionID string, payload models.StartGamePayload) error {
	observability.Info(ctx, "host starting game", "sessionId", payload.SessionID)

	session, err := e.DB.GetSession(ctx, payload.SessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return fmt.Errorf("session not found")
	}
	if session.Status != models.SessionStatusLobby {
		return fmt.Errorf("game already started")
	}

	// Verify caller is host
	conn, err := e.DB.GetSessionByConnectionID(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("failed to verify host: %w", err)
	}
	if conn.Role != models.PlayerRoleHost {
		return fmt.Errorf("only the host can start the game")
	}

	// Get quiz
	quiz, err := e.DB.GetQuiz(ctx, session.QuizID)
	if err != nil {
		return fmt.Errorf("failed to get quiz: %w", err)
	}
	if quiz == nil {
		return fmt.Errorf("quiz not found")
	}

	// Update session status
	if err := e.DB.UpdateSessionStatus(ctx, payload.SessionID, models.SessionStatusActive, 0); err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	// Broadcast game started
	if err := e.Broadcaster.BroadcastToSession(ctx, payload.SessionID, models.WSOutbound{
		Type: models.WSTypeGameStarted,
		Payload: models.GameStartedPayload{
			TotalQuestions: len(quiz.Questions),
		},
	}); err != nil {
		return err
	}

	// Send first question
	return e.sendQuestion(ctx, payload.SessionID, quiz, 0)
}

// HandleSubmitAnswer processes a player's answer submission.
func (e *Engine) HandleSubmitAnswer(ctx context.Context, connectionID string, payload models.SubmitAnswerPayload) error {
	observability.Info(ctx, "answer submitted", "connectionId", connectionID, "questionId", payload.QuestionID)

	// Find the player's connection info
	conn, err := e.DB.GetSessionByConnectionID(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("failed to find connection: %w", err)
	}

	session, err := e.DB.GetSession(ctx, conn.SessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil || session.Status != models.SessionStatusActive {
		return fmt.Errorf("game is not active")
	}

	// Check if already answered
	existing, err := e.DB.GetAnswer(ctx, conn.SessionID, conn.UserID, payload.QuestionID)
	if err != nil {
		return fmt.Errorf("failed to check existing answer: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("already answered this question")
	}

	// Get quiz for correct answer
	quiz, err := e.DB.GetQuiz(ctx, session.QuizID)
	if err != nil {
		return fmt.Errorf("failed to get quiz: %w", err)
	}

	// Find the current question
	var question *models.Question
	for i := range quiz.Questions {
		if quiz.Questions[i].QuestionID == payload.QuestionID {
			question = &quiz.Questions[i]
			break
		}
	}
	if question == nil {
		return fmt.Errorf("question not found")
	}

	// Calculate score
	isCorrect := payload.SelectedOptionID == question.CorrectOptionID
	timeLimitMs := int64(question.TimeLimitSeconds * 1000)
	pointsEarned := CalculateScore(isCorrect, payload.TimeTakenMs, timeLimitMs, question.Points)

	// Store answer
	answer := &models.Answer{
		SessionID:        conn.SessionID,
		UserIDQuestionID: conn.UserID + "#" + payload.QuestionID,
		QuestionID:       payload.QuestionID,
		UserID:           conn.UserID,
		SelectedOptionID: payload.SelectedOptionID,
		IsCorrect:        isCorrect,
		TimeTakenMs:      payload.TimeTakenMs,
		PointsEarned:     pointsEarned,
		AnsweredAt:       time.Now().UTC(),
	}
	if err := e.DB.PutAnswer(ctx, answer); err != nil {
		return fmt.Errorf("failed to store answer: %w", err)
	}

	// Update leaderboard
	if pointsEarned > 0 {
		if err := e.Cache.IncrementScore(ctx, conn.SessionID, conn.UserID, float64(pointsEarned)); err != nil {
			slog.Warn("failed to update leaderboard", "error", err.Error())
		}
	}

	// Get updated rank and total score
	rank, _ := e.Cache.GetPlayerRank(ctx, conn.SessionID, conn.UserID)
	totalScore, _ := e.Cache.GetPlayerScore(ctx, conn.SessionID, conn.UserID)

	// Send personal result to the player
	return e.Broadcaster.SendToConnection(ctx, connectionID, models.WSOutbound{
		Type: models.WSTypeAnswerResult,
		Payload: models.AnswerResultPayload{
			IsCorrect:     isCorrect,
			PointsEarned:  pointsEarned,
			TotalScore:    int(totalScore),
			Rank:          rank,
			CorrectOption: question.CorrectOptionID,
		},
	})
}

// HandleNextQuestion sends the next question or ends the game.
func (e *Engine) HandleNextQuestion(ctx context.Context, connectionID string, payload models.NextQuestionPayload) error {
	observability.Info(ctx, "next question requested", "sessionId", payload.SessionID)

	// Verify host
	conn, err := e.DB.GetSessionByConnectionID(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("failed to verify host: %w", err)
	}
	if conn.Role != models.PlayerRoleHost {
		return fmt.Errorf("only the host can advance questions")
	}

	session, err := e.DB.GetSession(ctx, payload.SessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	quiz, err := e.DB.GetQuiz(ctx, session.QuizID)
	if err != nil {
		return fmt.Errorf("failed to get quiz: %w", err)
	}

	nextIndex := session.CurrentQuestionIndex + 1

	// Send current question's end results first
	leaderboard, _ := e.Cache.GetTopN(ctx, payload.SessionID, 10)
	currentQuestion := quiz.Questions[session.CurrentQuestionIndex]
	_ = e.Broadcaster.BroadcastToSession(ctx, payload.SessionID, models.WSOutbound{
		Type: models.WSTypeQuestionEnded,
		Payload: models.QuestionEndedPayload{
			CorrectOption: currentQuestion.CorrectOptionID,
			Leaderboard:   leaderboard,
		},
	})

	if nextIndex >= len(quiz.Questions) {
		return e.endGame(ctx, payload.SessionID)
	}

	// Update session
	if err := e.DB.UpdateSessionStatus(ctx, payload.SessionID, models.SessionStatusActive, nextIndex); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return e.sendQuestion(ctx, payload.SessionID, quiz, nextIndex)
}

// HandleEndGame ends the game early.
func (e *Engine) HandleEndGame(ctx context.Context, connectionID string, payload models.EndGamePayload) error {
	observability.Info(ctx, "end game requested", "sessionId", payload.SessionID)

	conn, err := e.DB.GetSessionByConnectionID(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("failed to verify host: %w", err)
	}
	if conn.Role != models.PlayerRoleHost {
		return fmt.Errorf("only the host can end the game")
	}

	return e.endGame(ctx, payload.SessionID)
}

// HandleMessage routes an incoming WebSocket message to the appropriate handler.
func (e *Engine) HandleMessage(ctx context.Context, connectionID string, rawMessage []byte) error {
	var msg models.WSInbound
	if err := json.Unmarshal(rawMessage, &msg); err != nil {
		return fmt.Errorf("invalid message format: %w", err)
	}

	observability.Debug(ctx, "handling WS message", "action", msg.Action, "connectionId", connectionID)

	switch msg.Action {
	case models.WSActionJoinSession:
		var payload models.JoinSessionPayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			return fmt.Errorf("invalid join_session payload: %w", err)
		}
		return e.HandleJoinSession(ctx, connectionID, payload)

	case models.WSActionSubmitAnswer:
		var payload models.SubmitAnswerPayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			return fmt.Errorf("invalid submit_answer payload: %w", err)
		}
		return e.HandleSubmitAnswer(ctx, connectionID, payload)

	case models.WSActionStartGame:
		var payload models.StartGamePayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			return fmt.Errorf("invalid start_game payload: %w", err)
		}
		return e.HandleStartGame(ctx, connectionID, payload)

	case models.WSActionNextQuestion:
		var payload models.NextQuestionPayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			return fmt.Errorf("invalid next_question payload: %w", err)
		}
		return e.HandleNextQuestion(ctx, connectionID, payload)

	case models.WSActionEndGame:
		var payload models.EndGamePayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			return fmt.Errorf("invalid end_game payload: %w", err)
		}
		return e.HandleEndGame(ctx, connectionID, payload)

	default:
		return fmt.Errorf("unknown action: %s", msg.Action)
	}
}

func (e *Engine) sendQuestion(ctx context.Context, sessionID string, quiz *models.Quiz, index int) error {
	q := quiz.Questions[index]

	return e.Broadcaster.BroadcastToSession(ctx, sessionID, models.WSOutbound{
		Type: models.WSTypeQuestion,
		Payload: models.QuestionPayload{
			QuestionIndex:  index,
			TotalQuestions: len(quiz.Questions),
			Text:           q.Text,
			Options:        q.Options, // correctOptionId is NOT included in QuestionPayload
			TimeLimitMs:    q.TimeLimitSeconds * 1000,
			Points:         q.Points,
		},
	})
}

func (e *Engine) endGame(ctx context.Context, sessionID string) error {
	observability.Info(ctx, "ending game", "sessionId", sessionID)

	if err := e.DB.UpdateSessionStatus(ctx, sessionID, models.SessionStatusFinished, -1); err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	leaderboard, _ := e.Cache.GetTopN(ctx, sessionID, 100)

	if err := e.Broadcaster.BroadcastToSession(ctx, sessionID, models.WSOutbound{
		Type: models.WSTypeGameOver,
		Payload: models.GameOverPayload{
			FinalLeaderboard: leaderboard,
		},
	}); err != nil {
		return err
	}

	// Clean up Redis (async-safe — if it fails, TTL will eventually clean it)
	go func() {
		_ = e.Cache.DeleteSession(context.Background(), sessionID)
	}()

	return nil
}

// getClaimsFromContext is a helper to pull auth claims from context.
// This is set by the auth middleware or WebSocket authentication.
func getClaimsFromContext(ctx context.Context) *Claims {
	type claimsKey string
	claims, _ := ctx.Value(claimsKey("userClaims")).(*Claims)
	return claims
}

// Claims mirrors auth.Claims for use within the game package.
type Claims struct {
	UserID   string
	Email    string
	Username string
	Role     string
}

// GenerateSessionPIN generates a random 6-digit PIN.
func GenerateSessionPIN() string {
	id := uuid.New()
	// Use first 6 hex chars converted to digits
	pin := fmt.Sprintf("%06d", int(id.ID())%1000000)
	return pin
}
