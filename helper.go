package logging

import "github.com/rs/zerolog"

func getLevel(level string) (zerolog.Level, error) {
	l, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.DebugLevel, err
	}
	return l, nil
}

func logEventBuilder(s *Service, level zerolog.Level) LogEvent {
	if s == nil || !s.isInitialized.Load() {
		return newLogEvent(nil)
	}
	if level == zerolog.NoLevel {
		return newLogEvent(nil)
	}

	// Acquire read lock to prevent Close() from running during log creation
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring lock
	if !s.isInitialized.Load() {
		return newLogEvent(nil)
	}

	logger := s.logger.Load()
	if logger == nil {
		return newLogEvent(nil)
	}

	if logger.GetLevel() > level {
		return newLogEvent(nil) // Return early if level is not enabled
	}

	switch level {
	case zerolog.DebugLevel:
		return newLogEvent(logger.Debug())
	case zerolog.InfoLevel:
		return newLogEvent(logger.Info())
	case zerolog.WarnLevel:
		return newLogEvent(logger.Warn())
	case zerolog.ErrorLevel:
		return newLogEvent(logger.Error())
	case zerolog.FatalLevel:
		return newLogEvent(logger.Fatal())
	case zerolog.PanicLevel:
		return newLogEvent(logger.Panic())
	//case zerolog.Disabled:
	//	return newLogEvent(nil)
	case zerolog.TraceLevel:
		return newLogEvent(logger.Trace())
	default:
		return newLogEvent(nil)
	}
}
