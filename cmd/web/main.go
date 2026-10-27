package main

import (
	"database/sql"
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
	posts   *repository.PostRepository
}

func (app *application) requireAuth(
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		authenticated, err := app.isAuthenticated(r)
		if err != nil {
			app.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if !authenticated {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

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
		posts:   repository.NewPostRepository(db),
	}

	fileserver := http.FileServerFS(assets.StaticFiles)

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static", fileserver))
	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /posts", app.listPosts)
	mux.HandleFunc("GET /posts/new", app.requireAuth(app.newPost))
	mux.HandleFunc("GET /posts/search", app.searchPost)
	mux.HandleFunc("POST /posts", app.requireAuth(app.createPosts))
	mux.HandleFunc("GET /posts/{post_id}", app.loadPost)
	mux.HandleFunc("GET /posts/{post_id}/content", app.getPost)
	mux.HandleFunc("GET /gopher", app.gopher)
	mux.HandleFunc("GET /users", app.requireAuth(app.listUsers))
	mux.HandleFunc("GET /login", app.getLogin)
	mux.HandleFunc("POST /login", app.login)
	mux.HandleFunc("POST /logout", app.requireAuth(app.logout))
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
