package db

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"kahootclone/internal/models"
	"kahootclone/internal/observability"
)

// CreateSession stores a new session in DynamoDB.
func (c *Client) CreateSession(ctx context.Context, session *models.Session) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	observability.Debug(ctx, "creating session", "sessionId", session.SessionID, "pin", session.PIN)

	item, err := attributevalue.MarshalMap(session)
	if err != nil {
		return err
	}

	_, err = c.DDB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.SessionsTable),
		Item:      item,
	})
	return err
}

// GetSession retrieves a session by its ID using consistent read.
func (c *Client) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	observability.Debug(ctx, "getting session", "sessionId", sessionID)

	result, err := c.DDB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(c.SessionsTable),
		ConsistentRead: aws.Bool(true),
		Key: map[string]types.AttributeValue{
			"sessionId": &types.AttributeValueMemberS{Value: sessionID},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}

	var session models.Session
	if err := attributevalue.UnmarshalMap(result.Item, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// GetSessionByPIN looks up a session by its 6-digit PIN using a GSI.
func (c *Client) GetSessionByPIN(ctx context.Context, pin string) (*models.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	observability.Debug(ctx, "getting session by PIN", "pin", pin)

	result, err := c.DDB.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(c.SessionsTable),
		IndexName:              aws.String("pin-index"),
		KeyConditionExpression: aws.String("pin = :pin"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pin": &types.AttributeValueMemberS{Value: pin},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, nil
	}

	var session models.Session
	if err := attributevalue.UnmarshalMap(result.Items[0], &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateSessionStatus atomically updates the status and related fields of a session.
func (c *Client) UpdateSessionStatus(ctx context.Context, sessionID string, status models.SessionStatus, questionIndex int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	observability.Debug(ctx, "updating session status", "sessionId", sessionID, "status", status)

	updateExpr := "SET #status = :status, currentQuestionIndex = :idx"
	exprAttrNames := map[string]string{
		"#status": "status",
	}
	exprAttrValues := map[string]types.AttributeValue{
		":status": &types.AttributeValueMemberS{Value: string(status)},
		":idx":    &types.AttributeValueMemberN{Value: intToString(questionIndex)},
	}

	if status == models.SessionStatusActive {
		now := time.Now().UTC().Format(time.RFC3339)
		updateExpr += ", startedAt = :startedAt"
		exprAttrValues[":startedAt"] = &types.AttributeValueMemberS{Value: now}
	} else if status == models.SessionStatusFinished {
		now := time.Now().UTC().Format(time.RFC3339)
		updateExpr += ", endedAt = :endedAt"
		exprAttrValues[":endedAt"] = &types.AttributeValueMemberS{Value: now}
	}

	_, err := c.DDB.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(c.SessionsTable),
		Key: map[string]types.AttributeValue{
			"sessionId": &types.AttributeValueMemberS{Value: sessionID},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeNames:  exprAttrNames,
		ExpressionAttributeValues: exprAttrValues,
	})
	return err
}

func intToString(i int) string {
	return fmt.Sprintf("%d", i)
}
