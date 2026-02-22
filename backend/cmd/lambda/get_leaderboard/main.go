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

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	requestID := uuid.New().String()
	ctx = observability.WithRequestID(ctx, requestID)

	sessionID := event.PathParameters["sessionId"]
	if sessionID == "" {
		return errorResponse(400, "VALIDATION_ERROR", "Session ID is required", requestID), nil
	}

	observability.Info(ctx, "getting leaderboard", "sessionId", sessionID)

	// Try Redis first for real-time leaderboard
	topN := 100
	leaderboard, err := redisClient.GetTopN(ctx, sessionID, topN)
	if err != nil {
		slog.Warn("failed to get leaderboard from Redis, falling back to DynamoDB", "error", err.Error())
		// Could fall back to computing from answers table if needed
		leaderboard = nil
	}

	response := map[string]interface{}{
		"sessionId":   sessionID,
		"leaderboard": leaderboard,
	}

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
