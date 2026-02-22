package main

import (
	"context"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

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
	gameEngine  *game.Engine
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

	broadcaster := game.NewBroadcaster(dbClient, cfg.Env)
	gameEngine = game.NewEngine(dbClient, redisClient, broadcaster)
}

func handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	connectionID := event.RequestContext.ConnectionID

	observability.Info(ctx, "WebSocket $default", "connectionId", connectionID)

	if err := gameEngine.HandleMessage(ctx, connectionID, []byte(event.Body)); err != nil {
		observability.Error(ctx, "failed to handle WS message", "connectionId", connectionID, "error", err.Error())

		// Send error back to client
		// In production, use API Gateway Management API to post back
		errPayload := models.WSOutbound{
			Type: models.WSTypeError,
			Payload: models.ErrorPayload{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		}
		_ = gameEngine.Broadcaster.SendToConnection(ctx, connectionID, errPayload)
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(handler)
}
