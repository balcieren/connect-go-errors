package examples

import (
	"context"

	"connectrpc.com/connect"

	connectgoerrors "github.com/balcieren/connect-go-errors"
)

func init() {
	connectgoerrors.Register(connectgoerrors.Error{
		Code:        "ERROR_INVALID_CREDENTIALS",
		MessageTpl:  "Invalid credentials for user '{{email}}'",
		ConnectCode: connect.CodeUnauthenticated,
		Retryable:   false,
	})
	connectgoerrors.Register(connectgoerrors.Error{
		Code:        "ERROR_ACCOUNT_LOCKED",
		MessageTpl:  "Account '{{email}}' is locked. Try again after {{unlock_at}}",
		ConnectCode: connect.CodePermissionDenied,
		Retryable:   false,
	})
	connectgoerrors.Register(connectgoerrors.Error{
		Code:        "ERROR_TOKEN_EXPIRED",
		MessageTpl:  "Token expired at {{expired_at}}",
		ConnectCode: connect.CodeUnauthenticated,
		Retryable:   true,
	})
}

// AuthService handles authentication RPCs.
type AuthService struct{}

// Login authenticates a user with email and password.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	if email == "" || password == "" {
		return "", connectgoerrors.New(connectgoerrors.InvalidArgument, connectgoerrors.M{
			"reason": "email and password are required",
		})
	}

	// Simulate authentication check
	if password != "correct" {
		return "", connectgoerrors.New("ERROR_INVALID_CREDENTIALS", connectgoerrors.M{
			"email": email,
		})
	}

	return "token-abc-123", nil
}

// RefreshToken refreshes an authentication token.
func (s *AuthService) RefreshToken(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", connectgoerrors.New(connectgoerrors.InvalidArgument, connectgoerrors.M{
			"reason": "token is required",
		})
	}

	if token == "expired" {
		return "", connectgoerrors.New("ERROR_TOKEN_EXPIRED", connectgoerrors.M{
			"expired_at": "2026-01-01T00:00:00Z",
		})
	}

	return "new-token-xyz", nil
}
