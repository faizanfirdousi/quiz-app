package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"

	"kahootclone/internal/config"
	"kahootclone/internal/db"
	"kahootclone/internal/models"
	"kahootclone/internal/observability"
)

var (
	cfg      *config.Config
	dbClient *db.Client
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
}

type createSessionRequest struct {
	QuizID string `json:"quizId"`
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	requestID := uuid.New().String()
	ctx = observability.WithRequestID(ctx, requestID)

	userId, _ := event.RequestContext.Authorizer["userId"].(string)
	ctx = observability.WithUserID(ctx, userId)

	observability.Info(ctx, "creating session")

	var req createSessionRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return errorResponse(400, "VALIDATION_ERROR", "Invalid request body", requestID), nil
	}

	if req.QuizID == "" {
		return errorResponse(400, "VALIDATION_ERROR", "Quiz ID is required", requestID), nil
	}

	// Verify quiz exists and caller is the host
	quiz, err := dbClient.GetQuiz(ctx, req.QuizID)
	if err != nil {
		return errorResponse(500, "INTERNAL_ERROR", "Failed to retrieve quiz", requestID), nil
	}
	if quiz == nil {
		return errorResponse(404, "NOT_FOUND", "Quiz not found", requestID), nil
	}
	if quiz.HostUserID != userId {
		return errorResponse(403, "FORBIDDEN", "You don't own this quiz", requestID), nil
	}

	// Generate unique 6-digit PIN
	pin, err := generateUniquePIN(ctx)
	if err != nil {
		return errorResponse(500, "INTERNAL_ERROR", "Failed to generate PIN", requestID), nil
	}

	session := &models.Session{
		SessionID:            uuid.New().String(),
		PIN:                  pin,
		QuizID:               req.QuizID,
		HostUserID:           userId,
		Status:               models.SessionStatusLobby,
		CurrentQuestionIndex: 0,
		CreatedAt:            time.Now().UTC(),
	}

	if err := dbClient.CreateSession(ctx, session); err != nil {
		observability.Error(ctx, "failed to create session", "error", err.Error())
		return errorResponse(500, "INTERNAL_ERROR", "Failed to create session", requestID), nil
	}

	observability.Info(ctx, "session created", "sessionId", session.SessionID, "pin", session.PIN)
	return successResponse(201, session, requestID), nil
}

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
	return "", fmt.Errorf("failed to generate unique PIN after 10 attempts")
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
