package models

import "time"

// Quiz represents a quiz created by a host.
type Quiz struct {
	QuizID      string     `json:"quizId" dynamodbav:"quizId"`
	HostUserID  string     `json:"hostUserId" dynamodbav:"hostUserId"`
	Title       string     `json:"title" dynamodbav:"title"`
	Description string     `json:"description" dynamodbav:"description"`
	Questions   []Question `json:"questions" dynamodbav:"questions"`
	CreatedAt   time.Time  `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt" dynamodbav:"updatedAt"`
}

// Question represents a single question within a quiz.
type Question struct {
	QuestionID       string   `json:"questionId" dynamodbav:"questionId"`
	Text             string   `json:"text" dynamodbav:"text"`
	Options          []Option `json:"options" dynamodbav:"options"`
	CorrectOptionID  string   `json:"correctOptionId" dynamodbav:"correctOptionId"`
	TimeLimitSeconds int      `json:"timeLimitSeconds" dynamodbav:"timeLimitSeconds"`
	Points           int      `json:"points" dynamodbav:"points"`
}

// Option represents an answer option for a question.
type Option struct {
	ID   string `json:"id" dynamodbav:"id"`
	Text string `json:"text" dynamodbav:"text"`
}

// QuestionPayloadForPlayer is the sanitized question sent to players (no correct answer).
type QuestionPayloadForPlayer struct {
	QuestionID     string   `json:"questionId"`
	QuestionIndex  int      `json:"questionIndex"`
	TotalQuestions int      `json:"totalQuestions"`
	Text           string   `json:"text"`
	Options        []Option `json:"options"`
	TimeLimitMs    int      `json:"timeLimitMs"`
	Points         int      `json:"points"`
}
