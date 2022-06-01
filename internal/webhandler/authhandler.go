package webhandler

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
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
			requestContext := r.Context()

			user, err := encryptor.GetUser(r)
			if err != nil && r.Method != http.MethodGet {
				user, err = repo.NewUser(requestContext)
				if err != nil {
					log.Println("webhandler: authhandler: cannot call repo.NewUser():", err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}

			http.SetCookie(w, encryptor.Cookie(user))
			r = r.WithContext(storage.PutUser(requestContext, user))
			next.ServeHTTP(w, r)
		})
	}
}

type Encryptor struct {
	aesblock cipher.Block
	key      []byte
}

func newEncryptor(key []byte) *Encryptor {
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal("cannot initialize Encryptor:", err.Error())
	}

	return &Encryptor{key: key, aesblock: aesblock}
}

func (e *Encryptor) GetUser(r *http.Request) (storage.User, error) {
	cookiedUser, err := r.Cookie(userCookieName)
	if err != nil {
		return storage.DefaultUser, err
	}

	stringedUser, err := e.decode(cookiedUser.Value)
	if err != nil {
		return storage.DefaultUser, err
	}

	user, err := storage.ParseUser(stringedUser)
	if err != nil {
		return storage.DefaultUser, err
	}

	return user, nil
}

func (e *Encryptor) Cookie(user storage.User) *http.Cookie {
	return &http.Cookie{
		Name:  userCookieName,
		Value: e.encode(fmt.Sprint(user)),
	}
}

func (e *Encryptor) encode(str string) string {
	decoded := []byte(str)
	encoded := make([]byte, e.aesblock.BlockSize())

	for len(encoded) > len(decoded) {
		decoded = append(decoded, ' ')
	}

	e.aesblock.Encrypt(encoded, decoded)

	return base64.StdEncoding.EncodeToString(encoded)
}

func (e *Encryptor) decode(str string) (string, error) {
	encoded, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}

	decoded := make([]byte, e.aesblock.BlockSize())

	for len(decoded) > len(encoded) {
		encoded = append(encoded, ' ')
	}

	e.aesblock.Decrypt(decoded, []byte(encoded))

	return string(decoded), nil
}
