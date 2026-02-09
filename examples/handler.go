// Package examples demonstrates usage of the connect-go-errors package
// in typical Connect RPC service handlers.
package examples

import (
	"context"

	cerr "github.com/balcieren/connect-go-errors"
)

// UserRepository is a mock interface for demonstration purposes.
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	Create(ctx context.Context, email, name string) (*User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
}

// User represents a user entity.
type User struct {
	ID    string
	Email string
	Name  string
}

// UserHandler handles user-related RPC requests.
type UserHandler struct {
	repo UserRepository
}

// GetUser retrieves a user by ID.
func (h *UserHandler) GetUser(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, cerr.New(cerr.ErrInvalidArgument, cerr.M{
			"reason": "id is required",
		})
	}

	user, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return nil, cerr.Wrap(cerr.ErrNotFound, err, cerr.M{
			"id": id,
		})
	}

	return user, nil
}

// CreateUser creates a new user.
func (h *UserHandler) CreateUser(ctx context.Context, email, name string) (*User, error) {
	exists, _ := h.repo.EmailExists(ctx, email)
	if exists {
		return nil, cerr.New(cerr.ErrAlreadyExists, cerr.M{
			"id": email,
		})
	}

	user, err := h.repo.Create(ctx, email, name)
	if err != nil {
		return nil, cerr.Wrap(cerr.ErrInternal, err, cerr.M{})
	}

	return user, nil
}
