package logging

import "github.com/rs/zerolog"

func getLevel(level string) (zerolog.Level, error) {
	l, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.DebugLevel, err
	}
	return l, nil
}

// logEventBuilder creates a log event for the given level.
// It uses reference counting to ensure the logger remains valid for the duration
// of the logging operation, preventing race conditions with Close().
func logEventBuilder(s *Service, level zerolog.Level) LogEvent {
	if s == nil || !s.isInitialized.Load() {
		return newLogEvent(nil)
	}
	if level == zerolog.NoLevel {
		return newLogEvent(nil)
	}

	// Increment active operations counter before acquiring lock
	s.activeOps.Add(1)

	// Acquire read lock to prevent Close() from running during log creation
	s.mu.RLock()

	// Double-check after acquiring lock
	if !s.isInitialized.Load() {
		s.mu.RUnlock()
		s.activeOps.Add(-1)
		return newLogEvent(nil)
	}

	logger := s.logger.Load()
	if logger == nil {
		s.mu.RUnlock()
		s.activeOps.Add(-1)
		return newLogEvent(nil)
	}

	if logger.GetLevel() > level {
		s.mu.RUnlock()
		s.activeOps.Add(-1)
		return newLogEvent(nil) // Return early if level is not enabled
	}

	var event *zerolog.Event
	switch level {
	case zerolog.DebugLevel:
		event = logger.Debug()
	case zerolog.InfoLevel:
		event = logger.Info()
	case zerolog.WarnLevel:
		event = logger.Warn()
	case zerolog.ErrorLevel:
		event = logger.Error()
	case zerolog.FatalLevel:
		event = logger.Fatal()
	case zerolog.PanicLevel:
		event = logger.Panic()
	case zerolog.TraceLevel:
		event = logger.Trace()
	default:
		s.mu.RUnlock()
		s.activeOps.Add(-1)
		return newLogEvent(nil)
	}

	s.mu.RUnlock()

	// Wrap the event to decrement counter when done
	return newTrackedLogEvent(event, s)
}
