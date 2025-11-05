package logging

// Logger provides both traditional string-based logging and structured logging capabilities.
// For new code, prefer the structured logging methods (InfoWith, ErrorWith, etc.) over
// the traditional methods (Info, Infof, etc.) for better queryability and observability.
type Logger interface {
	InfoWith() LogEvent
	WarnWith() LogEvent
	ErrorWith() LogEvent
	DebugWith() LogEvent
	FatalWith() LogEvent

	// With for context logger creation
	// Creates a new logger with pre-populated fields that will be included in all subsequent logs
	// Example: reqLogger := logger.With().Str("request_id", id).Logger()
	With() LogContext
}
