package logs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// EntryMaker is any function that creates a new, mutable log entry.
type EntryMaker[T any] func() *T

// Level represents the level of logging.
type Level int

const (
	// DEBUG is the lowest level of logging for your most verbose information.
	DEBUG Level = iota
	// INFO is the default logging level for general information.
	INFO
	// WARN is for logging more important information, but not critical.
	WARN
	// ERROR is for logging failures.
	ERROR
	// FATAL is for logging failures that are likely to crash your application.
	FATAL
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

type option struct {
	out         io.Writer
	entryLevel  Level
	printLevel  Level
	timer       Timer
	body        bool
	allHeaders  bool
	someHeaders []string
	now         time.Time
	since       time.Duration
	fakeTime    bool
}

// PrintOption is a configuration option for printing logs.
type PrintOption func(*option)

// WithOutput sets the output for the log entry. The default is os.Stdout.
func WithOutput(out io.Writer) PrintOption {
	return func(o *option) {
		o.out = out
	}
}

// WithLevel sets the log level for printing the log entry. The default is
// INFO. If the log entry's level is less than the level set here, it will not
// be printed.
func WithLevel(level Level) PrintOption {
	return func(o *option) {
		o.printLevel = level
	}
}

// Timer is an interface for measuring HTTP request duration. Provide your own
// implementation to use as a custom timer if you want to test your logging
// system.
type Timer interface {
	Now() time.Time
	Since(time.Time) time.Duration
}

type defaultTimer struct{}

func (defaultTimer) Now() time.Time {
	return time.Now()
}

func (defaultTimer) Since(t time.Time) time.Duration {
	return time.Since(t)
}

type fakeTimer struct {
	now   time.Time
	since time.Duration
}

func (t fakeTimer) Now() time.Time {
	return t.now
}

func (t fakeTimer) Since(time.Time) time.Duration {
	return t.since
}

// WithCurrentTime configures logs to always print with the same timestamp.
func WithCurrentTime(now time.Time) PrintOption {
	return func(o *option) {
		o.now = now
		o.fakeTime = true
	}
}

// Option is configuration for a log entry.
type Option func(*option)

// WithDefaultLevel sets the log level for the log entry. This can be
// overridden while collecting log data using functions like [Debug] and
// [Error]. The default level for an entry if this configuration option is not
// applied is INFO.
func WithDefaultLevel(level Level) Option {
	return func(o *option) {
		o.entryLevel = level
	}
}

func applyOptions[T ~func(*option)](opts ...T) option {
	o := option{
		out:        os.Stdout,
		entryLevel: INFO,
		printLevel: INFO,
		timer:      defaultTimer{},
	}

	for _, opt := range opts {
		opt(&o)
	}

	if o.fakeTime {
		o.timer = fakeTimer{
			now:   o.now,
			since: o.since,
		}
	}

	return o
}

type entryKey struct{}

var eKey = entryKey{}

type entry[T any] struct {
	level Level
	data  *T
}

func addEntry[T any](ctx context.Context, create EntryMaker[T], opts ...Option) context.Context {
	log := entry[T]{data: create()}
	options := applyOptions(opts...)
	log.level = options.entryLevel
	return context.WithValue(ctx, eKey, &log)
}

func getEntry[T any](ctx context.Context) *entry[T] {
	if entry, ok := ctx.Value(eKey).(*entry[T]); ok {
		return entry
	}

	return nil
}

// adjuster is a function that adjust a log entry.
type adjuster[T any] func(*T)

// adjust mutates the log entry in the context. The function will return false
// if no log entry of the correct type is found in the context.
func adjust[T any](ctx context.Context, fns ...adjuster[T]) bool {
	if entry := getEntry[T](ctx); entry != nil {
		for _, fn := range fns {
			fn(entry.data)
		}
		return true
	}

	return false
}

func print[T any](ctx context.Context, opts ...PrintOption) bool {
	if entry := getEntry[T](ctx); entry != nil {
		options := applyOptions(opts...)

		if entry.level < options.printLevel {
			return false
		}

		data, err := json.Marshal(entry.data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal log entry to JSON: %v\n", err)
			return false
		}

		if bytes.Index(data, []byte("{")) == 0 {
			tpl := `{"@level":"%s","@time":"%s",`
			if bytes.Index(data, []byte("}")) == 1 {
				tpl = `{"@level":"%s","@time":"%s"`
			}
			now := options.timer.Now().Format(time.RFC3339)
			meta := []byte(fmt.Sprintf(tpl, entry.level, now))
			data = append(meta, data[1:]...)
			data = append(data, '\n')
		}

		if _, err := options.out.Write(data); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write log entry: %v\n", err)
			return false
		}

		return true
	}

	return false
}

func debug[T any](ctx context.Context) bool {
	if entry := getEntry[T](ctx); entry != nil {
		entry.level = DEBUG
		return true
	}

	return false
}

func info[T any](ctx context.Context) bool {
	if entry := getEntry[T](ctx); entry != nil {
		entry.level = INFO
		return true
	}

	return false
}

func warn[T any](ctx context.Context) bool {
	if entry := getEntry[T](ctx); entry != nil {
		entry.level = WARN
		return true
	}

	return false
}

func err[T any](ctx context.Context) bool {
	if entry := getEntry[T](ctx); entry != nil {
		entry.level = ERROR
		return true
	}

	return false
}

func fatal[T any](ctx context.Context) bool {
	if entry := getEntry[T](ctx); entry != nil {
		entry.level = FATAL
		return true
	}

	return false
}

// MiddlewareOption is used to configure logs generated by the [Middleware].
type MiddlewareOption func(*option)

// DefaultLevel sets the default log level for the log entries produced by the
// [Middleware].
func DefaultLevel(level Level) MiddlewareOption {
	return MiddlewareOption(WithDefaultLevel(level))
}

// PrintLevel sets the minimum log level for printing log entries produced by
// the [Middleware].
func PrintLevel(level Level) MiddlewareOption {
	return MiddlewareOption(WithLevel(level))
}

// WithOutput sets the output for the log entries produced by the [Middleware].
func Output(out io.Writer) MiddlewareOption {
	return MiddlewareOption(WithOutput(out))
}

// WithTiming configures the middleware to always print logs with the given
// timestamp and the given duration.
func WithTiming(now time.Time, since time.Duration) MiddlewareOption {
	return func(o *option) {
		o.now = now
		o.since = since
		o.fakeTime = true
	}
}
