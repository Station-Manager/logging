# Station-Manager: logging module

This package provides a thin, concurrency-safe wrapper around `rs/zerolog` with a structured-first API, safe lifecycle (`Initialize`/`Close`), and file rotation via `lumberjack`.

## Key features
- Structured logging only (no `Info/Infof` helpers): prefer typed fields for queryability
- Context loggers via `With()` for per-request scoping
- Safe concurrent use; graceful shutdown waits for in-flight logs with a timeout
- File rotation (size/age/backups), optional compression
- Console writer options (no-color, custom time format)

## Quick start
```go
svc := &logging.Service{AppConfig: cfgService}
if err := svc.Initialize(); err != nil { panic(err) }
defer svc.Close()

// Basic structured entry
svc.InfoWith().Str("user_id", id).Int("count", 3).Msg("processed")

// Context (scoped) logger
req := svc.With().Str("request_id", reqID).Logger()
req.DebugWith().Str("path", path).Msg("incoming request")

// Nested structures
svc.InfoWith().Dict("db", func(e logging.LogEvent) {
    e.Str("op", "insert").Int("rows", 5)
}).Msg("batch")
```

## Configuration (types.LoggingConfig)
Relevant fields (non-exhaustive):
- `Level`: `trace|debug|info|warn|error|fatal|panic`
- `WithTimestamp`: include timestamp field
- `SkipFrameCount`: enable caller info with given skip frames when > 0
- `ConsoleLogging` / `FileLogging`: enable writers; if both false, file logging is enabled by default
- `RelLogFileDir`: relative directory for log files (validated for safety; created on init)
- `LogFileMaxBackups`, `LogFileMaxAgeDays`, `LogFileMaxSizeMB`, `LogFileCompress`
- `ConsoleNoColor`, `ConsoleTimeFormat`
- `ShutdownTimeoutMS`, `ShutdownTimeoutWarning`

## Shutdown semantics
On `Close()`, the service stops accepting new logs and waits for in-flight operations to finish up to `ShutdownTimeoutMS` (default 100ms). If the timeout elapses and `ShutdownTimeoutWarning` is true, a warning is emitted.

## Dump helper
`Dump(v any)` recursively logs the content at Debug level, with depth/size limits. Intended for debugging; may delay shutdown if used right before `Close()`.