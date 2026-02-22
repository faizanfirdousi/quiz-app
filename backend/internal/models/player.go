package models

import "time"

// PlayerRole represents whether a connection belongs to a host or player.
type PlayerRole string

const (
	PlayerRoleHost   PlayerRole = "HOST"
	PlayerRolePlayer PlayerRole = "PLAYER"
)

// Player represents a connected participant in a session.
// This maps to the DynamoDB connections table.
type Player struct {
	SessionID    string     `json:"sessionId" dynamodbav:"sessionId"`
	ConnectionID string     `json:"connectionId" dynamodbav:"connectionId"`
	UserID       string     `json:"userId" dynamodbav:"userId"`
	Nickname     string     `json:"nickname" dynamodbav:"nickname"`
	Role         PlayerRole `json:"role" dynamodbav:"role"`
	ConnectedAt  time.Time  `json:"connectedAt" dynamodbav:"connectedAt"`
	TTL          int64      `json:"ttl" dynamodbav:"ttl"` // Unix timestamp + 24h for DynamoDB TTL
}

// PlayerScore is used for leaderboard display.
type PlayerScore struct {
	UserID   string  `json:"userId"`
	Nickname string  `json:"nickname"`
	Score    float64 `json:"score"`
	Rank     int64   `json:"rank"`
}
