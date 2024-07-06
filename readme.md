[![Go](https://github.com/rclark/logs/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/rclark/logs/actions/workflows/go.yml)

# logs

A package to support context-based logging.

In most situations where you find yourself reading logs, you're trying to understand what happened during some "unit of work" that your application performed. For example, you might be troubleshooting a failure to respond to an HTTP request the way that you expected the application would respond. In that case, the HTTP request represents a "unit of work" in your application.

Often, applications will generate some unique identifier for each unit of work it performs. The application will make sure (using a package like `slog`) that every time a log is printed related to a unit of work, it includes the unique identifier. When troubleshooting, the process of reading logs then goes something like this:

1. Find _any log_ from the unit of work you're troubleshooting.
2. Get the unit of work's unique identifier.
3. Find all the other logs that share the same unique identifier.

This package instead encourages you to put all the log information from any single unit of work into a single log entry, printed as a JSON object. This makes all the data related to a unit of work easier to find, and it makes it easier to extract quantitative values (i.e. metrics) from your logs. Several commonly-used observability platforms give you tools to perform this kind of quantitative log analysis.

It is common that in a Go application, one unit of work will tend to share a single `context.Context` object. This package attaches a log entry to that context, and gives your application tools to manipulate the log entry as it performs each unit of work.

This approach to logging is more expensive, performance-wise, than standard logging. And it can take a little getting-used to for an application developer. Instead of thinking about log "messages", you begin thinking about log "data" that describes the actions your application performed not as prose, but as more primitive values like counts, flags, durations, etc. On the other end of that performance/learning tradeoff are massive improvements to a system's observability improvements.

## Usage

```go
import "github.com/rclark/logs"
```

Package logs supports structured, context\-based logging.

When you import the package, it will start in [FreeformMode](<#FreeformMode>). This means that you'll be working to generate logs that are represented internally as \[FreeformEntry\]s. In [FreeformMode](<#FreeformMode>):

- [Add](<#Add>) allows you to add arbitrary, nested, key\-value pairs to the log entry.
- [Append](<#Append>) allows to you append items to a slice that's located at some nested key within the log entry.
- [GetFreeformEntry](<#GetFreeformEntry>) allows you to get the log entry and manipulate it directly.

You can switch the package into [StructuredMode](<#StructuredMode>). In this mode, you must design your own struct that represents your desired log entry, and provide a function for creating a mutable representation of that struct \(i.e. a pointer\). In [StructuredMode](<#StructuredMode>):

- [Adjust](<#Adjust>) allows you to provide functions that mutate the underlying log entry.
- [GetEntry](<#GetEntry>) allows you to get a pointer to the log entry and manipulate it directly.

In either mode, you will use a number of functions which are exported as variables. These variables will change their behavior depending on the mode, to accommodate the different types of log entries. These variables include:

- [AddEntry](<#AddEntry>) to place a new, empty log entry in some context.
- [Print](<#Print>) to write the log entry found in a context, as JSON, to os.Stdout \(or your io.Writer of choice\).
- [Debug](<#Debug>), [Info](<#Info>), etc. functions to set the log level.

Output logs in [FreeformMode](<#FreeformMode>) will always include "@level" and "@time" properties to describe the log's level and the time that [Print](<#Print>) was called. In [StructuredMode](<#StructuredMode>) these properties will be added as long as your provided type marshals to JSON as an object.

<details><summary>Example (Freeform)</summary>
<p>



```go
package main

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

func main() {
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
}
```

#### Output

```
{"@level":"INFO","@time":"0001-01-01T00:00:00Z","count":42,"flag":true,"messages":["hello","world"],"name":"test"}
```

</p>
</details>

<details><summary>Example (Structured)</summary>
<p>



```go
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
```

#### Output

```
{"@level":"INFO","@time":"0001-01-01T00:00:00Z","Name":"test","Count":42,"Flag":true,"Messages":["hello","world"]}
```

</p>
</details>

<details><summary>Example (Structured Logger)</summary>
<p>



```go
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
```

#### Output

```
{"@level":"INFO","@time":"0001-01-01T00:00:00Z","Name":"test","Count":42,"Flag":false,"Messages":["hello","world"]}
```

</p>
</details>

## Index

- [Variables](<#variables>)
- [func Add\(ctx context.Context, args ...any\) bool](<#Add>)
- [func Adjust\[T any\]\(ctx context.Context, fns ...Adjuster\[T\]\) bool](<#Adjust>)
- [func AdjustFreeform\(ctx context.Context, fns ...func\(\*FreeformEntry\)\) bool](<#AdjustFreeform>)
- [func Append\[T any\]\(ctx context.Context, key string, values ...T\) bool](<#Append>)
- [func FreeformMode\(\)](<#FreeformMode>)
- [func GetEntry\[T any\]\(ctx context.Context\) \*T](<#GetEntry>)
- [func StructuredMode\[T any\]\(create EntryMaker\[T\]\)](<#StructuredMode>)
- [type Adjuster](<#Adjuster>)
- [type EntryMaker](<#EntryMaker>)
- [type FreeformEntry](<#FreeformEntry>)
  - [func GetFreeformEntry\(ctx context.Context\) \*FreeformEntry](<#GetFreeformEntry>)
- [type HttpData](<#HttpData>)
- [type Level](<#Level>)
  - [func \(l Level\) String\(\) string](<#Level.String>)
- [type Logger](<#Logger>)
  - [func NewLogger\[T any\]\(create EntryMaker\[T\]\) Logger\[T\]](<#NewLogger>)
  - [func \(s Logger\[T\]\) AddEntry\(ctx context.Context, opts ...Option\) context.Context](<#Logger[T].AddEntry>)
  - [func \(s Logger\[T\]\) Adjust\(ctx context.Context, fns ...Adjuster\[T\]\) bool](<#Logger[T].Adjust>)
  - [func \(s Logger\[T\]\) Debug\(ctx context.Context\) bool](<#Logger[T].Debug>)
  - [func \(s Logger\[T\]\) Error\(ctx context.Context\) bool](<#Logger[T].Error>)
  - [func \(s Logger\[T\]\) Fatal\(ctx context.Context\) bool](<#Logger[T].Fatal>)
  - [func \(s Logger\[T\]\) Info\(ctx context.Context\) bool](<#Logger[T].Info>)
  - [func \(s Logger\[T\]\) Middleware\(opts ...MiddlewareOption\) func\(http.Handler\) http.Handler](<#Logger[T].Middleware>)
  - [func \(s Logger\[T\]\) Print\(ctx context.Context, opts ...PrintOption\) bool](<#Logger[T].Print>)
  - [func \(s Logger\[T\]\) Retrieve\(ctx context.Context\) \*T](<#Logger[T].Retrieve>)
  - [func \(s Logger\[T\]\) Warn\(ctx context.Context\) bool](<#Logger[T].Warn>)
- [type MiddlewareOption](<#MiddlewareOption>)
  - [func DefaultLevel\(level Level\) MiddlewareOption](<#DefaultLevel>)
  - [func Output\(out io.Writer\) MiddlewareOption](<#Output>)
  - [func PrintLevel\(level Level\) MiddlewareOption](<#PrintLevel>)
  - [func WithAllHeaders\(\) MiddlewareOption](<#WithAllHeaders>)
  - [func WithBody\(\) MiddlewareOption](<#WithBody>)
  - [func WithHeaders\(headers ...string\) MiddlewareOption](<#WithHeaders>)
  - [func WithTimer\(t Timer\) MiddlewareOption](<#WithTimer>)
- [type Option](<#Option>)
  - [func WithDefaultLevel\(level Level\) Option](<#WithDefaultLevel>)
- [type PrintOption](<#PrintOption>)
  - [func WithCurrentTime\(t Timer\) PrintOption](<#WithCurrentTime>)
  - [func WithLevel\(level Level\) PrintOption](<#WithLevel>)
  - [func WithOutput\(out io.Writer\) PrintOption](<#WithOutput>)
- [type Timer](<#Timer>)


## Variables

<a name="AddEntry"></a>AddEntry adds a log entry to the context. This variable's value changes when you change the package to [StructuredMode](<#StructuredMode>).

```go
var AddEntry func(ctx context.Context, opts ...Option) context.Context = addFreeformEntry
```

<a name="Debug"></a>Debug sets the log entry's level to DEBUG. The function will return false if no log entry is found in the context. This variable's value changes when you change the package to [StructuredMode](<#StructuredMode>).

```go
var Debug func(ctx context.Context) bool = debug[FreeformEntry]
```

<a name="Error"></a>Error sets the log entry's level to ERROR. The function will return false if no log entry is found in the context. This variable's value changes when you change the package to [StructuredMode](<#StructuredMode>).

```go
var Error func(ctx context.Context) bool = err[FreeformEntry]
```

<a name="Fatal"></a>Fatal sets the log entry's level to FATAL. The function will return false if no log entry is found in the context. This variable's value changes when you change the package to [StructuredMode](<#StructuredMode>).

```go
var Fatal func(ctx context.Context) bool = fatal[FreeformEntry]
```

<a name="Info"></a>Info sets the log entry's level to INFO. The function will return false if no log entry is found in the context. This variable's value changes when you change the package to [StructuredMode](<#StructuredMode>).

```go
var Info func(ctx context.Context) bool = info[FreeformEntry]
```

<a name="Middleware"></a>Middleware adds structured, context\-based logging to an HTTP handler. This variable's value changes when you change the package to [StructuredMode](<#StructuredMode>).

```go
var Middleware func(opts ...MiddlewareOption) func(http.Handler) http.Handler = freeformMiddleware
```

<a name="Print"></a>Print prints the log entry in the context as JSON. The function will return false if no log entry is found. This variable's value changes when you change the package to [StructuredMode](<#StructuredMode>).

The default output is os.Stdout. You can change this by providing a custom io.Writer using the [WithOutput](<#WithOutput>) option.

The default log level for printing is INFO. You can change this by providing a custom log level using the [WithLevel](<#WithLevel>) option.

The default timer is the system clock. You can change this by providing a custom timer using the [WithCurrentTime](<#WithCurrentTime>) option. This is useful if you need to write tests to confirm that your application is logging as expected, as your custom timer can be used to control the log entry's "@time" property.

If the you've provided a custom struct for your log entries and it fails to marshal to JSON using the standard json.Marshal\(\), the function will write an error message to os.Stderr and return false.

```go
var Print func(ctx context.Context, opts ...PrintOption) bool = print[FreeformEntry]
```

<a name="Warn"></a>Warn sets the log entry's level to WARN. The function will return false if no log entry is found in the context. This variable's value changes when you change the package to [StructuredMode](<#StructuredMode>).

```go
var Warn func(ctx context.Context) bool = warn[FreeformEntry]
```

<a name="Add"></a>
## func Add

```go
func Add(ctx context.Context, args ...any) bool
```

Add adds key\-value pairs to a freeform log entry. The function will return false if no freeform log entry is found in the context.

<details><summary>Example (Nested Keys)</summary>
<p>



```go
package main

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

func main() {
	// Set the logging mode to freeform. This is the default.
	logs.FreeformMode()

	ctx := logs.AddEntry(context.Background())

	logs.Add(ctx,
		"user.name", "test",
		"user.count", 42,
		"user.flags.flag", true,
	)

	logs.Print(ctx, logs.WithCurrentTime(fakeTimer{}))
}
```

#### Output

```
{"@level":"INFO","@time":"0001-01-01T00:00:00Z","user":{"count":42,"flags":{"flag":true},"name":"test"}}
```

</p>
</details>

<a name="Adjust"></a>
## func Adjust

```go
func Adjust[T any](ctx context.Context, fns ...Adjuster[T]) bool
```

Adjust mutates the log entry in the context. The function will return false if no log entry of the correct type is found in the context.

<a name="AdjustFreeform"></a>
## func AdjustFreeform

```go
func AdjustFreeform(ctx context.Context, fns ...func(*FreeformEntry)) bool
```

AdjustFreeform adjusts the freeform log entry in the context. The function will return false if no freeform log entry is found in the context.

<a name="Append"></a>
## func Append

```go
func Append[T any](ctx context.Context, key string, values ...T) bool
```

Append adds values to an existing key of the freeform log entry in the context. If the key does not exist, it will be created. The function will return false if no freeform log entry is found in the context, or if the key exists but its value is not \[\]T.

<a name="FreeformMode"></a>
## func FreeformMode

```go
func FreeformMode()
```

FreeformMode sets the logging mode to free\-form.

This is the default logging mode. It allows you to log key\-value pairs using the [Add](<#Add>) and [Append](<#Append>) functions. You can manipulate the [FreeformEntry](<#FreeformEntry>) \(which is a map\[string\]any\) using the [AdjustFreeform](<#AdjustFreeform>) function, or retrieve it using the [GetFreeformEntry](<#GetFreeformEntry>) function.

<a name="GetEntry"></a>
## func GetEntry

```go
func GetEntry[T any](ctx context.Context) *T
```

GetEntry gets the log entry from the context for direct manipulation. The function will return nil if no log entry of the correct type is found.

<a name="StructuredMode"></a>
## func StructuredMode

```go
func StructuredMode[T any](create EntryMaker[T])
```

StructuredMode sets the logging mode to structured for a specific type of log entry.

This mode allows you to log data structured according to a specific type of your design. You can manipulate the log entry using the [Adjust](<#Adjust>) function or retrieve it using the [GetEntry](<#GetEntry>) function.

<a name="Adjuster"></a>
## type Adjuster

Adjuster is a function that adjust a log entry.

```go
type Adjuster[T any] func(*T)
```

<a name="EntryMaker"></a>
## type EntryMaker

EntryMaker is any function that creates a new, mutable log entry.

```go
type EntryMaker[T any] func() *T
```

<a name="FreeformEntry"></a>
## type FreeformEntry

FreeformEntry is a freeform log entry.

```go
type FreeformEntry map[string]any
```

<a name="GetFreeformEntry"></a>
### func GetFreeformEntry

```go
func GetFreeformEntry(ctx context.Context) *FreeformEntry
```

GetFreeformEntry retrieves the freeform log entry from the context. The function will return nil if no freeform log entry is found in the context.

<a name="HttpData"></a>
## type HttpData

HttpData is the data structure for HTTP data that the middleware will apply to log entries under the \`@http\` key of a [FreeformEntry](<#FreeformEntry>).

```go
type HttpData struct {
    Method   string            `json:"method"`
    Path     string            `json:"path"`
    Headers  map[string]string `json:"headers,omitempty"`
    Body     string            `json:"body,omitempty"`
    Duration time.Duration     `json:"duration"`
}
```

<a name="Level"></a>
## type Level

Level represents the level of logging.

```go
type Level int
```

<a name="DEBUG"></a>

```go
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
```

<a name="Level.String"></a>
### func \(Level\) String

```go
func (l Level) String() string
```



<a name="Logger"></a>
## type Logger

Logger is a logger that logs structured data.

```go
type Logger[T any] struct {
    // contains filtered or unexported fields
}
```

<a name="NewLogger"></a>
### func NewLogger

```go
func NewLogger[T any](create EntryMaker[T]) Logger[T]
```

NewLogger creates a new structured logger for the logs of the specified type.

<a name="Logger[T].AddEntry"></a>
### func \(Logger\[T\]\) AddEntry

```go
func (s Logger[T]) AddEntry(ctx context.Context, opts ...Option) context.Context
```

AddEntry adds a log entry to the context.

<a name="Logger[T].Adjust"></a>
### func \(Logger\[T\]\) Adjust

```go
func (s Logger[T]) Adjust(ctx context.Context, fns ...Adjuster[T]) bool
```

Adjust mutates the log entry in the context as JSON. The function will return false if no log entry of the correct type is found in the context.

<a name="Logger[T].Debug"></a>
### func \(Logger\[T\]\) Debug

```go
func (s Logger[T]) Debug(ctx context.Context) bool
```

Debug sets the log entry's level to DEBUG and adds data to it. The function will return false if no log entry of the correct type is found in the context.

<a name="Logger[T].Error"></a>
### func \(Logger\[T\]\) Error

```go
func (s Logger[T]) Error(ctx context.Context) bool
```

Error sets the log entry's level to ERROR and adds data to it. The function will return false if no log entry of the correct type is found in the context.

<a name="Logger[T].Fatal"></a>
### func \(Logger\[T\]\) Fatal

```go
func (s Logger[T]) Fatal(ctx context.Context) bool
```

Fatal sets the log entry's level to FATAL and adds data to it. The function will return false if no log entry of the correct type is found in the context.

<a name="Logger[T].Info"></a>
### func \(Logger\[T\]\) Info

```go
func (s Logger[T]) Info(ctx context.Context) bool
```

Info sets the log entry's level to INFO and adds data to it. The function will return false if no log entry of the correct type is found in the context.

<a name="Logger[T].Middleware"></a>
### func \(Logger\[T\]\) Middleware

```go
func (s Logger[T]) Middleware(opts ...MiddlewareOption) func(http.Handler) http.Handler
```



<a name="Logger[T].Print"></a>
### func \(Logger\[T\]\) Print

```go
func (s Logger[T]) Print(ctx context.Context, opts ...PrintOption) bool
```

Print prints the log entry in the context as JSON. The function will return false if no log entry of the correct type is found in the context.

<a name="Logger[T].Retrieve"></a>
### func \(Logger\[T\]\) Retrieve

```go
func (s Logger[T]) Retrieve(ctx context.Context) *T
```

Retrieve gets the log entry from the context for direct manipulation. The function will return nil if no log entry of the correct type is found in the context.

<a name="Logger[T].Warn"></a>
### func \(Logger\[T\]\) Warn

```go
func (s Logger[T]) Warn(ctx context.Context) bool
```

Warn sets the log entry's level to WARN and adds data to it. The function will return false if no log entry of the correct type is found in the context.

<a name="MiddlewareOption"></a>
## type MiddlewareOption

MiddlewareOption is used to configure logs generated by the [Middleware](<#Middleware>).

```go
type MiddlewareOption func(*option)
```

<a name="DefaultLevel"></a>
### func DefaultLevel

```go
func DefaultLevel(level Level) MiddlewareOption
```

DefaultLevel sets the default log level for the log entries produced by the [Middleware](<#Middleware>).

<a name="Output"></a>
### func Output

```go
func Output(out io.Writer) MiddlewareOption
```

WithOutput sets the output for the log entries produced by the [Middleware](<#Middleware>).

<a name="PrintLevel"></a>
### func PrintLevel

```go
func PrintLevel(level Level) MiddlewareOption
```

PrintLevel sets the minimum log level for printing log entries produced by the [Middleware](<#Middleware>).

<a name="WithAllHeaders"></a>
### func WithAllHeaders

```go
func WithAllHeaders() MiddlewareOption
```

WithAllHeaders configures the middleware to write all request headers into each log entry. This option will have no effect unless [Middleware](<#Middleware>) is operating on a [FreeformEntry](<#FreeformEntry>).

<a name="WithBody"></a>
### func WithBody

```go
func WithBody() MiddlewareOption
```

WithBody configures the middleware to write request bodies into each log entry. This option will have no effect unless [Middleware](<#Middleware>) is operating on a [FreeformEntry](<#FreeformEntry>).

<a name="WithHeaders"></a>
### func WithHeaders

```go
func WithHeaders(headers ...string) MiddlewareOption
```

WithHeaders configures the middleware to write specific request headers into each log entry. This option will have no effect unless [Middleware](<#Middleware>) is operating on a [FreeformEntry](<#FreeformEntry>).

<a name="WithTimer"></a>
### func WithTimer

```go
func WithTimer(t Timer) MiddlewareOption
```

WithTimer configures the middleware to use a custom timer for measuring request duration and the current time.

<a name="Option"></a>
## type Option

Option is configuration for a log entry.

```go
type Option func(*option)
```

<a name="WithDefaultLevel"></a>
### func WithDefaultLevel

```go
func WithDefaultLevel(level Level) Option
```

WithDefaultLevel sets the log level for the log entry. This can be overridden while collecting log data using functions like [Debug](<#Debug>) and [Error](<#Error>). The default level for an entry if this configuration option is not applied is INFO.

<a name="PrintOption"></a>
## type PrintOption

PrintOption is a configuration option for printing logs.

```go
type PrintOption func(*option)
```

<a name="WithCurrentTime"></a>
### func WithCurrentTime

```go
func WithCurrentTime(t Timer) PrintOption
```

WithCurrentTime configures a custom timer for measuring the current time.

<a name="WithLevel"></a>
### func WithLevel

```go
func WithLevel(level Level) PrintOption
```

WithLevel sets the log level for printing the log entry. The default is INFO. If the log entry's level is less than the level set here, it will not be printed.

<a name="WithOutput"></a>
### func WithOutput

```go
func WithOutput(out io.Writer) PrintOption
```

WithOutput sets the output for the log entry. The default is os.Stdout.

<a name="Timer"></a>
## type Timer

Timer is an interface for measuring HTTP request duration. Provide your own implementation to use as a custom timer if you want to test your logging system.

```go
type Timer interface {
    Now() time.Time
    Since(time.Time) time.Duration
}
```

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
