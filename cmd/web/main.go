package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/taro0079/go-htmx/assets"
	"github.com/taro0079/go-htmx/cmd/internal/repository"
)

type application struct {
	logger  *slog.Logger
	html    *htmlRenderer
	users   *repository.UserRepository
	session *repository.SessionRepository
}

func (app *application) requireAuth(
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(
				w,
				r,
				"/login",
				http.StatusSeeOther,
			)
			return
		}
		hash := sha256.Sum256([]byte(cookie.Value))
		tokenHash := hex.EncodeToString(hash[:])
		userId, err := app.session.FindUserID(
			r.Context(),
			tokenHash,
		)

		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				app.logger.Error(err.Error())
				http.Error(
					w,
					"Internal Server Error",
					http.StatusInternalServerError,
				)
				return
			}

			http.Redirect(
				w,
				r,
				"/login",
				http.StatusSeeOther,
			)
			return
		}

		_ = userId

		next(w, r)
	}

}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	htmlRenderer, err := newHTMLRenderer(assets.HTMLFiles, "base.tmpl", "partials/*.tmpl")

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", "./app.db")

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	app := &application{
		logger:  logger,
		html:    htmlRenderer,
		users:   repository.NewUserRepository(db),
		session: repository.NewSessionRepository(db),
	}

	fileserver := http.FileServerFS(assets.StaticFiles)

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static", fileserver))
	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /posts", app.listPosts)
	mux.HandleFunc("GET /posts/search", app.searchPost)
	mux.HandleFunc("GET /posts/{post_id}", app.getPost)
	mux.HandleFunc("GET /gopher", app.gopher)
	mux.HandleFunc("GET /users", app.requireAuth(app.listUsers))
	mux.HandleFunc("GET /login", app.getLogin)
	mux.HandleFunc("POST /login", app.login)
	mux.HandleFunc("GET /sign_up", app.getSignUp)
	mux.HandleFunc("POST /sign_up", app.postSignUp)
	handler := app.loggingMiddleware(mux)

	logger.Info("starting server", "port", 5051)
	err = http.ListenAndServe(":5051", handler)

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
