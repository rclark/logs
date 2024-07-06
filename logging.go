/*
Package logs supports structured, context-based logging.

When you import the package, it will start in [FreeformMode]. This means that
you'll be working to generate logs that are represented internally as
[FreeformEntry]s. In [FreeformMode]:

  - [Add] allows you to add arbitrary, nested, key-value pairs to the log entry.
  - [Append] allows to you append items to a slice that's located at some nested
    key within the log entry.
  - [GetFreeformEntry] allows you to get the log entry and manipulate it directly.

You can switch the package into [StructuredMode]. In this mode, you must design
your own struct that represents your desired log entry, and provide a
function for creating a mutable representation of that struct (i.e. a
pointer). In [StructuredMode]:

  - [Adjust] allows you to provide functions that mutate the underlying log entry.
  - [GetEntry] allows you to get a pointer to the log entry and manipulate it
    directly.

In either mode, you will use a number of functions which are exported as
variables. These variables will change their behavior depending on the mode,
to accommodate the different types of log entries. These variables include:

  - [AddEntry] to place a new, empty log entry in some context.
  - [Print] to write the log entry found in a context, as JSON, to
    os.Stdout (or your io.Writer of choice).
  - [Debug], [Info], etc. functions to set the log level.

Output logs in [FreeformMode] will always include "@level" and "@time"
properties to describe the log's level and the time that [Print] was called. In
[StructuredMode] these properties will be added as long as your provided type
marshals to JSON as an object.
*/
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

type mode int

const (
	// Freeform is the default logging mode. It allows you to log key-value pairs
	// using the [Add] and [Append] functions. You can manipulate the
	// [FreeformEntry] (which is a map[string]any) using the [AdjustFreeform]
	// function, or retrieve it using the [GetFreeformEntry] function.
	Freeform mode = iota

	// Structured is a logging mode that allows you to log data structured
	// according to a specific type of your design. You can manipulate the log
	// entry using the [Adjust] function or retrieve it using the [GetEntry]
	// function.
	Structured
)

var m mode

// CurrentMode returns the current logging mode.
func CurrentMode() mode {
	return m
}

// StructuredMode sets the logging mode to structured for a specific type of
// log entry.
//
// This mode allows you to log data structured according to a specific type of
// your design. You can manipulate the log entry using the [Adjust] function
// or retrieve it using the [GetEntry] function.
func StructuredMode[T any](create EntryMaker[T]) {
	m = Structured
	logger := NewLogger(create)
	AddEntry = logger.AddEntry
	Print = logger.Print
	Debug = logger.Debug
	Info = logger.Info
	Warn = logger.Warn
	Error = logger.Error
	Fatal = logger.Fatal
	Middleware = logger.Middleware
}

// FreeformMode sets the logging mode to free-form.
//
// This is the default logging mode. It allows you to log key-value pairs
// using the [Add] and [Append] functions. You can manipulate the
// [FreeformEntry] (which is a map[string]any) using the [AdjustFreeform]
// function, or retrieve it using the [GetFreeformEntry] function.
func FreeformMode() {
	m = Freeform
	AddEntry = addFreeformEntry
	Print = print[FreeformEntry]
	Debug = debug[FreeformEntry]
	Info = info[FreeformEntry]
	Warn = warn[FreeformEntry]
	Error = err[FreeformEntry]
	Fatal = fatal[FreeformEntry]
	Middleware = freeformMiddleware
}

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

// WithCurrentTime configures a custom timer for measuring the current time.
func WithCurrentTime(t Timer) PrintOption {
	return func(o *option) {
		o.timer = t
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

	return o
}

type ctxKey struct{}

var entryKey = ctxKey{}

type entry[T any] struct {
	level Level
	data  *T
}

// AddEntry adds a log entry to the context. This variable's value changes when
// you change the package to [StructuredMode].
var AddEntry func(ctx context.Context, opts ...Option) context.Context = addFreeformEntry

func addEntry[T any](ctx context.Context, create EntryMaker[T], opts ...Option) context.Context {
	log := entry[T]{data: create()}
	options := applyOptions(opts...)
	log.level = options.entryLevel
	return context.WithValue(ctx, entryKey, &log)
}

func addFreeformEntry(ctx context.Context, opts ...Option) context.Context {
	return addEntry(ctx, newFreeformEntry, opts...)
}

func find[T any](ctx context.Context) *entry[T] {
	if entry, ok := ctx.Value(entryKey).(*entry[T]); ok {
		return entry
	}

	return nil
}

// Print prints the log entry in the context as JSON. The function will return
// false if no log entry is found. This variable's value changes when you
// change the package to [StructuredMode].
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
var Print func(ctx context.Context, opts ...PrintOption) bool = print[FreeformEntry]

func print[T any](ctx context.Context, opts ...PrintOption) bool {
	if entry := find[T](ctx); entry != nil {
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
		}

		if _, err := options.out.Write(data); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write log entry: %v\n", err)
			return false
		}

		return true
	}

	return false
}

// Debug sets the log entry's level to DEBUG. The function will return false if
// no log entry is found in the context. This variable's value changes when you
// change the package to [StructuredMode].
var Debug func(ctx context.Context) bool = debug[FreeformEntry]

func debug[T any](ctx context.Context) bool {
	if entry := find[T](ctx); entry != nil {
		entry.level = DEBUG
		return true
	}

	return false
}

// Info sets the log entry's level to INFO. The function will return false if no
// log entry is found in the context. This variable's value changes when you
// change the package to [StructuredMode].
var Info func(ctx context.Context) bool = info[FreeformEntry]

func info[T any](ctx context.Context) bool {
	if entry := find[T](ctx); entry != nil {
		entry.level = INFO
		return true
	}

	return false
}

// Warn sets the log entry's level to WARN. The function will return false if no
// log entry is found in the context. This variable's value changes when you
// change the package to [StructuredMode].
var Warn func(ctx context.Context) bool = warn[FreeformEntry]

func warn[T any](ctx context.Context) bool {
	if entry := find[T](ctx); entry != nil {
		entry.level = WARN
		return true
	}

	return false
}

// Error sets the log entry's level to ERROR. The function will return false if
// no log entry is found in the context. This variable's value changes when you
// change the package to [StructuredMode].
var Error func(ctx context.Context) bool = err[FreeformEntry]

func err[T any](ctx context.Context) bool {
	if entry := find[T](ctx); entry != nil {
		entry.level = ERROR
		return true
	}

	return false
}

// Fatal sets the log entry's level to FATAL. The function will return false if
// no log entry is found in the context. This variable's value changes when you
// change the package to [StructuredMode].
var Fatal func(ctx context.Context) bool = fatal[FreeformEntry]

func fatal[T any](ctx context.Context) bool {
	if entry := find[T](ctx); entry != nil {
		entry.level = FATAL
		return true
	}

	return false
}
