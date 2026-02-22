package db

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"kahootclone/internal/models"
	"kahootclone/internal/observability"
)

// CreateQuiz stores a new quiz in DynamoDB.
func (c *Client) CreateQuiz(ctx context.Context, quiz *models.Quiz) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	observability.Debug(ctx, "creating quiz", "quizId", quiz.QuizID)

	item, err := attributevalue.MarshalMap(quiz)
	if err != nil {
		return err
	}

	_, err = c.DDB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.QuizzesTable),
		Item:      item,
	})
	return err
}

// GetQuiz retrieves a quiz by its ID using consistent read.
func (c *Client) GetQuiz(ctx context.Context, quizID string) (*models.Quiz, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	observability.Debug(ctx, "getting quiz", "quizId", quizID)

	result, err := c.DDB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(c.QuizzesTable),
		ConsistentRead: aws.Bool(true),
		Key: map[string]types.AttributeValue{
			"quizId": &types.AttributeValueMemberS{Value: quizID},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}

	var quiz models.Quiz
	if err := attributevalue.UnmarshalMap(result.Item, &quiz); err != nil {
		return nil, err
	}
	return &quiz, nil
}

// ListQuizzesByHost retrieves all quizzes for a given host.
// Uses a scan with filter — acceptable for low-volume use; consider GSI for production scale.
func (c *Client) ListQuizzesByHost(ctx context.Context, hostUserID string) ([]models.Quiz, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	observability.Debug(ctx, "listing quizzes by host", "hostUserId", hostUserID)

	result, err := c.DDB.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(c.QuizzesTable),
		FilterExpression: aws.String("hostUserId = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: hostUserID},
		},
	})
	if err != nil {
		return nil, err
	}

	var quizzes []models.Quiz
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &quizzes); err != nil {
		return nil, err
	}
	return quizzes, nil
}
