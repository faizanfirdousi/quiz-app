package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

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

func handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	connectionID := event.RequestContext.ConnectionID

	observability.Info(ctx, "WebSocket $connect", "connectionId", connectionID)

	// Extract userId from authorizer context (set by Lambda authorizer)
	userId, _ := event.RequestContext.Authorizer.(map[string]interface{})["userId"].(string)

	// Extract sessionId from query string
	sessionID := event.QueryStringParameters["sessionId"]
	role := event.QueryStringParameters["role"]

	if sessionID == "" {
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	playerRole := models.PlayerRolePlayer
	if role == "HOST" {
		playerRole = models.PlayerRoleHost
	}

	// Register connection in DynamoDB
	player := &models.Player{
		SessionID:    sessionID,
		ConnectionID: connectionID,
		UserID:       userId,
		Role:         playerRole,
		ConnectedAt:  time.Now().UTC(),
	}
	if err := dbClient.PutConnection(ctx, player); err != nil {
		observability.Error(ctx, "failed to register connection", "error", err.Error())
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	observability.Info(ctx, "WebSocket connected",
		"connectionId", connectionID,
		"sessionId", sessionID,
		"userId", userId,
		"role", string(playerRole),
	)

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(handler)
}
