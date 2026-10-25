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
	logger *slog.Logger
	html   *htmlRenderer
	users  *repository.UserRepository
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
		logger: logger,
		html:   htmlRenderer,
		users:  repository.NewUserRepository(db),
	}

	fileserver := http.FileServerFS(assets.StaticFiles)

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static", fileserver))
	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /posts", app.listPosts)
	mux.HandleFunc("GET /posts/search", app.searchPost)
	mux.HandleFunc("GET /posts/{post_id}", app.getPost)
	mux.HandleFunc("GET /gopher", app.gopher)
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
