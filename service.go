package logging

import (
	"errors"
	"fmt"
	"github.com/7Q-Station-Manager/config"
	"github.com/7Q-Station-Manager/types"
	"github.com/7Q-Station-Manager/utils"
	"github.com/rs/zerolog"
	"go.uber.org/atomic"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Service struct {
	WorkingDir    string          `di.inject:"WorkingDir"`
	AppConfig     *config.Service `di.inject:"config"`
	LoggingConfig *types.LoggingConfig
	logger        atomic.Pointer[zerolog.Logger]
	initialized   atomic.Bool
}

// sprintPool is a buffer pool for legacy fmt.Sprint operations to reduce allocations
var sprintPool = sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

func NewLogger() *Service {
	return &Service{}
}

// Initialize initializes the logger.
func (l *Service) Initialize() error {
	if l.WorkingDir == emptyString {
		return errors.New("working dir has not been set/injected")
	}
	if l.AppConfig == nil {
		return errors.New("logger config has not been set/injected")
	}

	cfg := l.AppConfig.LoggingConfig()
	l.LoggingConfig = &cfg

	dir := filepath.Join(l.WorkingDir, l.LoggingConfig.RelLogFileDir)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	exeName, err := utils.ExecName(true)
	if err != nil {
		return fmt.Errorf("failed to get executable name: %w", err)
	}

	var writers []io.Writer

	if l.LoggingConfig.FileLogging {
		writers = append(writers, l.initializeRollingFileLogger(exeName))
	}
	if l.LoggingConfig.ConsoleLogging {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr})
	}
	if len(writers) == 0 {
		return errors.New("no logging channels enabled")
	}

	mw := io.MultiWriter(writers...)

	logger := zerolog.New(mw).With().Logger()

	level, err := getLevel(l.LoggingConfig.Level)
	if err != nil {
		return fmt.Errorf("setting logging level: %w", err)
	}
	logger = logger.Level(level)

	if l.LoggingConfig.WithTimestamp {
		logger = logger.With().Timestamp().Logger()
	}

	if l.LoggingConfig.SkipFrameCount > 0 {
		logger = logger.With().CallerWithSkipFrameCount(l.LoggingConfig.SkipFrameCount).Logger()
	}

	// Store logger atomically
	l.logger.Store(&logger)

	l.initialized.Store(true)
	return nil
}

// Close closes the logger and releases any resources.
// It's safe to call Close multiple times.
func (l *Service) Close() error {
	// No resources to clean up currently
	// Method kept for interface compatibility
	return nil
}

// Deprecated: use InfoWith()
func (l *Service) Info(fields ...interface{}) {
	if !l.initialized.Load() {
		return
	}
	logger := l.logger.Load()
	if logger == nil {
		return
	}

	// Use buffer pool to avoid allocations
	buf := sprintPool.Get().(*strings.Builder)
	buf.Reset()
	defer sprintPool.Put(buf)

	fmt.Fprint(buf, fields...)
	logger.Info().Msg(buf.String())
}

// Deprecated: use InfoWith()
func (l *Service) Infof(format string, fields ...interface{}) {
	if !l.initialized.Load() {
		return
	}
	logger := l.logger.Load()
	if logger == nil {
		return
	}
	logger.Info().Msgf(format, fields...)
}

// Deprecated: use DebugWith()
func (l *Service) Debug(fields ...interface{}) {
	if !l.initialized.Load() {
		return
	}
	logger := l.logger.Load()
	if logger == nil {
		return
	}

	buf := sprintPool.Get().(*strings.Builder)
	buf.Reset()
	defer sprintPool.Put(buf)

	fmt.Fprint(buf, fields...)
	logger.Debug().Msg(buf.String())
}

// Deprecated: use DebugWith()
func (l *Service) Debugf(format string, fields ...interface{}) {
	if !l.initialized.Load() {
		return
	}
	logger := l.logger.Load()
	if logger == nil {
		return
	}
	logger.Debug().Msgf(format, fields...)
}

// Deprecated: use WarnWith()
func (l *Service) Warn(fields ...interface{}) {
	if !l.initialized.Load() {
		return
	}
	logger := l.logger.Load()
	if logger == nil {
		return
	}

	buf := sprintPool.Get().(*strings.Builder)
	buf.Reset()
	defer sprintPool.Put(buf)

	fmt.Fprint(buf, fields...)
	logger.Warn().Msg(buf.String())
}

// Deprecated: use WarnWith()
func (l *Service) Warnf(format string, fields ...interface{}) {
	if !l.initialized.Load() {
		return
	}
	logger := l.logger.Load()
	if logger == nil {
		return
	}
	logger.Warn().Msgf(format, fields...)
}

// Deprecated: use ErrorWith()
func (l *Service) Error(fields ...interface{}) {
	if !l.initialized.Load() {
		return
	}
	logger := l.logger.Load()
	if logger == nil {
		return
	}

	buf := sprintPool.Get().(*strings.Builder)
	buf.Reset()
	defer sprintPool.Put(buf)

	fmt.Fprint(buf, fields...)
	logger.Error().Msg(buf.String())
}

