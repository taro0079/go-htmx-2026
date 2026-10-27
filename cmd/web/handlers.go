package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/taro0079/go-htmx/cmd/internal/models"
	"github.com/taro0079/go-htmx/cmd/users"
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

func (app *application) getSignUp(w http.ResponseWriter, r *http.Request) {
	err := app.html.render(w, http.StatusOK, nil, "base", "pages/sign_up/new.tmpl")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (app *application) postSignUp(w http.ResponseWriter, r *http.Request) {
	hased_password, err := users.HashPassword(r.FormValue("password"))

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	user := &models.User{
		Email:        r.FormValue("email"),
		PasswordHash: hased_password,
	}

	user, err = app.users.Create(r.Context(), user)

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	err = app.html.render(w, http.StatusOK, nil, "base", "pages/sign_up/create.tmpl")

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

func (app *application) newPost(w http.ResponseWriter, r *http.Request) {
	err := app.html.render(w, http.StatusOK, nil, "base", "pages/posts/new.tmpl")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (app *application) createPosts(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	content := r.FormValue("content")
	post := &models.Post{
		Title:   title,
		Content: content,
	}
	post, err := app.posts.Create(r.Context(), post)

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = app.html.render(w, http.StatusOK, post, "base", "pages/posts/create.tmpl")

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (app *application) listPosts(w http.ResponseWriter, r *http.Request) {
	isAuthenticated, err := app.isAuthenticated(r)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	posts, err := app.posts.FindAll(r.Context())
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := struct {
		Posts           []models.Post
		IsAuthenticated bool
	}{
		Posts:           posts,
		IsAuthenticated: isAuthenticated,
	}
	err = app.html.render(w, http.StatusOK, data, "base", "pages/posts/index.tmpl")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (app *application) getPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("post_id"), 10, 64)

	if err != nil || id <= 0 {
		http.NotFound(w,r)
		return
	}

	post, err := app.posts.FindById(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w,r)
		return
	}

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w,"Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = app.html.render(w, http.StatusOK, post, "base", "pages/posts/show.tmpl")

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (app *application) searchPost(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	app.logger.Info(query)

	var posts []models.Post
	var err error

	if query == "" {
		posts, err = app.posts.FindAll(r.Context())

		if err != nil {
			app.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	} else {
		posts, err = app.posts.Search(r.Context(), query)
		if err != nil {
			app.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}

	err = app.html.render(w, http.StatusOK, posts, "posts:rows", "pages/posts/index.tmpl")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (app *application) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := app.users.FindAll(r.Context())
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	err = app.html.render(w, http.StatusOK, users, "base", "pages/users/index.tmpl")

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
func (app *application) getLogin(w http.ResponseWriter, r *http.Request) {
	err := app.html.render(w, http.StatusOK, nil, "base", "pages/login/index.tmpl")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := app.users.FindByEmail(r.Context(), email)

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "メールアドレスが間違っています", http.StatusUnauthorized)
		return
	}

	err = users.CheckPassword(password, user.PasswordHash)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "パスワードが間違っています", http.StatusUnauthorized)
		return
	}

	if err := app.createSession(w, r, user.ID); err != nil {
		app.logger.Error(err.Error())

		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/users",
		http.StatusSeeOther,
	)

}

func (app *application) createSession(w http.ResponseWriter, r *http.Request, userId int64) error {
	token, tokenHash, err := GenerateSessionToken()

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return err
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	err = app.session.Create(
		r.Context(),
		tokenHash,
		userId,
		expiresAt,
	)

	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil && err != http.ErrNoCookie {
		http.Error(w, "Invalid cookie", http.StatusBadRequest)
		return
	}

	if err == nil {
		hash := sha256.Sum256([]byte(cookie.Value))
		tokenHash := hex.EncodeToString(hash[:])

		if err := app.session.Delete(r.Context(), tokenHash); err != nil {
			app.logger.Error(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/posts", http.StatusSeeOther)

}
