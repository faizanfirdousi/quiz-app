package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func main() {
	endpoint := "http://localhost:8000"
	region := "ap-south-1"

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "")),
	)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tables := []struct {
		name  string
		input *dynamodb.CreateTableInput
	}{
		{
			name: "kahootclone-quizzes",
			input: &dynamodb.CreateTableInput{
				TableName:   aws.String("kahootclone-quizzes"),
				BillingMode: types.BillingModePayPerRequest,
				AttributeDefinitions: []types.AttributeDefinition{
					{AttributeName: aws.String("quizId"), AttributeType: types.ScalarAttributeTypeS},
				},
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("quizId"), KeyType: types.KeyTypeHash},
				},
			},
		},
		{
			name: "kahootclone-sessions",
			input: &dynamodb.CreateTableInput{
				TableName:   aws.String("kahootclone-sessions"),
				BillingMode: types.BillingModePayPerRequest,
				AttributeDefinitions: []types.AttributeDefinition{
					{AttributeName: aws.String("sessionId"), AttributeType: types.ScalarAttributeTypeS},
					{AttributeName: aws.String("pin"), AttributeType: types.ScalarAttributeTypeS},
				},
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("sessionId"), KeyType: types.KeyTypeHash},
				},
				GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
					{
						IndexName: aws.String("pin-index"),
						KeySchema: []types.KeySchemaElement{
							{AttributeName: aws.String("pin"), KeyType: types.KeyTypeHash},
						},
						Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
					},
				},
			},
		},
		{
			name: "kahootclone-connections",
			input: &dynamodb.CreateTableInput{
				TableName:   aws.String("kahootclone-connections"),
				BillingMode: types.BillingModePayPerRequest,
				AttributeDefinitions: []types.AttributeDefinition{
					{AttributeName: aws.String("sessionId"), AttributeType: types.ScalarAttributeTypeS},
					{AttributeName: aws.String("connectionId"), AttributeType: types.ScalarAttributeTypeS},
				},
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("sessionId"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("connectionId"), KeyType: types.KeyTypeRange},
				},
				GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
					{
						IndexName: aws.String("connectionId-index"),
						KeySchema: []types.KeySchemaElement{
							{AttributeName: aws.String("connectionId"), KeyType: types.KeyTypeHash},
						},
						Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
					},
				},
			},
		},
		{
			name: "kahootclone-answers",
			input: &dynamodb.CreateTableInput{
				TableName:   aws.String("kahootclone-answers"),
				BillingMode: types.BillingModePayPerRequest,
				AttributeDefinitions: []types.AttributeDefinition{
					{AttributeName: aws.String("sessionId"), AttributeType: types.ScalarAttributeTypeS},
					{AttributeName: aws.String("userIdQuestionId"), AttributeType: types.ScalarAttributeTypeS},
				},
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("sessionId"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("userIdQuestionId"), KeyType: types.KeyTypeRange},
				},
			},
		},
	}

	for _, t := range tables {
		fmt.Printf("Creating table: %s ...\n", t.name)
		_, err := client.CreateTable(ctx, t.input)
		if err != nil {
			fmt.Printf("  ⚠ %s: %v\n", t.name, err)
		} else {
			fmt.Printf("  ✅ %s created\n", t.name)
		}
	}

	// List tables to verify
	fmt.Println("\nVerifying tables...")
	result, err := client.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		fmt.Printf("Failed to list tables: %v\n", err)
		os.Exit(1)
	}
	for _, name := range result.TableNames {
		fmt.Printf("  📋 %s\n", name)
	}
	fmt.Println("\nDone!")
}
