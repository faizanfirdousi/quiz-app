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

// PutAnswer stores a player's answer to a question.
func (c *Client) PutAnswer(ctx context.Context, answer *models.Answer) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	observability.Debug(ctx, "putting answer",
		"sessionId", answer.SessionID,
		"userId", answer.UserID,
		"questionId", answer.QuestionID,
	)

	item, err := attributevalue.MarshalMap(answer)
	if err != nil {
		return err
	}

	_, err = c.DDB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.AnswersTable),
		Item:      item,
	})
	return err
}

// GetAnswersBySession retrieves all answers for a given session.
func (c *Client) GetAnswersBySession(ctx context.Context, sessionID string) ([]models.Answer, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	observability.Debug(ctx, "getting answers by session", "sessionId", sessionID)

	result, err := c.DDB.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(c.AnswersTable),
		KeyConditionExpression: aws.String("sessionId = :sid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":sid": &types.AttributeValueMemberS{Value: sessionID},
		},
	})
	if err != nil {
		return nil, err
	}

	var answers []models.Answer
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &answers); err != nil {
		return nil, err
	}
	return answers, nil
}

// GetAnswer retrieves a specific player's answer to a specific question.
func (c *Client) GetAnswer(ctx context.Context, sessionID, userID, questionID string) (*models.Answer, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	compositeKey := userID + "#" + questionID

	result, err := c.DDB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(c.AnswersTable),
		ConsistentRead: aws.Bool(true),
		Key: map[string]types.AttributeValue{
			"sessionId":        &types.AttributeValueMemberS{Value: sessionID},
			"userIdQuestionId": &types.AttributeValueMemberS{Value: compositeKey},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}

	var answer models.Answer
	if err := attributevalue.UnmarshalMap(result.Item, &answer); err != nil {
		return nil, err
	}
	return &answer, nil
}
