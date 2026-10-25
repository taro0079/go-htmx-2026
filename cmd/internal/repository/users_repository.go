package repository

import (
	"context"
	"database/sql"
	"github.com/taro0079/go-htmx/cmd/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(
	ctx context.Context,
	user *models.User,
) (*models.User, error) {
	query := `
		INSERT INTO users (email, password_hash)
		VALUES (?, ?)
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.Email,
		user.PasswordHash,
	)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return nil, err
	}

	user.ID = id

	return user, nil
}
