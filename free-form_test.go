package logs_test

import (
	"context"
	"time"

	"github.com/rclark/logs"
)

type fakeTimer struct{}

func (fakeTimer) Now() time.Time {
	return time.Time{}
}

func (fakeTimer) Since(time.Time) time.Duration {
	return time.Duration(1234)
}

func Example_freeform() {
	// Set the logging mode to freeform. This is the default.
	logs.FreeformMode()

	ctx := logs.AddEntry(context.Background())

	logs.Add(ctx,
		"name", "test",
		"count", 42,
		"flag", true,
		"messages", []string{"hello", "world"},
	)

	logs.Print(ctx, logs.WithCurrentTime(fakeTimer{}))
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","count":42,"flag":true,"messages":["hello","world"],"name":"test"}
}

func ExampleAdd_nestedKeys() {
	// Set the logging mode to freeform. This is the default.
	logs.FreeformMode()

	ctx := logs.AddEntry(context.Background())

	logs.Add(ctx,
		"user.name", "test",
		"user.count", 42,
		"user.flags.flag", true,
	)

	logs.Print(ctx, logs.WithCurrentTime(fakeTimer{}))
	// Output: {"@level":"INFO","@time":"0001-01-01T00:00:00Z","user":{"count":42,"flags":{"flag":true},"name":"test"}}
}
