package models

import "encoding/json"

// --- Inbound (client → server) ---

// WSInbound is the envelope for all incoming WebSocket messages.
type WSInbound struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

// JoinSessionPayload is sent when a player wants to join a session.
type JoinSessionPayload struct {
	SessionID string `json:"sessionId"`
	Nickname  string `json:"nickname"`
}

// SubmitAnswerPayload is sent when a player answers a question.
type SubmitAnswerPayload struct {
	QuestionID       string `json:"questionId"`
	SelectedOptionID string `json:"selectedOptionId"`
	TimeTakenMs      int64  `json:"timeTakenMs"`
}

// StartGamePayload is sent by the host to start the game.
type StartGamePayload struct {
	SessionID string `json:"sessionId"`
}

// NextQuestionPayload is sent by the host to advance to the next question.
type NextQuestionPayload struct {
	SessionID string `json:"sessionId"`
}

// EndGamePayload is sent by the host to end the game.
type EndGamePayload struct {
	SessionID string `json:"sessionId"`
}

// --- Outbound (server → client) ---

// WSOutbound is the envelope for all outgoing WebSocket messages.
type WSOutbound struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// PlayerJoinedPayload is broadcast when a new player joins the session.
type PlayerJoinedPayload struct {
	Nickname    string `json:"nickname"`
	PlayerCount int    `json:"playerCount"`
}

// GameStartedPayload is broadcast when the host starts the game.
type GameStartedPayload struct {
	TotalQuestions int `json:"totalQuestions"`
}

// QuestionPayload is broadcast when a new question begins.
type QuestionPayload struct {
	QuestionIndex  int      `json:"questionIndex"`
	TotalQuestions int      `json:"totalQuestions"`
	Text           string   `json:"text"`
	Options        []Option `json:"options"` // NOTE: never send correctOptionId to players
	TimeLimitMs    int      `json:"timeLimitMs"`
	Points         int      `json:"points"`
}

// AnswerResultPayload is sent only to the player who answered.
type AnswerResultPayload struct {
	IsCorrect     bool   `json:"isCorrect"`
	PointsEarned  int    `json:"pointsEarned"`
	TotalScore    int    `json:"totalScore"`
	Rank          int64  `json:"rank"`
	CorrectOption string `json:"correctOptionId"`
}

// QuestionEndedPayload is broadcast to all after the timer expires.
type QuestionEndedPayload struct {
	CorrectOption string        `json:"correctOptionId"`
	Leaderboard   []PlayerScore `json:"leaderboard"` // top 10
}

// LeaderboardUpdatePayload is broadcast between questions.
type LeaderboardUpdatePayload struct {
	Leaderboard []PlayerScore `json:"leaderboard"`
}

// GameOverPayload is broadcast when the game ends.
type GameOverPayload struct {
	FinalLeaderboard []PlayerScore `json:"finalLeaderboard"`
}

// ErrorPayload is sent to a client when an error occurs.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WebSocket event type constants for outbound messages.
const (
	WSTypePlayerJoined      = "player_joined"
	WSTypeGameStarted       = "game_started"
	WSTypeQuestion          = "question"
	WSTypeAnswerResult      = "answer_result"
	WSTypeQuestionEnded     = "question_ended"
	WSTypeLeaderboardUpdate = "leaderboard_update"
	WSTypeGameOver          = "game_over"
	WSTypeError             = "error"
)

// WebSocket action constants for inbound messages.
const (
	WSActionJoinSession  = "join_session"
	WSActionSubmitAnswer = "submit_answer"
	WSActionStartGame    = "start_game"
	WSActionNextQuestion = "next_question"
	WSActionEndGame      = "end_game"
)
