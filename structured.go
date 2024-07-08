//go:build structuredlogs
// +build structuredlogs

/*
Package logs enables structured, context-aware logging with type enforcement.
When built with the "structuredlogs" tag, you must define a custom type for your
logs, create a [Logger] for that type, and use [Logger.Set] to add it to the
context. You may manipulate log entries using the [Logger]'s methods or use the
package-level functions that work with your type.

Without the "structuredlogs" tag, logging is more flexible, allowing you to add
any key-value pairs to a log entry, represented as a map with string keys.
*/
package logs

import (
	"context"
	"net/http"
)

// AddEntry adds a log entry to the context. You must have first created a
// [Logger] and added it to the context using the [Logger.Set] function.
func AddEntry[T any](ctx context.Context, opts ...Option) context.Context {
	if logger := Get[T](ctx); logger != nil {
		return logger.AddEntry(ctx, opts...)
	}

	return ctx
}

// Print prints the log entry in the context as JSON. The function will return
// false if no log entry is found.
//
// The default output is os.Stdout. You can change this by providing a custom
// io.Writer using the [WithOutput] option.
//
// The default log level for printing is INFO. You can change this by providing
// a custom log level using the [WithLevel] option.
//
// The default timer is the system clock. You can change this by providing a
// custom timer using the [WithCurrentTime] option. This is useful if you need
// to write tests to confirm that your application is logging as expected, as
// your custom timer can be used to control the log entry's "@time" property.
//
// If the you've provided a custom struct for your log entries and it fails to
// marshal to JSON using the standard json.Marshal(), the function will write an
// error message to os.Stderr and return false.
func Print[T any](ctx context.Context, opts ...PrintOption) bool {
	if logger := Get[T](ctx); logger != nil {
		return logger.Print(ctx, opts...)
	}

	return false
}

// Debug sets the log entry's level to DEBUG. The function will return false if
// no log entry is found in the context.
func Debug[T any](ctx context.Context) bool {
	if logger := Get[T](ctx); logger != nil {
		return logger.Debug(ctx)
	}

	return false
}

// Info sets the log entry's level to INFO. The function will return false if no
// log entry is found in the context.
func Info[T any](ctx context.Context) bool {
	if logger := Get[T](ctx); logger != nil {
		return logger.Info(ctx)
	}

	return false
}

// Warn sets the log entry's level to WARN. The function will return false if no
// log entry is found in the context.
func Warn[T any](ctx context.Context) bool {
	if logger := Get[T](ctx); logger != nil {
		return logger.Warn(ctx)
	}

	return false
}

// Error sets the log entry's level to ERROR. The function will return false if
// no log entry is found in the context.
func Error[T any](ctx context.Context) bool {
	if logger := Get[T](ctx); logger != nil {
		return logger.Error(ctx)
	}

	return false
}

// Fatal sets the log entry's level to FATAL. The function will return false if
// no log entry is found in the context.
func Fatal[T any](ctx context.Context) bool {
	if logger := Get[T](ctx); logger != nil {
		return logger.Fatal(ctx)
	}

	return false
}

// GetEntry gets the log entry from the context for direct manipulation. The
// function will return nil if no log entry of the correct type is found.
func GetEntry[T any](ctx context.Context) *T {
	if entry := getEntry[T](ctx); entry != nil {
		return entry.data
	}

	return nil
}

// Adjuster is a function that adjust a log entry.
type Adjuster[T any] func(*T)

// Adjust mutates the log entry in the context. The function will return false
// if no log entry of the correct type is found in the context.
func Adjust[T any](ctx context.Context, fns ...Adjuster[T]) bool {
	if entry := getEntry[T](ctx); entry != nil {
		for _, fn := range fns {
			fn(entry.data)
		}
		return true
	}

	return false
}

// Middleware adds structured, context-based logging to an HTTP handler. All
// requests will include a log entry in their context of the requested type.
func Middleware[T any](create EntryMaker[T], opts ...MiddlewareOption) func(http.Handler) http.Handler {
	logger := NewLogger(create)
	return logger.Middleware(opts...)
}
