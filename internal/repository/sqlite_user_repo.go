package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"osto-auth-cli/internal/models"
)

// SQLiteUserRepository implements UserRepository using SQLite.
type SQLiteUserRepository struct {
	db *sql.DB
}

// NewSQLiteUserRepository creates a new SQLiteUserRepository.
func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

func (r *SQLiteUserRepository) Create(ctx context.Context, user *models.User) error {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO users (username, password_hash, name, birth_date)
		VALUES (?, ?, ?, ?)
	`, user.Username, user.PasswordHash, user.Name, user.BirthDate)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.username") {
			return ErrDuplicateUsername
		}
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = id
	return nil
}

func (r *SQLiteUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	return nil, errors.New("not implemented")
}

func (r *SQLiteUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash, name, birth_date, created_at, last_login_at, mfa_enabled, mfa_secret_enc, failed_attempts, locked_until
		FROM users
		WHERE username = ?
	`, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Name, &user.BirthDate,
		&user.CreatedAt, &user.LastLoginAt, &user.MFAEnabled, &user.MFASecretEnc,
		&user.FailedAttempts, &user.LockedUntil,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}


func (r *SQLiteUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", username).Scan(&exists)
	return exists, err
}

func (r *SQLiteUserRepository) RecordLoginSuccess(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (r *SQLiteUserRepository) RecordLoginFailure(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (r *SQLiteUserRepository) SetMFA(ctx context.Context, id int64, enabled bool, secret string) error {
	return errors.New("not implemented")
}
