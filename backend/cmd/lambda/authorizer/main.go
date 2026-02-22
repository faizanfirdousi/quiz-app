package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"kahootclone/internal/auth"
	"kahootclone/internal/config"
	"kahootclone/internal/observability"
)

var (
	cfg       *config.Config
	validator *auth.CognitoValidator
)

func init() {
	cfg = config.Load()
	observability.InitLogger(cfg.LogLevel, cfg.Env)
	observability.InitTracer(cfg.Env)

	validator = auth.NewCognitoValidator(cfg.CognitoRegion, cfg.CognitoUserPoolID, cfg.CognitoClientID)
	if err := validator.Init(context.Background()); err != nil {
		panic(fmt.Sprintf("failed to initialize Cognito validator: %v", err))
	}
}

func handler(ctx context.Context, event events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	observability.Info(ctx, "Lambda authorizer invoked")

	// Extract token — may be "Bearer <token>" or just "<token>"
	tokenString := event.AuthorizationToken
	if strings.HasPrefix(strings.ToLower(tokenString), "bearer ") {
		tokenString = strings.TrimSpace(tokenString[7:])
	}

	if tokenString == "" {
		return generatePolicy("", "Deny", event.MethodArn, nil), nil
	}

	claims, err := validator.ValidateToken(ctx, tokenString)
	if err != nil {
		observability.Warn(ctx, "token validation failed", "error", err.Error())
		return generatePolicy("", "Deny", event.MethodArn, nil), nil
	}

	// Build context to pass to downstream Lambda
	authContext := map[string]interface{}{
		"userId":   claims.UserID,
		"email":    claims.Email,
		"username": claims.Username,
		"role":     claims.Role,
	}

	observability.Info(ctx, "authorization successful", "userId", claims.UserID)

	return generatePolicy(claims.UserID, "Allow", event.MethodArn, authContext), nil
}

func generatePolicy(principalID, effect, resource string, context map[string]interface{}) events.APIGatewayCustomAuthorizerResponse {
	resp := events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: principalID,
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: []string{resource},
				},
			},
		},
	}

	if context != nil {
		resp.Context = context
	}

	return resp
}

func main() {
	lambda.Start(handler)
}
