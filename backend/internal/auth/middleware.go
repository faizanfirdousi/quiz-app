package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"kahootclone/internal/observability"
)

type contextKey string

const userClaimsKey contextKey = "userClaims"

// GetClaims extracts the authenticated user's claims from the request context.
func GetClaims(ctx context.Context) *Claims {
	claims, _ := ctx.Value(userClaimsKey).(*Claims)
	return claims
}

// Middleware returns an HTTP middleware that validates Cognito JWTs.
// On success, it injects Claims into the request context.
func Middleware(validator *CognitoValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate request ID
			requestID := uuid.New().String()
			ctx := observability.WithRequestID(r.Context(), requestID)

			tokenString, err := ParseBearerToken(r)
			if err != nil {
				writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error(), requestID)
				return
			}

			claims, err := validator.ValidateToken(ctx, tokenString)
			if err != nil {
				observability.Warn(ctx, "auth validation failed", "error", err.Error())
				writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token", requestID)
				return
			}

			// Inject claims and userId into context
			ctx = context.WithValue(ctx, userClaimsKey, claims)
			ctx = observability.WithUserID(ctx, claims.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, code, message, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)

	resp := map[string]interface{}{
		"success":   false,
		"error":     map[string]string{"code": code, "message": message},
		"requestId": requestID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(resp)
}
