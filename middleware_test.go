package logs_test

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/rclark/logs"
)

var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusInternalServerError)
	}

	logs.Add(r.Context(),
		"messages", []string{"hello", "world"},
		"foo", string(body),
	)
})

func ExampleMiddleware_freeform() {
	logs.FreeformMode()

	middleware := logs.Middleware(logs.WithTimer(fakeTimer{}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/path", strings.NewReader("bar"))

	middleware(handler).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		log.Fatal("unexpected status code")
	}
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","@http":{"method":"POST","path":"/path","duration":1234},"foo":"bar","messages":["hello","world"]}
}

func ExampleMiddleware_withBody() {
	logs.FreeformMode()

	middleware := logs.Middleware(logs.WithTimer(fakeTimer{}), logs.WithBody())

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/path", strings.NewReader("bar"))

	middleware(handler).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		log.Fatal("unexpected status code")
	}
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","@http":{"method":"POST","path":"/path","body":"bar","duration":1234},"foo":"bar","messages":["hello","world"]}
}

func ExampleMiddleware_someHeaders() {
	logs.FreeformMode()

	middleware := logs.Middleware(logs.WithTimer(fakeTimer{}), logs.WithHeaders("X-Header"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/path", strings.NewReader("bar"))
	r.Header.Set("X-Header", "x")
	r.Header.Set("Y-Header", "y")

	middleware(handler).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		log.Fatal("unexpected status code")
	}
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","@http":{"method":"POST","path":"/path","headers":{"X-Header":"x"},"duration":1234},"foo":"bar","messages":["hello","world"]}
}

func ExampleMiddleware_allHeaders() {
	logs.FreeformMode()

	middleware := logs.Middleware(logs.WithTimer(fakeTimer{}), logs.WithAllHeaders())

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/path", strings.NewReader("bar"))
	r.Header.Set("X-Header", "x")
	r.Header.Set("Y-Header", "y")

	middleware(handler).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		log.Fatal("unexpected status code")
	}
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","@http":{"method":"POST","path":"/path","headers":{"X-Header":"x","Y-Header":"y"},"duration":1234},"foo":"bar","messages":["hello","world"]}
}
