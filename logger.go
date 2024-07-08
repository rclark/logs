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
func (logger Logger[T]) AddEntry(ctx context.Context, opts ...Option) context.Context {
	return addEntry(ctx, logger.create, opts...)
}

// Adjust mutates the log entry in the context as JSON. The function will return
// false if no log entry of the correct type is found in the context.
func (Logger[T]) Adjust(ctx context.Context, fns ...adjuster[T]) bool {
	return adjust(ctx, fns...)
}

// Print prints the log entry in the context as JSON. The function will return
// false if no log entry of the correct type is found in the context.
func (Logger[T]) Print(ctx context.Context, opts ...PrintOption) bool {
	return print[T](ctx, opts...)
}

// GetEntry gets the log entry from the context for direct manipulation. The
// function will return nil if no log entry of the correct type is found in the
// context.
func (Logger[T]) GetEntry(ctx context.Context) *T {
	if e := getEntry[T](ctx); e != nil {
		return e.data
	}

	return nil
}

// Debug sets the log entry's level to DEBUG and adds data to it. The function
// will return false if no log entry of the correct type is found in the
// context.
func (Logger[T]) Debug(ctx context.Context) bool {
	return debug[T](ctx)
}

// Info sets the log entry's level to INFO and adds data to it. The function
// will return false if no log entry of the correct type is found in the
// context.
func (Logger[T]) Info(ctx context.Context) bool {
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
func (Logger[T]) Error(ctx context.Context) bool {
	return err[T](ctx)
}

// Fatal sets the log entry's level to FATAL and adds data to it. The function
// will return false if no log entry of the correct type is found in the
// context.
func (Logger[T]) Fatal(ctx context.Context) bool {
	return fatal[T](ctx)
}

// Middleware adds structured, context-based logging to an HTTP handler. All
// requests will include a log entry in their context of the requested type.
func (logger Logger[T]) Middleware(opts ...MiddlewareOption) func(http.Handler) http.Handler {
	opt := applyOptions(opts...)

	var options = func(o *option) {
		*o = opt
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := logger.Set(r.Context())
			ctx = logger.AddEntry(ctx, options)
			next.ServeHTTP(w, r.WithContext(ctx))
			logger.Print(ctx, options)
		})
	}
}

type loggerKey struct{}

var lKey = loggerKey{}

// Set places the logger in the context.
func (logger Logger[T]) Set(ctx context.Context) context.Context {
	return context.WithValue(ctx, lKey, logger)
}

// Get retrieves the logger from the context. The function will return nil if no
// logger of the requested type is found.
func Get[T any](ctx context.Context) *Logger[T] {
	if v := ctx.Value(lKey); v != nil {
		if logger, ok := v.(Logger[T]); ok {
			return &logger
		}
	}

	return nil
}

// ExampleLog is an example of a struct designed to be used as a log entry.
type ExampleLog struct {
	Name     string   `json:"name"`
	Count    int      `json:"count"`
	Flag     bool     `json:"flag"`
	Messages []string `json:"messages,omitempty"`
}

// NewExampleLog defines how to create an empty, mutable version of an
// [ExampleLog].
func NewExampleLog() *ExampleLog {
	return &ExampleLog{}
}
