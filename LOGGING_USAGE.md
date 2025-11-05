# Logging Usage Guide

This document provides comprehensive guidance on using the 7Q-Station-Manager logging module.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Traditional Logging (Legacy)](#traditional-logging-legacy)
- [Structured Logging (Recommended)](#structured-logging-recommended)
- [Context Loggers](#context-loggers)
- [Advanced Features](#advanced-features)
- [Best Practices](#best-practices)
- [Performance Considerations](#performance-considerations)
- [Migration Guide](#migration-guide)

---

## Overview

The logging module provides two styles of logging:

1. **Traditional/Legacy Logging**: String-based methods like `Info()`, `Infof()`, `Error()`, etc.
2. **Structured Logging**: Type-safe, queryable logging with methods like `InfoWith()`, `ErrorWith()`, etc.

**For all new code, use structured logging.** It provides better observability, queryability, and integration with modern log aggregation tools.

---

## Quick Start

### Initialization

```go
import "github.com/7Q-Station-Manager/logging"

// Create and initialize logger
logger := logging.NewLogger()
logger.WorkingDir = "/path/to/workdir"
logger.AppConfig = configService // your config.Service instance

if err := logger.Initialize(); err != nil {
    panic(err)
}
defer logger.Close()
```

### Basic Structured Logging

```go
// Simple structured log
logger.InfoWith().
    Str("user_id", "12345").
    Int("age", 30).
    Msg("User logged in")

// Output: {"level":"info","user_id":"12345","age":30,"message":"User logged in"}
```

---

## Traditional Logging (Legacy)

> ⚠️ **Note**: These methods are kept for backward compatibility. Use structured logging for new code.

### String-based Methods

```go
// Simple messages
logger.Info("Application started")
logger.Warn("Connection slow")
logger.Error("Failed to connect")
logger.Debug("Cache miss")

// Formatted messages
logger.Infof("User %s logged in from %s", userID, ipAddr)
logger.Errorf("Query failed after %d retries", retryCount)

// Fatal (exits program)
logger.Fatal("Critical error, cannot continue")
logger.Fatalf("Database unreachable at %s", dbHost)
```

### Limitations

❌ Not queryable - everything is a string
❌ Type information lost
❌ Cannot filter by field values
❌ Poor integration with log aggregation tools
❌ Performance overhead from string formatting

---

## Structured Logging (Recommended)

Structured logging produces JSON output with typed fields that can be queried, filtered, and analyzed programmatically.

### Basic Usage

```go
logger.InfoWith().
    Str("user_id", "user-123").
    Int("count", 42).
    Bool("success", true).
    Msg("Operation completed")

// Output:
// {"level":"info","user_id":"user-123","count":42,"success":true,"message":"Operation completed"}
```

### Available Field Types

#### String Fields
```go
logger.InfoWith().
    Str("key", "value").                              // Single string
    Strs("tags", []string{"golang", "logging"}).      // String array
    Stringer("obj", customStringer).                  // fmt.Stringer interface
    Msg("String fields")
```

#### Numeric Fields
```go
logger.InfoWith().
    Int("count", 100).
    Int8("small", 1).
    Int16("medium", 1000).
    Int32("large", 100000).
    Int64("huge", 9223372036854775807).
    Uint("port", 8080).
    Uint8("byte", 255).
    Uint16("word", 65535).
    Uint32("dword", 4294967295).
    Uint64("qword", 18446744073709551615).
    Float32("temp", 98.6).
    Float64("pi", 3.14159265359).
    Msg("Numeric fields")
```

#### Boolean Fields
```go
logger.InfoWith().
    Bool("active", true).
    Bool("verified", false).
    Bools("flags", []bool{true, false, true}).
    Msg("Boolean fields")
```

#### Time and Duration
```go
import "time"

logger.InfoWith().
    Time("created_at", time.Now()).
    Dur("elapsed", 250*time.Millisecond).
    Msg("Timing fields")

// Output: {"level":"info","created_at":"2025-10-04T10:30:00Z","elapsed":250,"message":"Timing fields"}
```

#### Error Fields
```go
err := errors.New("connection timeout")

logger.ErrorWith().
    Err(err).                           // Primary error (appears as "error" field)
    Str("operation", "database").
    Int("retry_count", 3).
    Msg("Operation failed")

// Multiple errors
logger.ErrorWith().
    Err(primaryErr).
    AnErr("validation_error", validationErr).
    AnErr("network_error", networkErr).
    Msg("Multiple errors occurred")
```

#### Binary and Bytes
```go
data := []byte{0xDE, 0xAD, 0xBE, 0xEF}

logger.DebugWith().
    Bytes("raw", data).                 // Base64 encoded
    Hex("checksum", data).              // Hex encoded
    Msg("Binary data")
```

#### Network Fields
```go
import "net"

ip := net.ParseIP("192.168.1.1")
mac, _ := net.ParseMAC("00:11:22:33:44:55")

logger.InfoWith().
    IPAddr("client_ip", ip).
    MACAddr("device_mac", mac).
    Msg("Network info")
```

#### Generic Interface
```go
// For arbitrary types (uses reflection/JSON encoding)
logger.InfoWith().
    Interface("config", configObject).
    Interface("metadata", map[string]interface{}{
        "version": "1.0",
        "build": 123,
    }).
    Msg("Complex objects")
```

### Nested Objects (Dictionaries)

Create nested JSON structures:

```go
logger.InfoWith().
    Str("event", "user_action").
    Dict("user", func(e logging.LogEvent) {
        e.Str("id", "user-123")
        e.Str("email", "user@example.com")
        e.Int("age", 30)
    }).
    Dict("metadata", func(e logging.LogEvent) {
        e.Str("ip", "192.168.1.1")
        e.Time("timestamp", time.Now())
        e.Bool("verified", true)
    }).
    Msg("User performed action")

// Output:
// {
//   "level": "info",
//   "event": "user_action",
//   "user": {
//     "id": "user-123",
//     "email": "user@example.com",
//     "age": 30
//   },
//   "metadata": {
//     "ip": "192.168.1.1",
//     "timestamp": "2025-10-04T10:30:00Z",
//     "verified": true
//   },
//   "message": "User performed action"
// }
```

### Log Levels

```go
// Debug - detailed diagnostic information
logger.DebugWith().Str("cache_key", key).Msg("Cache lookup")

// Info - general informational messages
logger.InfoWith().Str("user_id", id).Msg("Request processed")

// Warn - warning messages for potentially harmful situations
logger.WarnWith().Int("retry_count", 5).Msg("Retrying operation")

// Error - error events that might still allow the app to continue
logger.ErrorWith().Err(err).Str("operation", "save").Msg("Failed to save")

// Fatal - severe errors that will cause the application to exit
logger.FatalWith().Err(err).Msg("Cannot start application")
```

---

## Context Loggers

Context loggers allow you to create child loggers with pre-populated fields that are included in all subsequent log messages. This is extremely useful for request tracing, session tracking, and correlation IDs.

### Creating Context Loggers

```go
// Create a logger with request context
requestLogger := logger.With().
    Str("request_id", requestID).
    Str("user_id", userID).
    Str("ip", clientIP).
    Logger()

// All logs from requestLogger will include these fields
requestLogger.InfoWith().Str("action", "start").Msg("Processing request")
requestLogger.InfoWith().Int("status", 200).Dur("duration", elapsed).Msg("Request completed")

// Both logs include request_id, user_id, and ip automatically
```

### Nested Context Loggers

```go
// Base logger with application context
appLogger := logger.With().
    Str("service", "api").
    Str("version", "1.0.0").
    Logger()

// Request-specific logger inherits application context
requestLogger := appLogger.With().
    Str("request_id", "req-123").
    Logger()

// Operation-specific logger inherits both contexts
dbLogger := requestLogger.With().
    Str("component", "database").
    Logger()

// This log includes: service, version, request_id, AND component
dbLogger.InfoWith().Str("query", "SELECT *").Msg("Query executed")
```

### Use Cases for Context Loggers

**HTTP Request Tracking**
```go
func HandleRequest(w http.ResponseWriter, r *http.Request, logger logging.Logger) {
    requestLogger := logger.With().
        Str("request_id", r.Header.Get("X-Request-ID")).
        Str("method", r.Method).
        Str("path", r.URL.Path).
        Str("remote_addr", r.RemoteAddr).
        Logger()

    requestLogger.InfoWith().Msg("Request started")

    // Pass requestLogger to handlers
    processRequest(requestLogger)

    requestLogger.InfoWith().Int("status", 200).Msg("Request completed")
}
```

**Database Operations**
```go
func ProcessOrder(orderID string, logger logging.Logger) error {
    orderLogger := logger.With().
        Str("order_id", orderID).
        Logger()

    orderLogger.InfoWith().Msg("Processing order")

    if err := validateOrder(orderLogger); err != nil {
        orderLogger.ErrorWith().Err(err).Msg("Validation failed")
        return err
    }

    if err := saveOrder(orderLogger); err != nil {
        orderLogger.ErrorWith().Err(err).Msg("Save failed")
        return err
    }

    orderLogger.InfoWith().Msg("Order processed successfully")
    return nil
}
```

**Background Jobs**
```go
func RunJob(jobID string, logger logging.Logger) {
    jobLogger := logger.With().
        Str("job_id", jobID).
        Time("started_at", time.Now()).
        Logger()

    jobLogger.InfoWith().Msg("Job started")
    defer jobLogger.InfoWith().Msg("Job completed")

    // All logs in job will include job_id and started_at
}
```

---

## Advanced Features

### Dump Function

For debugging complex objects:

```go
type User struct {
    ID    string
    Name  string
    Age   int
    Roles []string
}

user := User{
    ID:    "user-123",
    Name:  "John Doe",
    Age:   30,
    Roles: []string{"admin", "user"},
}

// Dump logs all fields with reflection
logger.Dump(user)

// Output (debug level):
// Struct: User
// ID: user-123
// Name: John Doe
// Age: 30
// Roles: []string (len: 2, cap: 2) {
//   [0]: admin
//   [1]: user
// }
```

**Dump Features:**
- Handles circular references
- Limits recursion depth (max 10 levels)
- Limits array/slice elements (max 10 elements shown)
- Skips unexported fields
- Works with maps, slices, structs, and basic types

### Hooks

Add custom behavior to logging:

```go
import "github.com/rs/zerolog"

// Example: Add hostname to every log
type HostnameHook struct {
    hostname string
}

func (h HostnameHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
    e.Str("hostname", h.hostname)
}

// Register hook
hostname, _ := os.Hostname()
logger.Hook(HostnameHook{hostname: hostname})

// All logs will now include hostname field
```

---

## Best Practices

### ✅ Do's

1. **Use Structured Logging for New Code**
   ```go
   // Good
   logger.InfoWith().Str("user_id", id).Int("count", n).Msg("Processed")

   // Avoid
   logger.Infof("Processed user_id=%s count=%d", id, n)
   ```

2. **Use Context Loggers for Related Operations**
   ```go
   requestLogger := logger.With().Str("request_id", reqID).Logger()
   // Use requestLogger throughout the request lifecycle
   ```

3. **Log Errors with Context**
   ```go
   logger.ErrorWith().
       Err(err).
       Str("operation", "database_query").
       Str("table", "users").
       Int("retry_count", 3).
       Msg("Query failed")
   ```

4. **Use Appropriate Log Levels**
   - Debug: Detailed diagnostic info (disabled in production)
   - Info: General informational messages
   - Warn: Potentially harmful situations
   - Error: Error events that might allow app to continue
   - Fatal: Severe errors that cause termination

5. **Add Timing Information**
   ```go
   start := time.Now()
   // ... operation ...
   logger.InfoWith().
       Dur("elapsed", time.Since(start)).
       Msg("Operation completed")
   ```

### ❌ Don'ts

1. **Don't Log Sensitive Data**
   ```go
   // Bad - logs password
   logger.InfoWith().Str("password", pwd).Msg("User created")

   // Good - redact sensitive data
   logger.InfoWith().Str("user_id", id).Msg("User created")
   ```

2. **Don't Log in Tight Loops (Without Sampling)**
   ```go
   // Bad - logs millions of times
   for i := 0; i < 1000000; i++ {
       logger.DebugWith().Int("i", i).Msg("Processing")
   }

   // Good - log periodically
   for i := 0; i < 1000000; i++ {
       if i%10000 == 0 {
           logger.DebugWith().Int("progress", i).Msg("Processing")
       }
   }
   ```

3. **Don't Use Interface{} Unnecessarily**
   ```go
   // Bad - loses type information
   logger.InfoWith().Interface("count", count).Msg("Count")

   // Good - use typed method
   logger.InfoWith().Int("count", count).Msg("Count")
   ```

4. **Don't Create Deep Nesting**
   ```go
   // Bad - hard to read and query
   logger.InfoWith().
       Dict("level1", func(e logging.LogEvent) {
           e.Dict("level2", func(e logging.LogEvent) {
               e.Dict("level3", func(e logging.LogEvent) {
                   e.Str("deeply", "nested")
               })
           })
       }).Msg("Too nested")

   // Good - flatten structure
   logger.InfoWith().
       Str("level1_level2_level3_deeply", "nested").
       Msg("Flattened")
   ```

5. **Don't Mix Logging Styles in Same Function**
   ```go
   // Bad - inconsistent
   logger.Info("Starting process")
   logger.InfoWith().Str("user_id", id).Msg("Processing user")
   logger.Infof("Completed in %d ms", elapsed)

   // Good - consistent structured logging
   logger.InfoWith().Msg("Starting process")
   logger.InfoWith().Str("user_id", id).Msg("Processing user")
   logger.InfoWith().Int("elapsed_ms", elapsed).Msg("Completed")
   ```

---

## Performance Considerations

### Zero Allocation Logging

Structured logging with zerolog is designed for zero allocations when possible:

```go
// No allocations (fields are added directly to the event)
logger.InfoWith().
    Str("key", "value").
    Int("count", 42).
    Msg("Efficient")
```

### Disabled Levels

When a log level is disabled (e.g., Debug in production), the entire chain short-circuits:

```go
// If debug is disabled, this has virtually zero cost
logger.DebugWith().
    Str("expensive", computeExpensiveString()). // Never called!
    Msg("Debug info")
```

### Conditional Logging

For expensive operations, check the level first:

```go
// Only compute if debug is enabled
if logger.DebugWith() != nil {
    expensiveData := computeExpensiveDebugging()
    logger.DebugWith().
        Interface("debug_data", expensiveData).
        Msg("Debug information")
}
```

### Concurrency

The logger is thread-safe and optimized for concurrent use:

```go
// Safe to use from multiple goroutines
go logger.InfoWith().Str("goroutine", "1").Msg("Worker 1")
go logger.InfoWith().Str("goroutine", "2").Msg("Worker 2")
go logger.InfoWith().Str("goroutine", "3").Msg("Worker 3")
```

---

## Migration Guide

### From Traditional to Structured Logging

**Before:**
```go
logger.Infof("User %s logged in from %s at %s", userID, ip, time.Now())
```

**After:**
```go
logger.InfoWith().
    Str("user_id", userID).
    Str("ip", ip).
    Time("logged_in_at", time.Now()).
    Msg("User logged in")
```

**Benefits:**
- ✅ Queryable: `jq '.user_id == "12345"'`
- ✅ Filterable by IP range
- ✅ Parseable timestamps
- ✅ Better performance

### Error Logging Migration

**Before:**
```go
if err != nil {
    logger.Errorf("Failed to save user %s: %v", userID, err)
}
```

**After:**
```go
if err != nil {
    logger.ErrorWith().
        Err(err).
        Str("user_id", userID).
        Str("operation", "save_user").
        Msg("Failed to save user")
}
```

### Gradual Migration Strategy

1. **Start with new code** - Use structured logging for all new features
2. **High-value logs first** - Migrate error and warning logs first
3. **Request paths** - Add context loggers to HTTP handlers
4. **Background jobs** - Migrate batch processing and scheduled tasks
5. **Legacy code** - Update as you touch old code (no rush)

**Coexistence is fine:**
```go
// Old code can keep using traditional logging
logger.Info("Legacy message")

// New code uses structured logging
logger.InfoWith().Str("new", "style").Msg("Modern message")
```

---

## Configuration

### Configuration Options

The logger is configured through the `types.LoggingConfig` structure:

```go
type LoggingConfig struct {
    Level              string  // "debug", "info", "warn", "error", "fatal"
    SkipFrameCount     int     // Number of stack frames to skip for caller info
    WithTimestamp      bool    // Include timestamp in logs
    ConsoleLogging     bool    // Log to console (stderr)
    FileLogging        bool    // Log to file
    RelLogFileDir      string  // Directory for log files (relative to WorkingDir)
    LogFileMaxBackups  int     // Maximum number of old log files to keep
    LogFileMaxAgeDays  int     // Maximum age of log files in days
    LogFileMaxSizeMB   int     // Maximum size of log file in MB before rotation
}
```

### Example Configuration (JSON)

```json
{
  "LoggingConfig": {
    "Level": "info",
    "SkipFrameCount": 0,
    "WithTimestamp": true,
    "ConsoleLogging": true,
    "FileLogging": true,
    "RelLogFileDir": "logs",
    "LogFileMaxBackups": 5,
    "LogFileMaxAgeDays": 30,
    "LogFileMaxSizeMB": 100
  }
}
```

---

## Querying Structured Logs

### Using jq (command-line JSON processor)

```bash
# Find all errors
cat logs/app.log | jq 'select(.level == "error")'

# Find logs for specific user
cat logs/app.log | jq 'select(.user_id == "user-123")'

# Find slow operations (>1 second)
cat logs/app.log | jq 'select(.elapsed > 1000)'

# Count errors by operation type
cat logs/app.log | jq 'select(.level == "error") | .operation' | sort | uniq -c

# Get average duration
cat logs/app.log | jq -s 'map(.elapsed) | add / length'
```

### Integration with Log Aggregation Tools

**Elasticsearch:**
```json
POST /logs/_search
{
  "query": {
    "bool": {
      "must": [
        { "term": { "level": "error" } },
        { "term": { "user_id": "user-123" } },
        { "range": { "@timestamp": { "gte": "now-1h" } } }
      ]
    }
  }
}
```

**Splunk:**
```
index=app level=error user_id="user-123" | stats count by operation
```

**Datadog:**
```
status:error user_id:user-123 @elapsed:>1000
```

---

## Troubleshooting

### Logs Not Appearing

1. **Check initialization:**
   ```go
   if err := logger.Initialize(); err != nil {
       fmt.Printf("Logger init failed: %v\n", err)
   }
   ```

2. **Check log level:**
   ```go
   // Debug logs won't appear if level is "info" or higher
   logger.DebugWith().Msg("This won't show if level >= info")
   ```

3. **Check output channels:**
   ```go
   // At least one must be enabled
   ConsoleLogging: true,  // or
   FileLogging: true,
   ```

### Performance Issues

1. **Avoid expensive operations in log statements:**
   ```go
   // Bad - always computed even if logging disabled
   logger.DebugWith().Str("data", expensiveOperation()).Msg("Debug")

   // Good - only computed if needed
   if isDebugEnabled() {
       data := expensiveOperation()
       logger.DebugWith().Str("data", data).Msg("Debug")
   }
   ```

2. **Use sampling for high-frequency logs:**
   ```go
   counter := 0
   for item := range items {
       if counter%100 == 0 { // Only log every 100th item
           logger.InfoWith().Int("progress", counter).Msg("Processing")
       }
       counter++
   }
   ```

---

## Examples

### Complete HTTP Handler Example

```go
func HandleUserRequest(w http.ResponseWriter, r *http.Request, logger logging.Logger) {
    start := time.Now()

    // Create request logger with context
    requestLogger := logger.With().
        Str("request_id", r.Header.Get("X-Request-ID")).
        Str("method", r.Method).
        Str("path", r.URL.Path).
        Str("remote_addr", r.RemoteAddr).
        Logger()

    requestLogger.InfoWith().Msg("Request started")

    // Extract user ID from path
    userID := r.URL.Query().Get("user_id")
    if userID == "" {
        requestLogger.WarnWith().Msg("Missing user_id parameter")
        http.Error(w, "Missing user_id", http.StatusBadRequest)
        return
    }

    // Create user-specific logger
    userLogger := requestLogger.With().
        Str("user_id", userID).
        Logger()

    // Fetch user data
    user, err := fetchUser(userID)
    if err != nil {
        userLogger.ErrorWith().
            Err(err).
            Str("operation", "fetch_user").
            Msg("Failed to fetch user")
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    userLogger.InfoWith().
        Str("user_name", user.Name).
        Bool("user_active", user.Active).
        Msg("User fetched successfully")

    // Process user data
    result, err := processUser(user, userLogger)
    if err != nil {
        userLogger.ErrorWith().
            Err(err).
            Str("operation", "process_user").
            Msg("Failed to process user")
        http.Error(w, "Processing failed", http.StatusInternalServerError)
        return
    }

    // Success
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(result)

    elapsed := time.Since(start)
    requestLogger.InfoWith().
        Int("status", http.StatusOK).
        Dur("elapsed", elapsed).
        Msg("Request completed")
}
```

---

## Additional Resources

- **Zerolog Documentation:** https://github.com/rs/zerolog
- **JSON Logging Best Practices:** https://www.loggly.com/ultimate-guide/json-logging-best-practices/
- **Log Aggregation:** Search for "ELK Stack", "Splunk", "Datadog" tutorials

---

## Summary

- ✅ Use **structured logging** for all new code (`InfoWith()`, `ErrorWith()`, etc.)
- ✅ Use **context loggers** (`With()`) for related operations
- ✅ Log **typed fields** (Str, Int, Err) instead of formatted strings
- ✅ Use appropriate **log levels** (Debug, Info, Warn, Error, Fatal)
- ✅ Add **timing information** (Dur) for operations
- ✅ Include **error context** (Err, operation, retry_count)
- ✅ Make logs **queryable** for better observability
- ⚠️ Avoid logging **sensitive data** (passwords, tokens, PII)
- ⚠️ Minimize logging in **tight loops** (use sampling)
- ⚠️ Traditional logging methods are **legacy** (use for backward compatibility only)

For questions or issues, refer to the module source code or consult the team.
