package logs

import (
	"context"
	"net/http"
)

// Logger is a logger that logs structured data.
type Logger[T any] struct {
	create EntryMaker[T]
}

// NewLogger creates a new structured logger for the logs of the specified
// type.
func NewLogger[T any](create EntryMaker[T]) Logger[T] {
	return Logger[T]{create}
}

// AddEntry adds a log entry to the context.
func (s Logger[T]) AddEntry(ctx context.Context, opts ...Option) context.Context {
	return addEntry(ctx, s.create, opts...)
}

// Adjust mutates the log entry in the context as JSON. The function will return
// false if no log entry of the correct type is found in the context.
func (s Logger[T]) Adjust(ctx context.Context, fns ...Adjuster[T]) bool {
	return Adjust(ctx, fns...)
}

// Print prints the log entry in the context as JSON. The function will return
// false if no log entry of the correct type is found in the context.
func (s Logger[T]) Print(ctx context.Context, opts ...PrintOption) bool {
	return print[T](ctx, opts...)
}

// Retrieve gets the log entry from the context for direct manipulation. The
// function will return nil if no log entry of the correct type is found in the
// context.
func (s Logger[T]) Retrieve(ctx context.Context) *T {
	return GetEntry[T](ctx)
}

// Debug sets the log entry's level to DEBUG and adds data to it. The function
// will return false if no log entry of the correct type is found in the
// context.
func (s Logger[T]) Debug(ctx context.Context) bool {
	return debug[T](ctx)
}

// Info sets the log entry's level to INFO and adds data to it. The function
// will return false if no log entry of the correct type is found in the
// context.
func (s Logger[T]) Info(ctx context.Context) bool {
	return info[T](ctx)
}

// Warn sets the log entry's level to WARN and adds data to it. The function
// will return false if no log entry of the correct type is found in the
// context.
func (s Logger[T]) Warn(ctx context.Context) bool {
	return warn[T](ctx)
}

// Error sets the log entry's level to ERROR and adds data to it. The function
// will return false if no log entry of the correct type is found in the
// context.
func (s Logger[T]) Error(ctx context.Context) bool {
	return err[T](ctx)
}

// Fatal sets the log entry's level to FATAL and adds data to it. The function
// will return false if no log entry of the correct type is found in the
// context.
func (s Logger[T]) Fatal(ctx context.Context) bool {
	return fatal[T](ctx)
}

func (s Logger[T]) Middleware(opts ...MiddlewareOption) func(http.Handler) http.Handler {
	return middleware(s.create, opts...)
}

// GetEntry gets the log entry from the context for direct manipulation. The
// function will return nil if no log entry of the correct type is found.
func GetEntry[T any](ctx context.Context) *T {
	if entry := find[T](ctx); entry != nil {
		return entry.data
	}

	return nil
}

// Adjuster is a function that adjust a log entry.
type Adjuster[T any] func(*T)

// Adjust mutates the log entry in the context. The function will return false
// if no log entry of the correct type is found in the context.
func Adjust[T any](ctx context.Context, fns ...Adjuster[T]) bool {
	if entry := GetEntry[T](ctx); entry != nil {
		for _, fn := range fns {
			fn(entry)
		}
		return true
	}

	return false
}
