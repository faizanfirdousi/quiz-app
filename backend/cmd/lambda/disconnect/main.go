package main

import (
	"context"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

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

func handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	connectionID := event.RequestContext.ConnectionID

	observability.Info(ctx, "WebSocket $disconnect", "connectionId", connectionID)

	// Look up the connection to find its session
	player, err := dbClient.GetSessionByConnectionID(ctx, connectionID)
	if err != nil {
		observability.Warn(ctx, "connection not found during disconnect", "connectionId", connectionID, "error", err.Error())
		return events.APIGatewayProxyResponse{StatusCode: 200}, nil
	}

	// Delete the connection
	if err := dbClient.DeleteConnection(ctx, player.SessionID, connectionID); err != nil {
		observability.Error(ctx, "failed to delete connection", "connectionId", connectionID, "error", err.Error())
	}

	observability.Info(ctx, "WebSocket disconnected",
		"connectionId", connectionID,
		"sessionId", player.SessionID,
		"userId", player.UserID,
	)

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(handler)
}
