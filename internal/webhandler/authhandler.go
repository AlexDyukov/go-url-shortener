package webhandler

import (
	"fmt"
	"log"
	"net/http"

	service "github.com/alexdyukov/go-url-shortener/internal/service"
	storage "github.com/alexdyukov/go-url-shortener/internal/storage"
)

const userCookieName = "URL-Shortener-User"

func newAuthHandler(encryptor *Encryptor, repo service.Repository) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			user, err := getCookiedUser(encryptor, r)
			if err != nil && r.Method == http.MethodPost {
				user, err = repo.NewUser(ctx)
				if err != nil {
					log.Println("webhandler: authhandler: cannot call repo.NewUser():", err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				//http.ResponseWriter
				cookie := &http.Cookie{
					Name:  userCookieName,
					Value: makeCookiedUser(encryptor, user),
				}
				http.SetCookie(w, cookie)
			}

			//http.Request
			r = r.WithContext(storage.PutUser(ctx, user))

			next.ServeHTTP(w, r)
		})
	}
}

func getCookiedUser(encryptor *Encryptor, r *http.Request) (storage.User, error) {
	cookieUserID, err := r.Cookie(userCookieName)
	if err != nil {
		return storage.DefaultUser, err
	}

	userStr, err := encryptor.Decode(cookieUserID.Value)
	if err != nil {
		return storage.DefaultUser, err
	}

	user, err := storage.ParseUser(userStr)
	if err != nil {
		return storage.DefaultUser, err
	}

	return user, nil
}

func makeCookiedUser(encryptor *Encryptor, user storage.User) string {
	return encryptor.Encode(fmt.Sprint(user))
}
