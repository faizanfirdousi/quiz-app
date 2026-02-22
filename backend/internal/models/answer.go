package models

import "time"

// Answer represents a player's answer to a question.
type Answer struct {
	SessionID        string    `json:"sessionId" dynamodbav:"sessionId"`
	UserIDQuestionID string    `json:"userIdQuestionId" dynamodbav:"userIdQuestionId"` // SK: "userId#questionId"
	QuestionID       string    `json:"questionId" dynamodbav:"questionId"`
	UserID           string    `json:"userId" dynamodbav:"userId"`
	SelectedOptionID string    `json:"selectedOptionId" dynamodbav:"selectedOptionId"`
	IsCorrect        bool      `json:"isCorrect" dynamodbav:"isCorrect"`
	TimeTakenMs      int64     `json:"timeTakenMs" dynamodbav:"timeTakenMs"`
	PointsEarned     int       `json:"pointsEarned" dynamodbav:"pointsEarned"`
	AnsweredAt       time.Time `json:"answeredAt" dynamodbav:"answeredAt"`
}
