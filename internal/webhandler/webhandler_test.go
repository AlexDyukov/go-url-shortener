package webhandler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	service "github.com/alexdyukov/go-url-shortener/internal/service"
	storage "github.com/alexdyukov/go-url-shortener/internal/storage"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

const counter = 100

var savedURL string = "https://www.google.com/search?q=there+is+search+string"
var savedID string
var nonsavedID string
var testRouter *httprouter.Router

func TestMain(m *testing.M) {
	// Init
	testStorage := storage.NewInMemory()
	testService := service.NewURLShortener(testStorage)

	id, err := testService.SaveURL(savedURL)
	if err != nil {
		panic("cannot save predefined valid url")
	}
	savedID = fmt.Sprint(id)

	nonsavedID = savedID
	for i := 0; nonsavedID == savedID && i < counter; i += 1 {
		nonsavedID = shuffle(i, nonsavedID)
	}
	if nonsavedID == savedID {
		panic("cannot generate valid but not saved ID")
	}

	testWebHandler := NewWebHandler(testService)
	testRouter = httprouter.New()
	testRouter.GET("/:id", testWebHandler.HandlerGet)
	testRouter.POST("/", testWebHandler.HandlerPost)

	// Run tests
	exitVal := m.Run()

	os.Exit(exitVal)
}

// non actual shuffle, just a little changing of input string
func shuffle(seed int, in string) string {
	if seed < 0 {
		seed = -seed
	}

	inRune := []rune(in)
	randomIndex := seed % len(inRune)
	inRune[len(inRune)-1] = inRune[randomIndex]

	return string(inRune)
}

func TestWebHandler_Get(t *testing.T) {
	type want struct {
		statusCode int
		location   string
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "test GET valid non saved ID",
			request: "/" + nonsavedID,
			want: want{
				statusCode: http.StatusNotFound,
				location:   "",
			},
		},
		{
			name:    "test GET valid saved ID",
			request: "/" + savedID,
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   savedURL,
			},
		},
		{
			name:    "test GET invalid param",
			request: "/-1",
			want: want{
				statusCode: http.StatusBadRequest,
				location:   "",
			},
		},
	}

	// run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()

			testRouter.ServeHTTP(w, r)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}

func TestWebHandler_Post(t *testing.T) {
	type want struct {
		statusCode       int
		locationEndsWith string
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "test POST valid url",
			request: savedURL,
			want: want{
				statusCode:       http.StatusCreated,
				locationEndsWith: savedID,
			},
		},
		{
			name:    "test POST invalid request",
			request: "",
			want: want{
				statusCode:       http.StatusBadRequest,
				locationEndsWith: "",
			},
		},
	}

	// run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.request))
			w := httptest.NewRecorder()

			testRouter.ServeHTTP(w, r)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, result.StatusCode, tt.want.statusCode)
		})
	}
}
