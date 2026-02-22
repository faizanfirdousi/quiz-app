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

// PutConnection registers a WebSocket connection in DynamoDB.
func (c *Client) PutConnection(ctx context.Context, player *models.Player) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	observability.Debug(ctx, "putting connection",
		"sessionId", player.SessionID,
		"connectionId", player.ConnectionID,
		"nickname", player.Nickname,
	)

	// Set TTL to 24 hours from now
	player.TTL = time.Now().Add(24 * time.Hour).Unix()

	item, err := attributevalue.MarshalMap(player)
	if err != nil {
		return err
	}

	_, err = c.DDB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.ConnectionsTable),
		Item:      item,
	})
	return err
}

// DeleteConnection removes a WebSocket connection from DynamoDB.
func (c *Client) DeleteConnection(ctx context.Context, sessionID, connectionID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	observability.Debug(ctx, "deleting connection", "sessionId", sessionID, "connectionId", connectionID)

	_, err := c.DDB.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(c.ConnectionsTable),
		Key: map[string]types.AttributeValue{
			"sessionId":    &types.AttributeValueMemberS{Value: sessionID},
			"connectionId": &types.AttributeValueMemberS{Value: connectionID},
		},
	})
	return err
}

// GetConnectionsBySession returns all connections for a given session.
func (c *Client) GetConnectionsBySession(ctx context.Context, sessionID string) ([]models.Player, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	observability.Debug(ctx, "getting connections by session", "sessionId", sessionID)

	result, err := c.DDB.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(c.ConnectionsTable),
		KeyConditionExpression: aws.String("sessionId = :sid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":sid": &types.AttributeValueMemberS{Value: sessionID},
		},
	})
	if err != nil {
		return nil, err
	}

	var players []models.Player
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &players); err != nil {
		return nil, err
	}
	return players, nil
}

// GetPlayerCountBySession returns the number of players (non-host) connected to a session.
func (c *Client) GetPlayerCountBySession(ctx context.Context, sessionID string) (int, error) {
	players, err := c.GetConnectionsBySession(ctx, sessionID)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, p := range players {
		if p.Role == models.PlayerRolePlayer {
			count++
		}
	}
	return count, nil
}

// GetSessionByConnectionID finds the session a connection belongs to using the GSI.
func (c *Client) GetSessionByConnectionID(ctx context.Context, connectionID string) (*models.Player, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	observability.Debug(ctx, "getting session by connectionId", "connectionId", connectionID)

	result, err := c.DDB.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(c.ConnectionsTable),
		IndexName:              aws.String("connectionId-index"),
		KeyConditionExpression: aws.String("connectionId = :cid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":cid": &types.AttributeValueMemberS{Value: connectionID},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("connection %s not found", connectionID)
	}

	var player models.Player
	if err := attributevalue.UnmarshalMap(result.Items[0], &player); err != nil {
		return nil, err
	}
	return &player, nil
}

// GetConnectionByUserID finds a specific player's connection in a session.
func (c *Client) GetConnectionByUserID(ctx context.Context, sessionID, userID string) (*models.Player, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	players, err := c.GetConnectionsBySession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	for _, p := range players {
		if p.UserID == userID {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("user %s not found in session %s", userID, sessionID)
}
