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

func (app *application) gopher(w http.ResponseWriter, r *http.Request) {
	width := 100
	err := app.html.render(w, http.StatusOK, width, "partial:image:gopher")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

func (app *application) listPosts(w http.ResponseWriter, r *http.Request) {
	var posts, _ = assets.LoadAll()
	err := app.html.render(w, 200, posts, "base", "pages/posts.tmpl")

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

func (app *application) getPost(w http.ResponseWriter, r *http.Request) {
	var post, err = assets.GetPost(r.PathValue("post_id"))
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}

	err = app.html.render(w, 200, post, "base", "pages/post.tmpl")

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
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

	err := app.html.render(w, 200, matches, "posts:rows", "pages/posts.tmpl")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}

}
