package repository

import (
	"context"
	"database/sql"
	"github.com/taro0079/go-htmx/cmd/internal/models"
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(
	ctx context.Context,
	post *models.Post,
) (*models.Post, error) {
	query := `
		INSERT INTO posts (title, content, published)
		VALUES (?, ?, ?)
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.Published,
	)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return nil, err
	}

	post.ID = id

	return post, nil
}

func (r *PostRepository) Search(ctx context.Context, q string) ([]models.Post, error) {
	query := `SELECT * FROM posts WHERE title LIKE ? OR content LIKE ?`
	pattern := "%" + q + "%"
	rows, err := r.db.QueryContext(
		ctx,
		query,
		pattern,
		pattern,
	)

	if err != nil {
		return nil, err
	}

	posts := make([]models.Post, 0)

	for rows.Next() {
		var post models.Post

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.Published,
			&post.CreatedAt,
			&post.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *PostRepository) FindById(
	ctx context.Context,
	postId int64,
) (*models.Post, error) {
	query := `SELECT * FROM posts WHERE id = ?`
	post := &models.Post{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		postId,
	).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.Published,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return post, nil
}

func (r *PostRepository) FindAll(
	ctx context.Context,
) ([]models.Post, error) {
	query := `SELECT * FROM posts`

	rows, err := r.db.QueryContext(
		ctx,
		query,
	)
	posts := make([]models.Post, 0)

	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var post models.Post

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.Published,
			&post.CreatedAt,
			&post.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
