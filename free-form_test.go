//go:build !structuredlogs
// +build !structuredlogs

package logs_test

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/rclark/logs"
)

func Example() {
	ctx := logs.AddEntry(context.Background())

	logs.Add(ctx,
		"name", "test",
		"count", 42,
		"flag", true,
		"messages", []string{"hello", "world"},
	)

	logs.Print(ctx, logs.WithCurrentTime(time.Time{}))
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","count":42,"flag":true,"messages":["hello","world"],"name":"test"}
}

func ExampleAdd_nestedKeys() {
	ctx := logs.AddEntry(context.Background())

	logs.Add(ctx,
		"user.name", "test",
		"user.count", 42,
		"user.flags.flag", true,
	)

	logs.Print(ctx, logs.WithCurrentTime(time.Time{}))
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","user":{"count":42,"flags":{"flag":true},"name":"test"}}
}

func ExampleAppend() {
	ctx := logs.AddEntry(context.Background())

	logs.Add(ctx,
		"messages", []string{"hello", "world"},
	)

	logs.Append(ctx, "messages", "goodbye")

	logs.Print(ctx, logs.WithCurrentTime(time.Time{}))
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","messages":["hello","world","goodbye"]}
}

func ExampleAdjust() {
	ctx := logs.AddEntry(context.Background())

	logs.Add(ctx,
		"count", 42,
	)

	logs.Adjust(ctx, func(fe *logs.FreeformEntry) {
		(*fe)["count"] = 43
	})

	logs.Print(ctx, logs.WithCurrentTime(time.Time{}))
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","count":43}
}

func ExampleGetEntry() {
	ctx := logs.AddEntry(context.Background())

	logs.Add(ctx,
		"count", 42,
	)

	entry := logs.GetEntry(ctx)
	if entry == nil {
		log.Fatal("entry not found")
	}

	(*entry)["count"] = 43

	logs.Print(ctx, logs.WithCurrentTime(time.Time{}))
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","count":43}
}

func Example_logLevels() {
	// Will only print logs at or above INFO level.
	printOptions := []logs.PrintOption{
		logs.WithCurrentTime(time.Time{}),
		logs.WithLevel(logs.INFO),
	}

	ctx := logs.AddEntry(context.Background(), logs.WithDefaultLevel(logs.DEBUG))
	logs.Add(ctx, "this one", "is debug")
	logs.Print(ctx, printOptions...)

	ctx = logs.AddEntry(context.Background(), logs.WithDefaultLevel(logs.DEBUG))
	logs.Info(ctx)
	logs.Add(ctx, "this one", "is info")
	logs.Print(ctx, printOptions...)

	ctx = logs.AddEntry(context.Background(), logs.WithDefaultLevel(logs.DEBUG))
	logs.Warn(ctx)
	logs.Print(ctx, printOptions...)

	ctx = logs.AddEntry(context.Background(), logs.WithDefaultLevel(logs.DEBUG))
	logs.Error(ctx)
	logs.Add(ctx, "this one", "is error")
	logs.Print(ctx, printOptions...)
	// Output:
	// {"@level":"INFO","@time":"0001-01-01T00:00:00Z","this one":"is info"}
	// {"@level":"WARN","@time":"0001-01-01T00:00:00Z"}
	// {"@level":"ERROR","@time":"0001-01-01T00:00:00Z","this one":"is error"}
}

var freeformHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func ExampleMiddleware() {
	middleware := logs.Middleware(logs.WithTiming(time.Time{}, time.Duration(1234)))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/path", strings.NewReader("bar"))

	middleware(freeformHandler).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		log.Fatal("unexpected status code")
	}
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","@http":{"method":"POST","path":"/path","duration":1234},"foo":"bar","messages":["hello","world"]}
}

func ExampleMiddleware_withBody() {
	middleware := logs.Middleware(logs.WithTiming(time.Time{}, time.Duration(1234)), logs.WithBody())

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/path", strings.NewReader("bar"))

	middleware(freeformHandler).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		log.Fatal("unexpected status code")
	}
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","@http":{"method":"POST","path":"/path","body":"bar","duration":1234},"foo":"bar","messages":["hello","world"]}
}

func ExampleMiddleware_someHeaders() {
	middleware := logs.Middleware(logs.WithTiming(time.Time{}, time.Duration(1234)), logs.WithHeaders("X-Header"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/path", strings.NewReader("bar"))
	r.Header.Set("X-Header", "x")
	r.Header.Set("Y-Header", "y")

	middleware(freeformHandler).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		log.Fatal("unexpected status code")
	}
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","@http":{"method":"POST","path":"/path","headers":{"X-Header":"x"},"duration":1234},"foo":"bar","messages":["hello","world"]}
}

func ExampleMiddleware_allHeaders() {
	middleware := logs.Middleware(logs.WithTiming(time.Time{}, time.Duration(1234)), logs.WithAllHeaders())

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/path", strings.NewReader("bar"))
	r.Header.Set("X-Header", "x")
	r.Header.Set("Y-Header", "y")

	middleware(freeformHandler).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		log.Fatal("unexpected status code")
	}
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","@http":{"method":"POST","path":"/path","headers":{"X-Header":"x","Y-Header":"y"},"duration":1234},"foo":"bar","messages":["hello","world"]}
}
