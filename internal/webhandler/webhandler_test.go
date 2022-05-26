package webhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	service "github.com/alexdyukov/go-url-shortener/internal/service"
	storage "github.com/alexdyukov/go-url-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
)

const shuffleCounter = 100

var baseURL string = "http://localhost:8080"
var savedURL string = "https://www.google.com/search?q=there+is+search+string"
var savedID string
var nonsavedID string
var testWebHandler *WebHandler

func TestMain(m *testing.M) {
	// Init
	testStorage := storage.NewInMemory()
	testService := service.NewURLShortener(testStorage, baseURL)

	savedURL, err := testService.SaveURL(context.Background(), savedURL)
	if err != nil {
		panic("cannot save predefined valid url")
	}
	savedID = strings.TrimPrefix(savedURL, baseURL)
	savedID = strings.TrimPrefix(savedID, "/")
	savedID = strings.TrimSuffix(savedID, "/")

	nonsavedID = savedID
	for i := 0; nonsavedID == savedID && i < shuffleCounter; i += 1 {
		nonsavedID = shuffle(i, nonsavedID)
	}
	if nonsavedID == savedID {
		panic("cannot generate valid but not saved ID")
	}

	testWebHandler = NewWebHandler(testService, "testtesttesttest")

	// Run tests
	os.Exit(m.Run())
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

func TestWebHandler_GetRoot(t *testing.T) {
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
			request: "/a",
			want: want{
				statusCode: http.StatusNotFound,
				location:   "",
			},
		},
	}

	// run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()

			testWebHandler.router.ServeHTTP(w, r)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}

func TestWebHandler_PostRoot(t *testing.T) {
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
			name:    "valid url",
			request: savedURL,
			want: want{
				statusCode:       http.StatusCreated,
				locationEndsWith: savedID,
			},
		},
		{
			name:    "invalid request",
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

			testWebHandler.router.ServeHTTP(w, r)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
		})
	}
}

func TestWebHandler_PostApiShorten(t *testing.T) {
	type want struct {
		statusCode int
		response   string
	}
	type request struct {
		contentType string `json:"-"`
		URL         string `json:"url"`
	}
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "valid url",
			request: request{
				URL:         savedURL,
				contentType: "application/json",
			},
			want: want{
				statusCode: http.StatusCreated,
				response:   "{\"result\":\"" + baseURL + "/" + savedID + "\"}",
			},
		},
		{
			name: "invalid url",
			request: request{
				URL:         "",
				contentType: "application/json",
			},
			want: want{
				statusCode: http.StatusBadRequest,
				response:   "",
			},
		},
		{
			name: "invalid content type",
			request: request{
				URL:         savedURL,
				contentType: "pikachu",
			},
			want: want{
				statusCode: http.StatusUnsupportedMediaType,
				response:   "",
			},
		},
	}

	// run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			r := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
			r.Header.Set("Content-Type", tt.request.contentType)
			w := httptest.NewRecorder()

			testWebHandler.router.ServeHTTP(w, r)
			result := w.Result()
			defer result.Body.Close()

			output, err := io.ReadAll(result.Body)
			assert.Nil(t, err)

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			// tests which does not need response check goes with tt.want.response=="" because any string contains empty string
			if !strings.Contains(string(output), tt.want.response) {
				t.Fatal("http response does not contain response. Want: '" + tt.want.response + "' but got '" + string(output) + "'")
			}
		})
	}
}
