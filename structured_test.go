//go:build structuredlogs
// +build structuredlogs

package logs_test

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/rclark/logs"
)

func Example() {
	ctx := logs.
		NewLogger(logs.NewExampleLog).
		Set(context.Background())

	ctx = logs.AddEntry[logs.ExampleLog](ctx)

	logs.Adjust(ctx, func(e *logs.ExampleLog) {
		e.Name = "test"
		e.Count = 42
		e.Flag = true
		e.Messages = []string{"hello", "world"}
	})

	logs.Print[logs.ExampleLog](ctx, logs.WithCurrentTime(time.Time{}))
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","name":"test","count":42,"flag":true,"messages":["hello","world"]}
}

func ExampleAdjust() {
	ctx := logs.
		NewLogger(logs.NewExampleLog).
		Set(context.Background())

	ctx = logs.AddEntry[logs.ExampleLog](ctx)

	logs.Adjust(ctx, func(e *logs.ExampleLog) {
		e.Name = "test"
		e.Count = 42
		e.Flag = true
		e.Messages = []string{"hello", "world"}
	})

	logs.Print[logs.ExampleLog](ctx, logs.WithCurrentTime(time.Time{}))
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","name":"test","count":42,"flag":true,"messages":["hello","world"]}
}

func ExampleGetEntry() {
	ctx := logs.
		NewLogger(logs.NewExampleLog).
		Set(context.Background())

	ctx = logs.AddEntry[logs.ExampleLog](ctx)

	logs.Adjust(ctx, func(e *logs.ExampleLog) {
		e.Name = "test"
		e.Count = 42
		e.Flag = true
		e.Messages = []string{"hello", "world"}
	})

	e := logs.GetEntry[logs.ExampleLog](ctx)
	if e == nil {
		log.Fatal("entry not found")
	}

	e.Flag = false

	logs.Print[logs.ExampleLog](ctx, logs.WithCurrentTime(time.Time{}))
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","name":"test","count":42,"flag":false,"messages":["hello","world"]}
}

func Example_logLevels() {
	// Will only print logs at or above INFO level.
	printOptions := []logs.PrintOption{
		logs.WithCurrentTime(time.Time{}),
		logs.WithLevel(logs.INFO),
	}

	newCtx := func() context.Context {
		return logs.
			NewLogger(logs.NewExampleLog).
			Set(context.Background())
	}

	ctx := logs.AddEntry[logs.ExampleLog](newCtx(), logs.WithDefaultLevel(logs.DEBUG))
	logs.Adjust(ctx, func(e *logs.ExampleLog) {
		e.Messages = []string{"debug"}
	})
	logs.Print[logs.ExampleLog](ctx, printOptions...)

	ctx = logs.AddEntry[logs.ExampleLog](newCtx(), logs.WithDefaultLevel(logs.DEBUG))
	logs.Info[logs.ExampleLog](ctx)
	logs.Adjust(ctx, func(e *logs.ExampleLog) {
		e.Messages = []string{"info"}
	})
	logs.Print[logs.ExampleLog](ctx, printOptions...)

	ctx = logs.AddEntry[logs.ExampleLog](newCtx(), logs.WithDefaultLevel(logs.DEBUG))
	logs.Warn[logs.ExampleLog](ctx)
	logs.Print[logs.ExampleLog](ctx, printOptions...)

	ctx = logs.AddEntry[logs.ExampleLog](newCtx(), logs.WithDefaultLevel(logs.DEBUG))
	logs.Error[logs.ExampleLog](ctx)
	logs.Adjust(ctx, func(e *logs.ExampleLog) {
		e.Messages = []string{"error"}
	})
	logs.Print[logs.ExampleLog](ctx, printOptions...)
	// Output:
	// {"@level":"INFO","@time":"0001-01-01T00:00:00Z","name":"","count":0,"flag":false,"messages":["info"]}
	// {"@level":"WARN","@time":"0001-01-01T00:00:00Z","name":"","count":0,"flag":false}
	// {"@level":"ERROR","@time":"0001-01-01T00:00:00Z","name":"","count":0,"flag":false,"messages":["error"]}
}

var structuredHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	logger := logs.Get[logs.ExampleLog](r.Context())
	if logger == nil {
		log.Fatal("logger not found")
	}

	logger.Adjust(r.Context(), func(e *logs.ExampleLog) {
		e.Name = "test"
		e.Count = 42
		e.Flag = true
		e.Messages = []string{"hello", "world"}
	})
})

func ExampleMiddleware() {
	middleware := logs.Middleware(logs.NewExampleLog, logs.WithTiming(time.Time{}, time.Duration(1234)))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/path", nil)

	middleware(structuredHandler).ServeHTTP(w, r)
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","name":"test","count":42,"flag":true,"messages":["hello","world"]}
}
