package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/taro0079/go-htmx/assets"
	"github.com/taro0079/go-htmx/cmd/internal/models"
	"github.com/taro0079/go-htmx/cmd/internal/repository"
)

func TestPostsLoginVisibility(t *testing.T) {
	for _, tt := range []struct {
		name     string
		token    string
		expiry   time.Duration
		wantLink bool
	}{
		{name: "anonymous"},
		{name: "invalid session", token: "unknown"},
		{name: "expired session", token: "expired", expiry: -time.Hour},
		{name: "logged in", token: "valid", expiry: time.Hour, wantLink: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db, err := sql.Open("sqlite3", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { db.Close() })
			db.SetMaxOpenConns(1)
			for _, path := range []string{
				"../../db/migrations/000001_create_users.up.sql",
				"../../db/migrations/000002_create_sessions.up.sql",
				"../../db/migrations/000003_create_posts.up.sql",
			} {
				migration, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(string(migration)); err != nil {
					t.Fatal(err)
				}
			}
			renderer, err := newHTMLRenderer(assets.HTMLFiles, "base.tmpl", "partials/*.tmpl")
			if err != nil {
				t.Fatal(err)
			}
			app := &application{
				logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
				html:    renderer,
				session: repository.NewSessionRepository(db),
				posts:   repository.NewPostRepository(db),
			}
			if _, err := app.posts.Create(t.Context(), &models.Post{Title: "Example post", Content: "content"}); err != nil {
				t.Fatal(err)
			}
			if tt.expiry != 0 {
				hash := sha256.Sum256([]byte(tt.token))
				if err := app.session.Create(t.Context(), hex.EncodeToString(hash[:]), 1, time.Now().Add(tt.expiry)); err != nil {
					t.Fatal(err)
				}
			}
			req := httptest.NewRequest(http.MethodGet, "/posts", nil)
			if tt.token != "" {
				req.AddCookie(&http.Cookie{Name: "session", Value: tt.token})
			}
			w := httptest.NewRecorder()
			app.listPosts(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
			}
			if got := strings.Contains(w.Body.String(), `href="/posts/new"`); got != tt.wantLink {
				t.Errorf("New Post link visible = %v, want %v", got, tt.wantLink)
			}
			if !strings.Contains(w.Body.String(), "Example post") {
				t.Error("post missing from list")
			}

			// Search renders the rows partial with a slice rather than the full page data.
			w = httptest.NewRecorder()
			app.searchPost(w, httptest.NewRequest(http.MethodGet, "/posts/search?query=", nil))
			if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Example post") {
				t.Fatalf("search status = %d, body = %s", w.Code, w.Body.String())
			}
		})
	}
}
