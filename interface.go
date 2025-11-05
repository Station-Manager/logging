package logging

// Logger provides both traditional string-based logging and structured logging capabilities.
// For new code, prefer the structured logging methods (InfoWith, ErrorWith, etc.) over
// the traditional methods (Info, Infof, etc.) for better queryability and observability.
type Logger interface {
	// Traditional logging methods (legacy, kept for backward compatibility)
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Dump(v interface{})

	// Structured logging methods (recommended)
	// These return a LogEvent that allows chaining typed field methods
	// Example: logger.InfoWith().Str("user_id", id).Int("count", 5).Msg("Processing complete")
	InfoWith() LogEvent
	WarnWith() LogEvent
	ErrorWith() LogEvent
	DebugWith() LogEvent
	FatalWith() LogEvent

	// Context logger creation
	// Creates a new logger with pre-populated fields that will be included in all subsequent logs
	// Example: reqLogger := logger.With().Str("request_id", id).Logger()
	With() LogContext
}
