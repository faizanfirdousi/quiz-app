package observability

import (
	"context"
	"log/slog"
	"os"
)

type contextKey string

const requestIDKey contextKey = "requestId"
const userIDKey contextKey = "userId"
const sessionIDKey contextKey = "sessionId"

// InitLogger sets up the global slog logger.
// Uses JSON handler for production, text handler for local development.
func InitLogger(level string, env string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	var handler slog.Handler
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler).With("service", "kahootclone")
	slog.SetDefault(logger)

	return logger
}

// WithRequestID adds a request ID to the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// WithUserID adds a user ID to the context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// WithSessionID adds a session ID to the context.
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

// LogAttrs extracts request/user/session IDs from context and returns slog attributes.
func LogAttrs(ctx context.Context) []slog.Attr {
	attrs := []slog.Attr{}

	if rid, ok := ctx.Value(requestIDKey).(string); ok && rid != "" {
		attrs = append(attrs, slog.String("requestId", rid))
	}
	if uid, ok := ctx.Value(userIDKey).(string); ok && uid != "" {
		attrs = append(attrs, slog.String("userId", uid))
	}
	if sid, ok := ctx.Value(sessionIDKey).(string); ok && sid != "" {
		attrs = append(attrs, slog.String("sessionId", sid))
	}

	return attrs
}

// Info logs an info-level message with context-derived attributes.
func Info(ctx context.Context, msg string, args ...any) {
	attrs := LogAttrs(ctx)
	allArgs := make([]any, 0, len(attrs)*2+len(args))
	for _, a := range attrs {
		allArgs = append(allArgs, a.Key, a.Value.Any())
	}
	allArgs = append(allArgs, args...)
	slog.InfoContext(ctx, msg, allArgs...)
}

// Error logs an error-level message with context-derived attributes.
func Error(ctx context.Context, msg string, args ...any) {
	attrs := LogAttrs(ctx)
	allArgs := make([]any, 0, len(attrs)*2+len(args))
	for _, a := range attrs {
		allArgs = append(allArgs, a.Key, a.Value.Any())
	}
	allArgs = append(allArgs, args...)
	slog.ErrorContext(ctx, msg, allArgs...)
}

// Debug logs a debug-level message with context-derived attributes.
func Debug(ctx context.Context, msg string, args ...any) {
	attrs := LogAttrs(ctx)
	allArgs := make([]any, 0, len(attrs)*2+len(args))
	for _, a := range attrs {
		allArgs = append(allArgs, a.Key, a.Value.Any())
	}
	allArgs = append(allArgs, args...)
	slog.DebugContext(ctx, msg, allArgs...)
}

// Warn logs a warn-level message with context-derived attributes.
func Warn(ctx context.Context, msg string, args ...any) {
	attrs := LogAttrs(ctx)
	allArgs := make([]any, 0, len(attrs)*2+len(args))
	for _, a := range attrs {
		allArgs = append(allArgs, a.Key, a.Value.Any())
	}
	allArgs = append(allArgs, args...)
	slog.WarnContext(ctx, msg, allArgs...)
}
