[![Go](https://github.com/rclark/logs/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/rclark/logs/actions/workflows/go.yml)

# logs

A package to support structured, context-based logging.

## Usage

- [How to use the package in its standard configuration](./standard.md).
- [How to use the package in "structured mode"](./structured.md).

## Why use context-based logging?

When analyzing logs, the goal is often to understand a specific "unit of work" within your application, such as diagnosing why a certain HTTP request didn't behave as expected. Often, applications assign a unique identifier to each unit of work, using packages like `slog` to make sure the identifier is included in each message. The analysis requires finding any log from the unit of work and then using its identifier to locate all related logs.

This package proposes a different approach: consolidate all log data for a unit of work into a single JSON log entry. This approach streamlines data retrieval while it enables quantitative analysis and metric extraction, with broad JSON support from the most common observability platforms.

This package relies on the fact that in Go applications, a `context.Context` object often accompanies a single unit of work. The package attaches a log entry to the context and provides tools for log manipulation throughout the unit of work.

While this logging method is less performant and requires a shift in developer mindset from log "messages" to log "data" (e.g., counts, flags, durations), it significantly enhances system observability.
