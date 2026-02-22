package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"

	"kahootclone/internal/config"
	"kahootclone/internal/db"
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

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	requestID := uuid.New().String()
	ctx = observability.WithRequestID(ctx, requestID)

	userId, _ := event.RequestContext.Authorizer["userId"].(string)
	ctx = observability.WithUserID(ctx, userId)

	quizID := event.PathParameters["quizId"]
	if quizID == "" {
		return errorResponse(400, "VALIDATION_ERROR", "Quiz ID is required", requestID), nil
	}

	observability.Info(ctx, "getting quiz", "quizId", quizID)

	quiz, err := dbClient.GetQuiz(ctx, quizID)
	if err != nil {
		observability.Error(ctx, "failed to get quiz", "error", err.Error())
		return errorResponse(500, "INTERNAL_ERROR", "Failed to retrieve quiz", requestID), nil
	}
	if quiz == nil {
		return errorResponse(404, "NOT_FOUND", "Quiz not found", requestID), nil
	}
	if quiz.HostUserID != userId {
		return errorResponse(403, "FORBIDDEN", "You don't have access to this quiz", requestID), nil
	}

	return successResponse(200, quiz, requestID), nil
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
