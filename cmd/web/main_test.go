package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/taro0079/go-htmx/cmd/internal/repository"
)

func setupTestApp(t *testing.T) *application {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	db.SetMaxOpenConns(1)

	t.Cleanup(func() {
		db.Close()
	})

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE sessions (
			token_hash TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			expires_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`)

	if err != nil {
		t.Fatal(err)
	}

	return &application{
		logger: slog.New(
			slog.NewTextHandler(io.Discard, nil),
		),
		users:   repository.NewUserRepository(db),
		session: repository.NewSessionRepository(db),
	}
}

func TestRequireAuth(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		createSession  bool
		expired        bool
		wantStatusCode int
		wantHandlerRun bool
	}{
		{
			name:           "no Cookie",
			wantStatusCode: http.StatusSeeOther,
		},
		{
			name:           "invalid Cookie",
			token:          "invalid-token",
			wantStatusCode: http.StatusSeeOther,
		},
		{
			name:           "valid session",
			token:          "valid-token",
			createSession:  true,
			wantStatusCode: http.StatusOK,
			wantHandlerRun: true,
		},
		{
			name:           "expired session",
			token:          "expired-token",
			createSession:  true,
			expired:        true,
			wantStatusCode: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestApp(t)

			if tt.createSession {
				hash := sha256.Sum256([]byte(tt.token))
				tokenHash := hex.EncodeToString(hash[:])
				expiresAt := time.Now().Add(time.Hour)

				if tt.expired {
					expiresAt = time.Now().Add(-time.Hour)
				}

				err := app.session.Create(
					t.Context(),
					tokenHash,
					1,
					expiresAt,
				)

				if err != nil {
					t.Fatal(err)
				}
			}

			handlerCalled := false
			next := func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			}
			handler := app.requireAuth(next)
			req := httptest.NewRequest(
				http.MethodGet,
				"/users",
				nil,
			)

			if tt.token != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session",
					Value: tt.token,
				})
			}

			recorder := httptest.NewRecorder()
			handler(recorder, req)

			result := recorder.Result()
			defer result.Body.Close()

			if result.StatusCode != tt.wantStatusCode {
				t.Errorf(
					"status = %d, want %d",
					result.StatusCode,
					tt.wantStatusCode,
				)
			}

			if handlerCalled != tt.wantHandlerRun {
				t.Errorf(
					"handler called = %v, want %v",
					handlerCalled,
					tt.wantHandlerRun,
				)
			}

			if tt.wantStatusCode == http.StatusSeeOther {
				location := result.Header.Get("Location")
				if location != "/login" {
					t.Errorf(
						"Location = %q, want /login",
						location,
					)
				}
			}
		})
	}

}
