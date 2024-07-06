package logs_test

import (
	"context"

	"github.com/rclark/logs"
)

// entry is the structure for each log entry you wish to capture.
type entry struct {
	Name     string
	Count    int
	Flag     bool
	Messages []string
}

// newEntry defines how to create an empty, mutable version of a log entry.
func newEntry() *entry {
	return &entry{}
}

func Example_structured() {
	// Set the logging mode to structured.
	logs.StructuredMode(newEntry)

	ctx := logs.AddEntry(context.Background())

	logs.Adjust(ctx, func(e *entry) {
		e.Name = "test"
		e.Count = 42
		e.Flag = true
		e.Messages = []string{"hello", "world"}
	})

	logs.Print(ctx, logs.WithCurrentTime(fakeTimer{}))
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","Name":"test","Count":42,"Flag":true,"Messages":["hello","world"]}
}

func Example_structuredLogger() {
	logger := logs.NewLogger(newEntry)

	ctx := logger.AddEntry(context.Background())

	// You can adjust with the logger.
	logger.Adjust(ctx, func(e *entry) {
		e.Name = "test"
		e.Count = 42
		e.Flag = true
		e.Messages = []string{"hello", "world"}
	})

	// You can still adjust without having to use the logger.
	logs.Adjust(ctx, func(e *entry) {
		e.Flag = false
	})

	logger.Print(ctx, logs.WithCurrentTime(fakeTimer{}))
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","Name":"test","Count":42,"Flag":false,"Messages":["hello","world"]}
}
