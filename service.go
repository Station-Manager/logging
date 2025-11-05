package logging

import (
	"github.com/Station-Manager/config"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	"github.com/Station-Manager/utils"
	"github.com/rs/zerolog"
	"go.uber.org/atomic"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type Service struct {
	WorkingDir    string          `di.inject:"WorkingDir"`
	AppConfig     *config.Service `di.inject:"appconfig"`
	LoggingConfig *types.LoggingConfig
	fileWriter    *lumberjack.Logger
	logger        atomic.Pointer[zerolog.Logger]
	isInitialized atomic.Bool
	initOnce      sync.Once
	mu            sync.Mutex
}

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

	s.initOnce.Do(func() {
		if s.WorkingDir == emptyString {
			exeDir, pathErr := utils.AbsDirPathForExecutable()
			if pathErr != nil {
				err = errors.New(op).Errorf("utils.AbsDirPathForExecutable: %w", pathErr)
				return
			}
			s.WorkingDir = exeDir
		}

		loggingCfg, cfgErr := s.AppConfig.LoggingConfig()
		if cfgErr != nil {
			err = errors.New(op).Errorf("s.AppConfig.LoggingConfig: %w", cfgErr)
			return
		}

		if cfgErr = validateConfig(&loggingCfg); cfgErr != nil {
			err = errors.New(op).Errorf("validateConfig: %w", cfgErr)
		}
		s.LoggingConfig = &loggingCfg

		loggingDir := filepath.Join(s.WorkingDir, s.LoggingConfig.RelLogFileDir)
		exists, existsErr := utils.PathExists(loggingDir)
		if existsErr != nil {
			err = errors.New(op).Errorf("utils.PathExists: %w", existsErr)
			return
		}

		if !exists {
			if mdErr := os.MkdirAll(loggingDir, 0755); mdErr != nil {
				err = errors.New(op).Errorf("os.MkdirAll: %w", mdErr)
				return
			}
		}

		exeName, exeErr := utils.ExecName(true)
		if exeErr != nil {
			err = errors.New(op).Errorf("utils.ExecName: %w", exeErr)
			return
		}

		mw := io.MultiWriter(s.initializeWriters(exeName)...)
		logger := zerolog.New(mw).With().Logger()

		level, levelErr := getLevel(s.LoggingConfig.Level)
		if levelErr != nil {
			err = errors.New(op).Errorf("getLevel: %w", levelErr)
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

		s.isInitialized.Store(true)
	})

	return err
}

func (s *Service) Close() error {
	const op errors.Op = "logging.Service.Close"
	if s == nil {
		return nil
	}
	if !s.isInitialized.Load() {
		return nil
	}

	s.isInitialized.Store(false)
	s.logger.Store(nil)

	// Close the file writer if it exists
	if s.fileWriter != nil {
		if err := s.fileWriter.Close(); err != nil {
			return errors.New(op).Errorf("s.fileWriter.Close: %w", err)
		}
	}

	return nil
}

// InfoWith returns a LogEvent for structured Info-level logging.
// Use this for queryable, structured logs instead of Info/Infof.
// Example: logger.InfoWith().Str("user_id", id).Int("count", 5).Msg("User processed")
func (s *Service) InfoWith() LogEvent {
	return logEventBuilder(s, zerolog.InfoLevel)
}

// WarnWith returns a LogEvent for structured Warn-level logging.
func (s *Service) WarnWith() LogEvent {
	return logEventBuilder(s, zerolog.WarnLevel)
}

// ErrorWith returns a LogEvent for structured Error-level logging.
// Example: logger.ErrorWith().Err(err).Str("operation", "database").Msg("Query failed")
func (s *Service) ErrorWith() LogEvent {
	return logEventBuilder(s, zerolog.ErrorLevel)
}

// DebugWith returns a LogEvent for structured Debug-level logging.
func (s *Service) DebugWith() LogEvent {
	return logEventBuilder(s, zerolog.DebugLevel)
}

// FatalWith returns a LogEvent for structured Fatal-level logging.
// The program will exit after the log is written.
func (s *Service) FatalWith() LogEvent {
	return logEventBuilder(s, zerolog.FatalLevel)
}

func (s *Service) PanicWith() LogEvent {
	return logEventBuilder(s, zerolog.PanicLevel)
}

// With returns a LogContext for creating a child logger with pre-populated fields.
// Example: reqLogger := logger.With().Str("request_id", id).Logger()
func (s *Service) With() LogContext {
	if s == nil || !s.isInitialized.Load() {
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
