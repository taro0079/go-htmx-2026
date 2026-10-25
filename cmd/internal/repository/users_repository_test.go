package repository

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/taro0079/go-htmx/cmd/internal/models"
	"testing"
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
			password_hash TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
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

func TestFindAll(t *testing.T) {
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
			password_hash TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)

	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}

	user2 := &models.User{
		Email:        "test2@exampl.com",
		PasswordHash: "hashed_password",
	}

	repo := NewUserRepository(db)
	_, err = repo.Create(context.Background(), user)

	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.Create(context.Background(), user2)

	if err != nil {
		t.Fatal(err)
	}

	users, err := repo.FindAll(context.Background())

	if err != nil {
		t.Fatal(err)
	}

	if len(users) != 2 {
		t.Errorf("The number of created user is not 2. got %d", len(users))
	}

}
