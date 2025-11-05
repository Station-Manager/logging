package logging

import (
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
)

func (s *Service) initializeRollingFileLogger(exeName string) *lumberjack.Logger {
	if exeName == emptyString {
		exeName = "app"
	}

	path := filepath.Join(s.WorkingDir, s.LoggingConfig.RelLogFileDir, exeName+".log")

	return &lumberjack.Logger{
		Filename:   path,
		MaxBackups: s.LoggingConfig.LogFileMaxBackups,
		MaxAge:     s.LoggingConfig.LogFileMaxAgeDays,
		MaxSize:    s.LoggingConfig.LogFileMaxSizeMB,
	}
}

func (s *Service) initializeWriters(logfile string) []io.Writer {
	var writers []io.Writer

	// If both writers are disabled, enable the file writer
	if !s.LoggingConfig.ConsoleLogging && !s.LoggingConfig.FileLogging {
		s.LoggingConfig.FileLogging = true
	}
	if s.LoggingConfig.FileLogging {
		s.fileWriter = s.initializeRollingFileLogger(logfile)
		writers = append(writers, s.fileWriter)
	}
	if s.LoggingConfig.ConsoleLogging {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return writers
}
