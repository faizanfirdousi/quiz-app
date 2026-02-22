package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// DynamoDB
	DynamoDBEndpoint string // empty in prod (uses default), set for local: "http://localhost:8000"
	DynamoDBRegion   string // e.g. "ap-south-1"
	QuizzesTable     string // "kahootclone-quizzes"
	SessionsTable    string // "kahootclone-sessions"
	ConnectionsTable string // "kahootclone-connections"
	AnswersTable     string // "kahootclone-answers"

	// Redis / ElastiCache
	RedisAddr     string // "localhost:6379" or ElastiCache endpoint
	RedisPassword string // empty for local
	RedisDB       int    // 0

	// Cognito
	CognitoRegion     string // "ap-south-1"
	CognitoUserPoolID string // "ap-south-1_XXXXXXX"
	CognitoClientID   string // app client ID

	// WebSocket (for local dev server and for broadcast Lambda)
	WSEndpoint string // local: "ws://localhost:8080/ws", prod: API Gateway management endpoint

	// App
	Env      string // "local" or "production"
	Port     string // "8080" for local dev server
	LogLevel string // "debug", "info", "warn", "error"
}

// Load reads environment variables and returns a validated Config.
// It attempts to load a .env file first (for local development).
// Panics if required variables are missing.
func Load() *Config {
	// Best-effort .env load — ignore error in prod where file may not exist
	_ = godotenv.Load()

	cfg := &Config{
		DynamoDBEndpoint: os.Getenv("DYNAMODB_ENDPOINT"),
		DynamoDBRegion:   requireEnv("DYNAMODB_REGION"),
		QuizzesTable:     requireEnv("QUIZZES_TABLE"),
		SessionsTable:    requireEnv("SESSIONS_TABLE"),
		ConnectionsTable: requireEnv("CONNECTIONS_TABLE"),
		AnswersTable:     requireEnv("ANSWERS_TABLE"),

		RedisAddr:     requireEnv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		CognitoRegion:     requireEnv("COGNITO_REGION"),
		CognitoUserPoolID: requireEnv("COGNITO_USER_POOL_ID"),
		CognitoClientID:   requireEnv("COGNITO_CLIENT_ID"),

		WSEndpoint: requireEnv("WS_ENDPOINT"),

		Env:      getEnvDefault("ENV", "local"),
		Port:     getEnvDefault("PORT", "8080"),
		LogLevel: getEnvDefault("LOG_LEVEL", "info"),
	}

	return cfg
}

// IsLocal returns true if the application is running in local development mode.
func (c *Config) IsLocal() bool {
	return c.Env == "local"
}

// IsProduction returns true if the application is running in production mode.
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return val
}

func getEnvDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		panic(fmt.Sprintf("environment variable %s must be an integer, got %q", key, val))
	}
	return i
}
