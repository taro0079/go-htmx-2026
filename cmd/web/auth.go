package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
)

func GenerateSessionToken() (string, string, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	token := hex.EncodeToString(bytes)

	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	return token, tokenHash, nil
}

func (app *application) isAuthenticated(r *http.Request) (bool, error) {
	cookie, err := r.Cookie("session")
	if errors.Is(err, http.ErrNoCookie) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	hash := sha256.Sum256([]byte(cookie.Value))
	tokenHash := hex.EncodeToString(hash[:])
	_, err = app.session.FindUserID(r.Context(), tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
