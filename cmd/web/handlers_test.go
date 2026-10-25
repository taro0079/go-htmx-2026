package main

import (
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/taro0079/go-htmx/assets"
	"github.com/taro0079/go-htmx/cmd/internal/repository"
)

func TestSignUpAndLogin(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(1)
	for _, path := range []string{
		"../../db/migrations/000001_create_users.up.sql",
		"../../db/migrations/000002_create_sessions.up.sql",
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
		users:   repository.NewUserRepository(db),
		session: repository.NewSessionRepository(db),
	}
	postForm := func(path, password string, handler http.HandlerFunc) *httptest.ResponseRecorder {
		t.Helper()
		form := url.Values{"email": {"test@example.com"}, "password": {password}}
		r := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		handler(w, r)
		return w
	}
	const password = "signup-password"
	if w := postForm("/sign_up", password, app.postSignUp); w.Code != http.StatusOK {
		t.Fatalf("signup status = %d, body = %s", w.Code, w.Body.String())
	}

	t.Run("wrong password", func(t *testing.T) {
		w := postForm("/login", "wrong-password", app.login)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("login status = %d, want 401", w.Code)
		}
		if len(w.Result().Cookies()) != 0 {
			t.Fatal("failed login issued a session cookie")
		}
	})

	t.Run("signup password", func(t *testing.T) {
		w := postForm("/login", password, app.login)
		if w.Code != http.StatusSeeOther {
			t.Fatalf("login status = %d, body = %s", w.Code, w.Body.String())
		}
		location := w.Header().Get("Location")
		if location != "/users" {
			t.Fatalf("login redirect = %q, want /users", location)
		}
		cookies := w.Result().Cookies()
		if len(cookies) != 1 || cookies[0].Name != "session" || cookies[0].Value == "" {
			t.Fatal("successful login did not issue a session cookie")
		}
		r := httptest.NewRequest(http.MethodGet, location, nil)
		r.AddCookie(cookies[0])
		page := httptest.NewRecorder()
		app.requireAuth(app.listUsers)(page, r)
		if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), "test@example.com") {
			t.Fatalf("users page status = %d, body = %s", page.Code, page.Body.String())
		}
	})
}
