package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// Claims contains the validated JWT claims from Cognito.
type Claims struct {
	UserID   string `json:"sub"`
	Email    string `json:"email"`
	Username string `json:"cognito:username"`
	Role     string `json:"custom:role"`
}

// CognitoValidator validates Cognito JWTs using JWKS.
type CognitoValidator struct {
	region     string
	userPoolID string
	clientID   string
	jwksURL    string

	mu        sync.RWMutex
	keySet    jwk.Set
	lastFetch time.Time
}

// NewCognitoValidator creates a new validator that fetches and caches JWKS.
func NewCognitoValidator(region, userPoolID, clientID string) *CognitoValidator {
	jwksURL := fmt.Sprintf(
		"https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json",
		region, userPoolID,
	)
	return &CognitoValidator{
		region:     region,
		userPoolID: userPoolID,
		clientID:   clientID,
		jwksURL:    jwksURL,
	}
}

// Init fetches the JWKS on startup. Should be called once during initialization.
// This blocks until the fetch completes or times out.
func (v *CognitoValidator) Init(ctx context.Context) error {
	return v.refreshKeySet(ctx)
}

// InitAsync fetches the JWKS in the background so it doesn't block startup.
// Useful for local development where outbound HTTPS may be slow.
func (v *CognitoValidator) InitAsync() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := v.refreshKeySet(ctx); err != nil {
			slog.Warn("async JWKS fetch failed — auth will initialize lazily on first request", "error", err.Error())
		}
	}()
}

func (v *CognitoValidator) refreshKeySet(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	slog.Info("fetching Cognito JWKS", "url", v.jwksURL)

	keySet, err := jwk.Fetch(ctx, v.jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS from %s: %w", v.jwksURL, err)
	}

	v.mu.Lock()
	v.keySet = keySet
	v.lastFetch = time.Now()
	v.mu.Unlock()

	slog.Info("Cognito JWKS fetched successfully", "keys", keySet.Len())
	return nil
}

func (v *CognitoValidator) getKeySet(ctx context.Context) (jwk.Set, error) {
	v.mu.RLock()
	ks := v.keySet
	lastFetch := v.lastFetch
	v.mu.RUnlock()

	// Refresh every hour
	if ks == nil || time.Since(lastFetch) > time.Hour {
		if err := v.refreshKeySet(ctx); err != nil {
			if ks != nil {
				// Use stale keys if refresh fails
				slog.Warn("JWKS refresh failed, using stale keys", "error", err.Error())
				return ks, nil
			}
			return nil, err
		}
		v.mu.RLock()
		ks = v.keySet
		v.mu.RUnlock()
	}

	return ks, nil
}

// ValidateToken validates a Cognito JWT and returns the extracted claims.
func (v *CognitoValidator) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	keySet, err := v.getKeySet(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWKS: %w", err)
	}

	issuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", v.region, v.userPoolID)

	token, err := jwt.Parse(
		[]byte(tokenString),
		jwt.WithKeySet(keySet),
		jwt.WithIssuer(issuer),
		jwt.WithValidate(true),
	)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Verify audience/client ID
	aud, ok := token.Get("client_id")
	if !ok {
		// Try aud claim (id tokens use aud, access tokens use client_id)
		audList := token.Audience()
		if len(audList) == 0 {
			return nil, fmt.Errorf("token missing audience/client_id claim")
		}
		aud = audList[0]
	}
	if fmt.Sprint(aud) != v.clientID {
		return nil, fmt.Errorf("token client_id mismatch: expected %s, got %s", v.clientID, aud)
	}

	claims := &Claims{
		UserID: token.Subject(),
	}

	if email, ok := token.Get("email"); ok {
		claims.Email = fmt.Sprint(email)
	}
	if username, ok := token.Get("cognito:username"); ok {
		claims.Username = fmt.Sprint(username)
	}
	if role, ok := token.Get("custom:role"); ok {
		claims.Role = fmt.Sprint(role)
	}

	return claims, nil
}

// ParseBearerToken extracts the token from an "Authorization: Bearer <token>" header.
func ParseBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("invalid Authorization header format, expected 'Bearer <token>'")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", fmt.Errorf("empty bearer token")
	}

	return token, nil
}
