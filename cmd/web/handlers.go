package main

import (
	"net/http"

	"github.com/taro0079/go-htmx/assets"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	err := app.html.render(w, 200, nil, "base", "pages/home.tmpl")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

type post struct {
	Title   string
	Content string
}

var posts, _ = assets.LoadAll()

func (app *application) listPosts(w http.ResponseWriter, r *http.Request) {
	err := app.html.render(w, 200, posts, "base", "pages/posts.tmpl")

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}
