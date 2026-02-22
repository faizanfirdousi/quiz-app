package models

import "time"

// SessionStatus represents the state of a game session.
type SessionStatus string

const (
	SessionStatusLobby    SessionStatus = "LOBBY"
	SessionStatusActive   SessionStatus = "ACTIVE"
	SessionStatusFinished SessionStatus = "FINISHED"
)

// Session represents a live game session.
type Session struct {
	SessionID            string        `json:"sessionId" dynamodbav:"sessionId"`
	PIN                  string        `json:"pin" dynamodbav:"pin"`
	QuizID               string        `json:"quizId" dynamodbav:"quizId"`
	HostUserID           string        `json:"hostUserId" dynamodbav:"hostUserId"`
	Status               SessionStatus `json:"status" dynamodbav:"status"`
	CurrentQuestionIndex int           `json:"currentQuestionIndex" dynamodbav:"currentQuestionIndex"`
	StartedAt            *time.Time    `json:"startedAt,omitempty" dynamodbav:"startedAt,omitempty"`
	EndedAt              *time.Time    `json:"endedAt,omitempty" dynamodbav:"endedAt,omitempty"`
	CreatedAt            time.Time     `json:"createdAt" dynamodbav:"createdAt"`
}
