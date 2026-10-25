package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/taro0079/go-htmx/cmd/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

func TestCreate(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")

	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL
		)
	`)

	if err != nil {
		t.Fatal(err)
	}

	repo := NewUserRepository(db)

	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}

	createdUser, err := repo.Create(context.Background(), user)

	if err != nil {
		t.Fatal(err)
	}


	if createdUser == nil {
		t.Fatal("Create() returned nil")
	}

	if createdUser.ID <= 0 {
		t.Errorf("expected ID > 0, got %d", createdUser.ID)
	}

}
