package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"

	"kahootclone/internal/cache"
	"kahootclone/internal/config"
	"kahootclone/internal/db"
	"kahootclone/internal/models"
	"kahootclone/internal/observability"
)

var (
	cfg         *config.Config
	dbClient    *db.Client
	redisClient *cache.RedisClient
)

func init() {
	cfg = config.Load()
	observability.InitLogger(cfg.LogLevel, cfg.Env)
	observability.InitTracer(cfg.Env)

	var err error
	dbClient, err = db.NewClient(context.Background(), cfg)
	if err != nil {
		slog.Error("failed to initialize DynamoDB client", "error", err.Error())
		panic(err)
	}
	redisClient, err = cache.NewRedisClient(context.Background(), cfg)
	if err != nil {
		slog.Error("failed to initialize Redis client", "error", err.Error())
		panic(err)
	}
}

type joinSessionRequest struct {
	Nickname string `json:"nickname"`
	PIN      string `json:"pin"` // Players use PIN to find session
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	requestID := uuid.New().String()
	ctx = observability.WithRequestID(ctx, requestID)

	userId, _ := event.RequestContext.Authorizer["userId"].(string)
	ctx = observability.WithUserID(ctx, userId)

	sessionID := event.PathParameters["sessionId"]

	observability.Info(ctx, "joining session", "sessionId", sessionID)

	var req joinSessionRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return errorResponse(400, "VALIDATION_ERROR", "Invalid request body", requestID), nil
	}

	if req.Nickname == "" || len(req.Nickname) > 20 {
		return errorResponse(400, "VALIDATION_ERROR", "Nickname must be between 1 and 20 characters", requestID), nil
	}

	// If sessionID is not provided, look up by PIN
	if sessionID == "" && req.PIN != "" {
		session, err := dbClient.GetSessionByPIN(ctx, req.PIN)
		if err != nil {
			return errorResponse(500, "INTERNAL_ERROR", "Failed to look up session", requestID), nil
		}
		if session == nil {
			return errorResponse(404, "NOT_FOUND", "No session found with this PIN", requestID), nil
		}
		sessionID = session.SessionID
	}

	session, err := dbClient.GetSession(ctx, sessionID)
	if err != nil {
		return errorResponse(500, "INTERNAL_ERROR", "Failed to retrieve session", requestID), nil
	}
	if session == nil {
		return errorResponse(404, "NOT_FOUND", "Session not found", requestID), nil
	}
	if session.Status != models.SessionStatusLobby {
		return errorResponse(409, "GAME_ALREADY_STARTED", "This game has already started", requestID), nil
	}

	// Check player count
	count, err := dbClient.GetPlayerCountBySession(ctx, sessionID)
	if err != nil {
		return errorResponse(500, "INTERNAL_ERROR", "Failed to check player count", requestID), nil
	}
	if count >= 2000 {
		return errorResponse(409, "SESSION_FULL", "This session is full (max 2000 players)", requestID), nil
	}

	// Initialize player in leaderboard
	if err := redisClient.UpsertScore(ctx, sessionID, userId, 0); err != nil {
		slog.Warn("failed to initialize leaderboard score", "error", err.Error())
	}
	if err := redisClient.SetNickname(ctx, sessionID, userId, req.Nickname); err != nil {
		slog.Warn("failed to set nickname", "error", err.Error())
	}

	response := map[string]interface{}{
		"sessionId": sessionID,
		"pin":       session.PIN,
		"nickname":  req.Nickname,
		"quizId":    session.QuizID,
	}

	observability.Info(ctx, "player joined session", "sessionId", sessionID, "nickname", req.Nickname)
	return successResponse(200, response, requestID), nil
}

func successResponse(statusCode int, data interface{}, requestID string) events.APIGatewayProxyResponse {
	body, _ := json.Marshal(map[string]interface{}{
		"success":   true,
		"data":      data,
		"requestId": requestID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type,Authorization",
		},
		Body: string(body),
	}
}

func errorResponse(statusCode int, code, message, requestID string) events.APIGatewayProxyResponse {
	body, _ := json.Marshal(map[string]interface{}{
		"success":   false,
		"error":     map[string]string{"code": code, "message": message},
		"requestId": requestID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type,Authorization",
		},
		Body: string(body),
	}
}

func main() {
	lambda.Start(handler)
}
