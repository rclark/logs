//go:build !structuredlogs
// +build !structuredlogs

/*
Package logs offers structured, context-based logging via two methods. The
freeform method lets you create log entries by adding key-value pairs to a
[FreeformEntry], which represents a log entry as a map with string keys and any
values.

If you want more control or enforcement over the structure of your log entries,
you can define a custom type for your logs. The [Logger]'s methods uses this
type for log entry manipulation. This will require you to pass the [Logger]
around in your application, but the [Logger.Set] and [Get] functions can help
make that easier.

Alternatively, build your application with the "structuredlogs" tag. This
exposes package-level generic functions, similar to the freeform method, but
using your custom type.
*/
package logs

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

// FreeformEntry is a freeform log entry.
type FreeformEntry map[string]any

func newFreeformEntry() *FreeformEntry { return &FreeformEntry{} }

// AddEntry adds a log entry to the context.
func AddEntry(ctx context.Context, opts ...Option) context.Context {
	return addEntry(ctx, newFreeformEntry, opts...)
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
func Print(ctx context.Context, opts ...PrintOption) bool {
	return print[FreeformEntry](ctx, opts...)
}

// Debug sets the log entry's level to DEBUG. The function will return false if
// no log entry is found in the context.
func Debug(ctx context.Context) bool {
	return debug[FreeformEntry](ctx)
}

// Info sets the log entry's level to INFO. The function will return false if no
// log entry is found in the context.
func Info(ctx context.Context) bool {
	return info[FreeformEntry](ctx)
}

// Warn sets the log entry's level to WARN. The function will return false if no
// log entry is found in the context.
func Warn(ctx context.Context) bool {
	return warn[FreeformEntry](ctx)
}

// Error sets the log entry's level to ERROR. The function will return false if
// no log entry is found in the context.
func Error(ctx context.Context) bool {
	return err[FreeformEntry](ctx)
}

// Fatal sets the log entry's level to FATAL. The function will return false if
// no log entry is found in the context.
func Fatal(ctx context.Context) bool {
	return fatal[FreeformEntry](ctx)
}

type keyValue struct {
	Key   string
	Value any
}

func (kv keyValue) adjust(e FreeformEntry, adj func(map[string]any, string)) {
	split := strings.Split(kv.Key, ".")

	current := e
	for i, sub := range split {
		if i == len(split)-1 {
			adj(current, sub)
			break
		}

		if _, ok := current[sub]; !ok {
			current[sub] = FreeformEntry{}
		}

		if nested, ok := current[sub].(map[string]any); ok {
			current = nested
		} else {
			m := make(map[string]any)
			current[sub] = m
			current = m
		}
	}
}

type keyValues []keyValue

func (k keyValues) adjust(e FreeformEntry) {
	for _, kv := range k {
		kv.adjust(e, func(m map[string]any, k string) {
			m[k] = kv.Value
		})
	}
}

func toKeyValues(args ...any) keyValues {
	kvs := make([]keyValue, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		if key, ok := args[i].(string); ok {
			kvs[i/2] = keyValue{Key: key, Value: args[i+1]}
		}
	}
	return kvs
}

// GetFreeformEntry retrieves the freeform log entry from the context. The
// function will return nil if no freeform log entry is found in the context.
func GetEntry(ctx context.Context) *FreeformEntry {
	if e := getEntry[FreeformEntry](ctx); e != nil {
		return e.data
	}

	return nil
}

// Adjust mutates the log entry in the context. The function will return false
// if no log entry of the correct type is found in the context.
func Adjust(ctx context.Context, fns ...func(*FreeformEntry)) bool {
	if entry := getEntry[FreeformEntry](ctx); entry != nil {
		for _, fn := range fns {
			fn(entry.data)
		}
		return true
	}

	return false
}

// Add adds key-value pairs to a freeform log entry. The function will return
// false if no freeform log entry is found in the context.
func Add(ctx context.Context, args ...any) bool {
	if e := GetEntry(ctx); e != nil {
		toKeyValues(args...).adjust(*e)
		return true
	}

	return false
}

// Append adds values to an existing key of the freeform log entry in the
// context. If the key does not exist, it will be created. The function will
// return false if no freeform log entry is found in the context, or if the key
// exists but its value is not []T.
func Append[T any](ctx context.Context, key string, values ...T) bool {
	adjusted := false

	adjust(ctx, func(e *FreeformEntry) {
		kv := keyValue{Key: key, Value: values}

		kv.adjust(*e, func(m map[string]any, k string) {
			if _, exists := m[k]; !exists {
				m[k] = values
				adjusted = true
			} else if existing, ok := m[k].([]T); ok {
				m[k] = append(existing, values...)
				adjusted = true
			}
		})
	})

	return adjusted
}

// WithBody configures the middleware to write request bodies into each log
// entry. This option will have no effect unless [Middleware] is operating
// on a [FreeformEntry].
func WithBody() MiddlewareOption {
	return func(o *option) {
		o.body = true
	}
}

// WithAllHeaders configures the middleware to write all request headers into
// each log entry. This option will have no effect unless [Middleware] is
// operating on a [FreeformEntry].
func WithAllHeaders() MiddlewareOption {
	return func(o *option) {
		o.allHeaders = true
	}
}

// WithHeaders configures the middleware to write specific request headers into
// each log entry. This option will have no effect unless [Middleware] is
// operating on a [FreeformEntry].
func WithHeaders(headers ...string) MiddlewareOption {
	return func(o *option) {
		o.someHeaders = headers
	}
}

// HttpData is the data structure for HTTP data that the middleware will apply
// to log entries under the `@http` key of a [FreeformEntry].
type HttpData struct {
	Method   string            `json:"method"`
	Path     string            `json:"path"`
	Headers  map[string]string `json:"headers,omitempty"`
	Body     string            `json:"body,omitempty"`
	Duration time.Duration     `json:"duration"`
}

type bodyWatcher struct {
	io.ReadCloser
	buf *bytes.Buffer
}

func (bw *bodyWatcher) Read(p []byte) (int, error) {
	n, err := bw.ReadCloser.Read(p)
	if n > 0 {
		bw.buf.Write(p[:n])
	}
	return n, err
}

// Middleware adds structured, context-based logging to an HTTP handler.
func Middleware(opts ...MiddlewareOption) func(http.Handler) http.Handler {
	opt := applyOptions(opts...)

	var options = func(o *option) {
		*o = opt
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := opt.timer.Now()

			ctx := AddEntry(r.Context(), options)
			data := HttpData{Method: r.Method, Path: r.URL.Path}

			var buf *bytes.Buffer
			if opt.body {
				buf = new(bytes.Buffer)
				r.Body = &bodyWatcher{r.Body, buf}
			}

			next.ServeHTTP(w, r.WithContext(ctx))

			data.Duration = opt.timer.Since(start)
			if opt.body {
				data.Body = buf.String()
			}

			if len(opt.someHeaders) > 0 {
				data.Headers = make(map[string]string)
				for _, h := range opt.someHeaders {
					data.Headers[h] = r.Header.Get(h)
				}
			} else if opt.allHeaders {
				data.Headers = make(map[string]string)
				for k := range r.Header {
					data.Headers[k] = r.Header.Get(k)
				}
			}

			Add(ctx, "@http", data)
			Print(ctx, options)
		})
	}
}