// Deprecated: use ErrorWith()
func (l *Service) Errorf(format string, fields ...interface{}) {
	if !l.initialized.Load() {
		return
	}
	logger := l.logger.Load()
	if logger == nil {
		return
	}
	logger.Error().Msgf(format, fields...)
}

// Deprecated: use FatalWith()
func (l *Service) Fatal(fields ...interface{}) {
	if !l.initialized.Load() {
		_, _ = fmt.Fprintln(os.Stderr, "FATAL:", fmt.Sprint(fields...))
		os.Exit(1)
	}
	logger := l.logger.Load()
	if logger == nil {
		_, _ = fmt.Fprintln(os.Stderr, "FATAL:", fmt.Sprint(fields...))
		os.Exit(1)
	}

	buf := sprintPool.Get().(*strings.Builder)
	buf.Reset()
	defer sprintPool.Put(buf)

	fmt.Fprint(buf, fields...)
	logger.Fatal().Msg(buf.String())
}

// Deprecated: use FatalWith()
func (l *Service) Fatalf(format string, fields ...interface{}) {
	if !l.initialized.Load() {
		_, _ = fmt.Fprintf(os.Stderr, "FATAL: "+format+"\n", fields...)
		os.Exit(1)
	}
	logger := l.logger.Load()
	if logger == nil {
		_, _ = fmt.Fprintf(os.Stderr, "FATAL: "+format+"\n", fields...)
		os.Exit(1)
	}
	logger.Fatal().Msgf(format, fields...)
}

func (l *Service) Hook(hooks ...zerolog.Hook) {
	if !l.initialized.Load() {
		return
	}

	// Atomic compare-and-swap loop for thread-safe hook installation
	for {
		oldLogger := l.logger.Load()
		if oldLogger == nil {
			return
		}

		newLogger := oldLogger.Hook(hooks...)

		// Try to swap - if another goroutine changed it, retry
		if l.logger.CompareAndSwap(oldLogger, &newLogger) {
			break
		}
	}
}

// Structured logging methods

// InfoWith returns a LogEvent for structured Info-level logging.
// Use this for queryable, structured logs instead of Info/Infof.
// Example: logger.InfoWith().Str("user_id", id).Int("count", 5).Msg("User processed")
func (l *Service) InfoWith() LogEvent {
	if !l.initialized.Load() {
		return newLogEvent(nil)
	}
	logger := l.logger.Load()
	if logger == nil {
		return newLogEvent(nil)
	}
	// Early return if debug level is not enabled
	if logger.GetLevel() > zerolog.InfoLevel {
		return newLogEvent(nil)
	}
	return newLogEvent(logger.Info())
}

// WarnWith returns a LogEvent for structured Warn-level logging.
func (l *Service) WarnWith() LogEvent {
	if !l.initialized.Load() {
		return newLogEvent(nil)
	}
	logger := l.logger.Load()
	if logger == nil {
		return newLogEvent(nil)
	}
	// Early return if debug level is not enabled
	if logger.GetLevel() > zerolog.WarnLevel {
		return newLogEvent(nil)
	}
	return newLogEvent(logger.Warn())
}

// ErrorWith returns a LogEvent for structured Error-level logging.
// Example: logger.ErrorWith().Err(err).Str("operation", "database").Msg("Query failed")
func (l *Service) ErrorWith() LogEvent {
	if !l.initialized.Load() {
		return newLogEvent(nil)
	}
	logger := l.logger.Load()
	if logger == nil {
		return newLogEvent(nil)
	}
	return newLogEvent(logger.Error())
}

// DebugWith returns a LogEvent for structured Debug-level logging.
func (l *Service) DebugWith() LogEvent {
	if !l.initialized.Load() {
		return newLogEvent(nil)
	}
	logger := l.logger.Load()
	if logger == nil {
		return newLogEvent(nil)
	}
	// Early return if debug level is not enabled
	if logger.GetLevel() > zerolog.DebugLevel {
		return newLogEvent(nil)
	}
	return newLogEvent(logger.Debug())
}

// FatalWith returns a LogEvent for structured Fatal-level logging.
// The program will exit after the log is written.
func (l *Service) FatalWith() LogEvent {
	if !l.initialized.Load() {
		// For fatal, we still need to exit even if not initialized
		return newLogEvent(nil)
	}
	logger := l.logger.Load()
	if logger == nil {
		return newLogEvent(nil)
	}
	return newLogEvent(logger.Fatal())
}

// With returns a LogContext for creating a child logger with pre-populated fields.
// Example: reqLogger := logger.With().Str("request_id", id).Logger()
func (l *Service) With() LogContext {
	if !l.initialized.Load() {
		// Return a context that will create a properly initialized logger later
		return &logContext{
			context: zerolog.New(nil).With(),
			service: l,
		}
	}
	logger := l.logger.Load()
	if logger == nil {
		return &logContext{
			context: zerolog.New(nil).With(),
			service: l,
		}
	}
	return &logContext{
		context: logger.With(),
		service: l,
	}
}
