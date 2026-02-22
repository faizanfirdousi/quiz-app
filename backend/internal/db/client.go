package db

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"kahootclone/internal/config"
)

// Client wraps the DynamoDB client and table names.
type Client struct {
	DDB              *dynamodb.Client
	QuizzesTable     string
	SessionsTable    string
	ConnectionsTable string
	AnswersTable     string
}

// NewClient creates a new DynamoDB client from the application config.
// If DynamoDBEndpoint is set (for local development), it overrides the default endpoint.
func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	optFns := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.DynamoDBRegion),
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, err
	}

	var ddbClient *dynamodb.Client
	if cfg.DynamoDBEndpoint != "" {
		// Local DynamoDB
		ddbClient = dynamodb.NewFromConfig(awsCfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(cfg.DynamoDBEndpoint)
		})
		slog.Info("DynamoDB client initialized with custom endpoint", "endpoint", cfg.DynamoDBEndpoint)
	} else {
		ddbClient = dynamodb.NewFromConfig(awsCfg)
		slog.Info("DynamoDB client initialized with default endpoint")
	}

	return &Client{
		DDB:              ddbClient,
		QuizzesTable:     cfg.QuizzesTable,
		SessionsTable:    cfg.SessionsTable,
		ConnectionsTable: cfg.ConnectionsTable,
		AnswersTable:     cfg.AnswersTable,
	}, nil
}
