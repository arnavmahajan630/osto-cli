package repository

import (
	"context"
	"database/sql"
	"time"

	"osto-auth-cli/internal/models"
)

// SQLiteSessionRepository implements SessionRepository using SQLite.
type SQLiteSessionRepository struct {
	db *sql.DB
}

// NewSQLiteSessionRepository creates a new SQLiteSessionRepository.
func NewSQLiteSessionRepository(db *sql.DB) *SQLiteSessionRepository {
	return &SQLiteSessionRepository{db: db}
}

func (r *SQLiteSessionRepository) Create(ctx context.Context, s *models.Session) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES (?, ?, ?)
	`, s.UserID, s.TokenHash, s.ExpiresAt)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	s.ID = id
	return id, nil
}

func (r *SQLiteSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	s := &models.Session{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, created_at, expires_at, last_active_at, revoked_at
		FROM sessions
		WHERE token_hash = ?
	`, tokenHash).Scan(
		&s.ID, &s.UserID, &s.TokenHash, &s.CreatedAt, &s.ExpiresAt, &s.LastActiveAt, &s.RevokedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s, nil
}

func (r *SQLiteSessionRepository) Touch(ctx context.Context, tokenHash string, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE sessions SET last_active_at = ? WHERE token_hash = ?
	`, at, tokenHash)
	return err
}

func (r *SQLiteSessionRepository) Revoke(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE sessions SET revoked_at = CURRENT_TIMESTAMP WHERE token_hash = ?
	`, tokenHash)
	return err
}

func (r *SQLiteSessionRepository) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM sessions WHERE expires_at < ?
	`, now)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
