package logging

import (
	"fmt"
	"testing"
)

func BenchmarkStructuredLogging(b *testing.B) {
	l, _ := newFileLogger(b, "info")
	b.Cleanup(func() { _ = l.Close() })

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			l.InfoWith().
				Str("user_id", "user-123").
				Int("count", i).
				Str("operation", "test").
				Msg("Benchmark log")
			i++
		}
	})
}

func BenchmarkStructuredLoggingWithError(b *testing.B) {
	l, _ := newFileLogger(b, "info")
	b.Cleanup(func() { _ = l.Close() })

	err := fmt.Errorf("test error")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			l.ErrorWith().
				Err(err).
				Str("operation", "benchmark").
				Int("retry", i).
				Msg("Error occurred")
			i++
		}
	})
}

func BenchmarkLegacyLogging(b *testing.B) {
	l, _ := newFileLogger(b, "info")
	b.Cleanup(func() { _ = l.Close() })

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			l.Infof("Benchmark log user_id=%s count=%d operation=%s", "user-123", i, "test")
			i++
		}
	})
}

func BenchmarkContextLoggerCreation(b *testing.B) {
	l, _ := newFileLogger(b, "info")
	b.Cleanup(func() { _ = l.Close() })

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reqLogger := l.With().
			Str("request_id", fmt.Sprintf("req-%d", i)).
			Str("user_id", "user-123").
			Logger()
		reqLogger.InfoWith().Str("action", "start").Msg("Request started")
	}
}

func BenchmarkHighConcurrency(b *testing.B) {
	l, _ := newFileLogger(b, "info")
	b.Cleanup(func() { _ = l.Close() })

	b.ResetTimer()
	b.SetParallelism(100) // 100 goroutines
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			l.InfoWith().
				Int("goroutine_id", i).
				Str("data", "benchmark").
				Msg("High concurrency test")
			i++
		}
	})
}
