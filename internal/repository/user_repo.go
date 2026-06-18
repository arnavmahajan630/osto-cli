package repository

import (
	"context"

	"osto-auth-cli/internal/models"
)

// UserRepository defines the persistence contract for users.
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	RecordLoginSuccess(ctx context.Context, id int64) error
	RecordLoginFailure(ctx context.Context, id int64) error
	SetMFA(ctx context.Context, id int64, enabled bool, secret string) error
}
