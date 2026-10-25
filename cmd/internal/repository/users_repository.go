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

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT * FROM users WHERE email = ?`

	user := &models.User{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil

}

func (r *UserRepository) FindAll(
	ctx context.Context,
) ([]models.User, error) {
	query := `SELECT * FROM users`

	rows, err := r.db.QueryContext(
		ctx,
		query,
	)
	users := make([]models.User, 0)

	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var user models.User

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
