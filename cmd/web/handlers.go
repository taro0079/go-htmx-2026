package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/taro0079/go-htmx/assets"
)

func (app *application) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)
		app.logger.Info(
			"HTTP Request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
		)
	})
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	err := app.html.render(w, http.StatusOK, nil, "base", "pages/home.tmpl")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (app *application) gopher(w http.ResponseWriter, r *http.Request) {
	width := 100
	err := app.html.render(w, http.StatusOK, width, "partial:image:gopher")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (app *application) listPosts(w http.ResponseWriter, r *http.Request) {
	var posts, _ = assets.LoadAll()
	err := app.html.render(w, http.StatusOK, posts, "base", "pages/posts.tmpl")

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (app *application) getPost(w http.ResponseWriter, r *http.Request) {
	var post, err = assets.GetPost(r.PathValue("post_id"))
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	err = app.html.render(w, http.StatusOK, post, "base", "pages/post.tmpl")

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (app *application) searchPost(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	var posts, _ = assets.LoadAll()

	var matches []assets.Post

	if query == "" {
		matches = posts
	} else {
		for _, p := range posts {
			if strings.Contains(p.Title, query) || strings.Contains(string(p.Content), query) {
				matches = append(matches, p)
			}
		}
	}

	err := app.html.render(w, http.StatusOK, matches, "posts:rows", "pages/posts.tmpl")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

}

// func (app *application) login(w http.ResponseWriter, r *http.Request) {
// 	if err := r.ParseForm(); err != nil {
// 		http.Error(w, "Invalid form", http.StatusBadRequest)
// 		return
// 	}
//
// 	email := r.PostForm.Get("email")
// 	password := r.PostForm.Get("password")
//
// 	var userID int64
// 	var passwordHash string
//
// 	err:=app.db.QueryRowContext(
// 		r.Context(),
//
// 		)
// }
