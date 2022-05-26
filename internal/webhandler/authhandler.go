package webhandler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	service "github.com/alexdyukov/go-url-shortener/internal/service"
	storage "github.com/alexdyukov/go-url-shortener/internal/storage"
)

const userCookieName = "URL-Shortener-User"

func newAuthHandler(encryptor *Encryptor) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := getCookiedUser(encryptor, r)
			cookie := &http.Cookie{
				Name:  userCookieName,
				Value: makeCookiedUser(encryptor, user),
			}
			http.SetCookie(w, cookie)
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), service.UserCtxKey{}, user)))
		})
	}
}

func getCookiedUser(encryptor *Encryptor, r *http.Request) storage.User {
	cookieUserID, err := r.Cookie(userCookieName)
	if err != nil && r.Method == http.MethodPost {
		return storage.NewUser()
	}
	if err != nil {
		return storage.DefaultUser
	}

	userStr, err := encryptor.Decode(cookieUserID.Value)
	if err != nil {
		return storage.NewUser()
	}

	user, err := storage.ParseUser(strings.TrimSpace(userStr))
	if err != nil {
		return storage.NewUser()
	}

	return user
}

func makeCookiedUser(encryptor *Encryptor, user storage.User) string {
	return encryptor.Encode(fmt.Sprint(user))
}
