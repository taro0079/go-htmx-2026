package repository

import (
	"context"
	"database/sql"
	"time"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(
	ctx context.Context,
	tokenHash string,
	userID int64,
	expiresAt time.Time,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO sessions (token_hash, user_id, expires_at)
		VALUES (?, ?, ?)`,
		tokenHash,
		userID,
		expiresAt.Unix(),
	)
	return err
}

func (r *SessionRepository) FindUserID(
	ctx context.Context,
	tokenHash string,
) (int64, error) {
	var userID int64

	err := r.db.QueryRowContext(
		ctx,
		`SELECT user_id
		FROM sessions
		WHERE token_hash = ? AND expires_at > ?`,
		tokenHash,
		time.Now().Unix(),
	).Scan(&userID)

	return userID, err
}

func (r *SessionRepository) Delete(
	ctx context.Context,
	tokenHash string,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM sessions WHERE token_hash = ?`,
		tokenHash,
	)
	return err
}
