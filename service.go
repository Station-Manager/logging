package logging

import (
	stderr "errors"
	"github.com/Station-Manager/config"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	"github.com/Station-Manager/utils"
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
	AppConfig     *config.Service `di.inject:"appconfig"`
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

var initOnce sync.Once

// Initialize initializes the logger.
func (s *Service) Initialize() error {
	const op errors.Op = "logging.Service.Initialize"
	if s == nil {
		return errors.New(op).Msg(errMsgNilService)
	}

	if s.AppConfig == nil {
		return errors.New(op).Msg(errMsgAppCfgNotSet)
	}

	var err error

	initOnce.Do(func() {
		if s.WorkingDir == emptyString {
			exeDir, pathErr := utils.AbsDirPathForExecutable()
			if pathErr != nil {
				err = stderr.Join(errors.New(op).Errorf("utils.AbsDirPathForExecutable: %w", pathErr))
				return
			}
			s.WorkingDir = filepath.Join(exeDir, "logs")
		}

		loggingCfg, cfgErr := s.AppConfig.LoggingConfig()
		if cfgErr != nil {
			err = stderr.Join(errors.New(op).Errorf("s.AppConfig.LoggingConfig: %w", cfgErr))
			return
		}
		s.LoggingConfig = &loggingCfg

		if s.LoggingConfig.RelLogFileDir == emptyString {
			s.LoggingConfig.RelLogFileDir = "logs"
		}

		loggingDir := filepath.Join(s.WorkingDir, s.LoggingConfig.RelLogFileDir)
		exists, existsErr := utils.PathExists(loggingDir)
		if existsErr != nil {
			err = stderr.Join(errors.New(op).Errorf("utils.PathExists: %w", existsErr))
			return
		}

		if !exists {
			if mdErr := os.MkdirAll(loggingDir, os.ModePerm); mdErr != nil {
				err = stderr.Join(errors.New(op).Errorf("os.MkdirAll: %w", mdErr))
				return
			}
		}

		exeName, exeErr := utils.ExecName(true)
		if exeErr != nil {
			err = stderr.Join(errors.New(op).Errorf("utils.ExecName: %w", exeErr))
			return
		}

		mw := io.MultiWriter(s.initializeWriters(exeName)...)
		logger := zerolog.New(mw).With().Logger()

		// If the level is not set, default to info. Also, assume a blank config
		// so fill in some sensible defaults. See config/defaults.go
		if s.LoggingConfig.Level == emptyString {
			s.LoggingConfig.Level = "info"
			s.LoggingConfig.WithTimestamp = true
			s.LoggingConfig.SkipFrameCount = 3
			s.LoggingConfig.LogFileMaxSizeMB = 100
			s.LoggingConfig.LogFileMaxAgeDays = 30
			s.LoggingConfig.LogFileMaxBackups = 5
		}

		level, levelErr := getLevel(s.LoggingConfig.Level)
		if levelErr != nil {
			err = stderr.Join(errors.New(op).Errorf("getLevel: %w", levelErr))
			return
		}
		logger = logger.Level(level)

		if s.LoggingConfig.WithTimestamp {
			logger = logger.With().Timestamp().Logger()
		}

		if s.LoggingConfig.SkipFrameCount > 0 {
			logger = logger.With().CallerWithSkipFrameCount(s.LoggingConfig.SkipFrameCount).Logger()
		}

		// Store logger atomically
		s.logger.Store(&logger)

		s.initialized.Store(true)
	})

	return err
}

func (s *Service) Hook(hooks ...zerolog.Hook) {
	if !s.initialized.Load() {
		return
	}

	// Atomic compare-and-swap loop for thread-safe hook installation
	for {
		oldLogger := s.logger.Load()
		if oldLogger == nil {
			return
		}

		newLogger := oldLogger.Hook(hooks...)

		// Try to swap - if another goroutine changed it, retry
		if s.logger.CompareAndSwap(oldLogger, &newLogger) {
			break
		}
	}
}

// Structured logging methods

// InfoWith returns a LogEvent for structured Info-level logging.
// Use this for queryable, structured logs instead of Info/Infof.
// Example: logger.InfoWith().Str("user_id", id).Int("count", 5).Msg("User processed")
func (s *Service) InfoWith() LogEvent {
	if !s.initialized.Load() {
		return newLogEvent(nil)
	}
	logger := s.logger.Load()
	if logger == nil {
		return newLogEvent(nil)
	}
	// Early return if info level is not enabled
	if logger.GetLevel() > zerolog.InfoLevel {
		return newLogEvent(nil)
	}
	return newLogEvent(logger.Info())
}

// WarnWith returns a LogEvent for structured Warn-level logging.
func (s *Service) WarnWith() LogEvent {
	if !s.initialized.Load() {
		return newLogEvent(nil)
	}
	logger := s.logger.Load()
	if logger == nil {
		return newLogEvent(nil)
	}
	// Early return if warn level is not enabled
	if logger.GetLevel() > zerolog.WarnLevel {
		return newLogEvent(nil)
	}
	return newLogEvent(logger.Warn())
}

// ErrorWith returns a LogEvent for structured Error-level logging.
// Example: logger.ErrorWith().Err(err).Str("operation", "database").Msg("Query failed")
func (s *Service) ErrorWith() LogEvent {
	if !s.initialized.Load() {
		return newLogEvent(nil)
	}
	logger := s.logger.Load()
	if logger == nil {
		return newLogEvent(nil)
	}
	return newLogEvent(logger.Error())
}

// DebugWith returns a LogEvent for structured Debug-level logging.
func (s *Service) DebugWith() LogEvent {
	if !s.initialized.Load() {
		return newLogEvent(nil)
	}
	logger := s.logger.Load()
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
func (s *Service) FatalWith() LogEvent {
	if !s.initialized.Load() {
		// For fatal, we still need to exit even if not initialized
		return newLogEvent(nil)
	}
	logger := s.logger.Load()
	if logger == nil {
		return newLogEvent(nil)
	}
	return newLogEvent(logger.Fatal())
}

// With returns a LogContext for creating a child logger with pre-populated fields.
// Example: reqLogger := logger.With().Str("request_id", id).Logger()
func (s *Service) With() LogContext {
	if !s.initialized.Load() {
		// Return a context that will create a properly initialized logger later
		return &logContext{
			context: zerolog.New(nil).With(),
			service: s,
		}
	}
	logger := s.logger.Load()
	if logger == nil {
		return &logContext{
			context: zerolog.New(nil).With(),
			service: s,
		}
	}
	return &logContext{
		context: logger.With(),
		service: s,
	}
}
