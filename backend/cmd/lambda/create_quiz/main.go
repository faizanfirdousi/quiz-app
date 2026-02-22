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

type createQuizRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Questions   []models.Question `json:"questions"`
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	requestID := uuid.New().String()
	ctx = observability.WithRequestID(ctx, requestID)

	userId, _ := event.RequestContext.Authorizer["userId"].(string)
	ctx = observability.WithUserID(ctx, userId)

	observability.Info(ctx, "creating quiz")

	var req createQuizRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return errorResponse(400, "VALIDATION_ERROR", "Invalid request body", requestID), nil
	}

	if req.Title == "" {
		return errorResponse(400, "VALIDATION_ERROR", "Title is required", requestID), nil
	}
	if len(req.Questions) == 0 {
		return errorResponse(400, "VALIDATION_ERROR", "At least one question is required", requestID), nil
	}
	if len(req.Questions) > 100 {
		return errorResponse(400, "VALIDATION_ERROR", "Maximum 100 questions per quiz", requestID), nil
	}

	// Assign IDs to questions and options
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
		HostUserID:  userId,
		Title:       req.Title,
		Description: req.Description,
		Questions:   req.Questions,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := dbClient.CreateQuiz(ctx, quiz); err != nil {
		observability.Error(ctx, "failed to create quiz", "error", err.Error())
		return errorResponse(500, "INTERNAL_ERROR", "Failed to create quiz", requestID), nil
	}

	observability.Info(ctx, "quiz created", "quizId", quiz.QuizID)
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
