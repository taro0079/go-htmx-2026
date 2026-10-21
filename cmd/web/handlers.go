package main

import (
	"net/http"
	"strings"

	"github.com/taro0079/go-htmx/assets"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	err := app.html.render(w, 200, nil, "base", "pages/home.tmpl")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

var posts, _ = assets.LoadAll()

func (app *application) listPosts(w http.ResponseWriter, r *http.Request) {
	err := app.html.render(w, 200, posts, "base", "pages/posts.tmpl")

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

func (app *application) searchPost(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")

	var matches []assets.Post

	if query == "" {
		matches = posts
	} else {
		for _, p := range posts {
			if strings.Contains(p.Title, query) || strings.Contains(p.Content, query) {
				matches = append(matches, p)
			}
		}
	}

	err := app.html.render(w, 200, matches, "posts:rows", "pages/posts.tmpl")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}

}
