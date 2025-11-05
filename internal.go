package logging

import (
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"path/filepath"
)

func (l *Service) initializeRollingFileLogger(exeName string) *lumberjack.Logger {
	if exeName == emptyString {
		exeName = "app"
	}

	path := filepath.Join(l.WorkingDir, l.LoggingConfig.RelLogFileDir, exeName+".log")
	fmt.Println("Initializing rolling file logger for:", path)
	return &lumberjack.Logger{
		Filename:   path,
		MaxBackups: l.LoggingConfig.LogFileMaxBackups,
		MaxAge:     l.LoggingConfig.LogFileMaxAgeDays,
		MaxSize:    l.LoggingConfig.LogFileMaxSizeMB,
	}
}
