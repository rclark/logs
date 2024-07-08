package logs_test

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/rclark/logs"
)

func ExampleLogger() {
	logger := logs.NewLogger(logs.NewExampleLog)

	ctx := logger.AddEntry(context.Background())

	logger.Adjust(ctx, func(e *logs.ExampleLog) {
		e.Name = "test"
		e.Count = 42
		e.Flag = true
		e.Messages = []string{"hello", "world"}
	})

	logger.Print(ctx, logs.WithCurrentTime(time.Time{}))
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","name":"test","count":42,"flag":true,"messages":["hello","world"]}
}

func ExampleLogger_GetEntry() {
	logger := logs.NewLogger(logs.NewExampleLog)

	ctx := logger.AddEntry(context.Background())

	logger.Adjust(ctx, func(e *logs.ExampleLog) {
		e.Name = "test"
		e.Count = 42
		e.Flag = true
		e.Messages = []string{"hello", "world"}
	})

	e := logger.GetEntry(ctx)
	if e != nil {
		e.Flag = false
	}

	logger.Print(ctx, logs.WithCurrentTime(time.Time{}))
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","name":"test","count":42,"flag":false,"messages":["hello","world"]}
}

func ExampleLogger_logLevels() {
	// Will only print logs at or above INFO level.
	printOptions := []logs.PrintOption{
		logs.WithCurrentTime(time.Time{}),
		logs.WithLevel(logs.INFO),
	}

	logger := logs.NewLogger(logs.NewExampleLog)

	ctx := logger.AddEntry(context.Background())
	logger.Adjust(ctx, func(e *logs.ExampleLog) {
		e.Messages = []string{"info"}
	})
	logger.Print(ctx, printOptions...)

	ctx = logger.AddEntry(context.Background())
	logger.Debug(ctx)
	logger.Adjust(ctx, func(e *logs.ExampleLog) {
		e.Messages = []string{"debug"}
	})
	logger.Print(ctx, printOptions...)

	ctx = logger.AddEntry(context.Background())
	logger.Fatal(ctx)
	logger.Print(ctx, printOptions...)
	// Output:
	// {"@level":"INFO","@time":"0001-01-01T00:00:00Z","name":"","count":0,"flag":false,"messages":["info"]}
	// {"@level":"FATAL","@time":"0001-01-01T00:00:00Z","name":"","count":0,"flag":false}
}

var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func ExampleLogger_Middleware() {
	logger := logs.NewLogger(logs.NewExampleLog)

	middleware := logger.Middleware(logs.WithTiming(time.Time{}, time.Duration(1234)))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/path", nil)

	middleware(handler).ServeHTTP(w, r)
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","name":"test","count":42,"flag":true,"messages":["hello","world"]}
}
