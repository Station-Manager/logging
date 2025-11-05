package logging

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/7Q-Station-Manager/config"
	"github.com/7Q-Station-Manager/utils"
	"github.com/stretchr/testify/require"
)

// helper to create a ready-to-use file-based logger in a temp dir using config.Service
func newFileLogger(t testing.TB, level string) (*Service, string) {
	t.Helper()
	wd := t.TempDir()
	// Bind utils.WorkingDir() to this test directory
	require.NoError(t, os.Setenv(utils.EnvSmWorkingDir, wd))
	t.Cleanup(func() { _ = os.Unsetenv(utils.EnvSmWorkingDir) })

	// Prepare a config.json in the working dir with desired logging settings
	cfgPath := filepath.Join(wd, "config.json")
	data := []byte("{\n  \"LoggingConfig\": {\n    \"Level\": \"" + level + "\",\n    \"SkipFrameCount\": 0,\n    \"WithTimestamp\": false,\n    \"ConsoleLogging\": false,\n    \"FileLogging\": true,\n    \"RelLogFileDir\": \"logs\",\n    \"LogFileMaxBackups\": 1,\n    \"LogFileMaxAgeDays\": 1,\n    \"LogFileMaxSizeMB\": 5\n  }\n}")
	require.NoError(t, os.WriteFile(cfgPath, data, 0644))

	appCfg := &config.Service{WorkDir: wd}
	require.NoError(t, appCfg.Initialize())

	l := NewLogger()
	l.WorkingDir = wd
	l.AppConfig = appCfg
	require.NoError(t, l.Initialize())
	return l, filepath.Join(wd, "logs", "logging.log")
}

func TestInitializeErrors(t *testing.T) {
	// No working dir
	{
		wd := t.TempDir()
		require.NoError(t, os.Setenv(utils.EnvSmWorkingDir, wd))
		t.Cleanup(func() { _ = os.Unsetenv(utils.EnvSmWorkingDir) })
		cfgPath := filepath.Join(wd, "config.json")
		data := []byte("{\n  \"LoggingConfig\": {\n    \"ConsoleLogging\": true\n  }\n}")
		require.NoError(t, os.WriteFile(cfgPath, data, 0644))
		appCfg := &config.Service{}
		require.NoError(t, appCfg.Initialize())

		l := NewLogger()
		l.AppConfig = appCfg
		require.Error(t, l.Initialize())
	}

	// No config (AppConfig not set)
	{
		l := NewLogger()
		l.WorkingDir = t.TempDir()
		require.Error(t, l.Initialize())
	}

	// No channels enabled
	{
		wd := t.TempDir()
		require.NoError(t, os.Setenv(utils.EnvSmWorkingDir, wd))
		t.Cleanup(func() { _ = os.Unsetenv(utils.EnvSmWorkingDir) })
		cfgPath := filepath.Join(wd, "config.json")
		data := []byte("{\n  \"LoggingConfig\": {\n    \"ConsoleLogging\": false,\n    \"FileLogging\": false,\n    \"RelLogFileDir\": \"logs\",\n    \"Level\": \"debug\"\n  }\n}")
		require.NoError(t, os.WriteFile(cfgPath, data, 0644))
		appCfg := &config.Service{}
		require.NoError(t, appCfg.Initialize())

		l := NewLogger()
		l.WorkingDir = wd
		l.AppConfig = appCfg
		err := l.Initialize()
		require.Error(t, err)
		require.Contains(t, err.Error(), "no logging channels enabled")
	}

	// Invalid level
	{
		wd := t.TempDir()
		require.NoError(t, os.Setenv(utils.EnvSmWorkingDir, wd))
		t.Cleanup(func() { _ = os.Unsetenv(utils.EnvSmWorkingDir) })
		cfgPath := filepath.Join(wd, "config.json")
		data := []byte("{\n  \"LoggingConfig\": {\n    \"ConsoleLogging\": true,\n    \"Level\": \"notalevel\"\n  }\n}")
		require.NoError(t, os.WriteFile(cfgPath, data, 0644))
		appCfg := &config.Service{}
		require.NoError(t, appCfg.Initialize())

		l := NewLogger()
		l.WorkingDir = wd
		l.AppConfig = appCfg
		require.Error(t, l.Initialize())
	}
}

func TestFileLoggingCreatesAndWrites(t *testing.T) {
	l, logPath := newFileLogger(t, "debug")
	t.Cleanup(func() { _ = l.Close() })

	l.Infof("hello %s", "world")
	l.Warn("be careful")

	// Ensure file exists and contains the messages
	_, err := os.Stat(logPath)
	require.NoError(t, err)

	content, err := os.ReadFile(logPath)
	require.NoError(t, err)
	text := string(content)
	require.Contains(t, text, "hello world")
	require.Contains(t, text, "be careful")
}

func TestLevelFiltering(t *testing.T) {
	l, logPath := newFileLogger(t, "warn")
	t.Cleanup(func() { _ = l.Close() })

	l.Debug("debug msg")
	l.Info("info msg")
	l.Warn("warn msg")
	l.Error("error msg")

	// Only warn and above should be present
	f, err := os.Open(logPath)
	require.NoError(t, err)
	defer f.Close()

	r := bufio.NewReader(f)
	b, err := io.ReadAll(r)
	require.NoError(t, err)
	s := string(b)
	require.NotContains(t, s, "debug msg")
	require.NotContains(t, s, "info msg")
	require.Contains(t, s, "warn msg")
	require.Contains(t, s, "error msg")
}

func TestDumpOutputs(t *testing.T) {
	type person struct {
		Name string
		Age  int
	}
	l, logPath := newFileLogger(t, "debug")
	t.Cleanup(func() { _ = l.Close() })

	m := map[string]int{"a": 1, "b": 2}
	s := []string{"x", "y"}
	p := person{Name: "Ada", Age: 37}

	l.Dump(nil)
	l.Dump(m)
	l.Dump(s)
	l.Dump(p)
	l.Dump(&p)

	data, err := os.ReadFile(logPath)
	require.NoError(t, err)
	str := string(data)
	// spot-check that dump wrote something meaningful
	require.Contains(t, str, "<nil>")
	require.True(t, strings.Contains(str, "a") || strings.Contains(str, "b"))
	require.Contains(t, str, "Ada")
}

func TestConcurrentLogging(t *testing.T) {
	l, _ := newFileLogger(t, "debug")
	t.Cleanup(func() { _ = l.Close() })

	const goroutines = 100
	const iterations = 100

	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			for j := 0; j < iterations; j++ {
				l.Info("goroutine", id, "iteration", j)
				l.Debug("debug msg", id)
				l.Warn("warn msg", id)
				l.Error("error msg", id)
				l.Infof("formatted %d:%d", id, j)
			}
			done <- true
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
}

func TestConcurrentDump(t *testing.T) {
	l, _ := newFileLogger(t, "debug")
	t.Cleanup(func() { _ = l.Close() })

	type testStruct struct {
		Field1 string
		Field2 int
	}

	const goroutines = 50
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			data := testStruct{
				Field1: fmt.Sprintf("test-%d", id),
				Field2: id,
			}
			for j := 0; j < 10; j++ {
				l.Dump(data)
			}
			done <- true
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
}

func TestStructuredLogging(t *testing.T) {
	l, logPath := newFileLogger(t, "debug")
	t.Cleanup(func() { _ = l.Close() })

	// Test basic structured fields
	l.InfoWith().
		Str("user_id", "12345").
		Int("count", 42).
		Bool("active", true).
		Msg("User processed")

	// Test error logging with structured fields
	testErr := fmt.Errorf("test error")
	l.ErrorWith().
		Err(testErr).
		Str("operation", "database").
		Int("retry_count", 3).
		Msg("Operation failed")

	// Test float and uint
	l.DebugWith().
		Float64("temperature", 98.6).
		Uint("port", 8080).
		Msg("Metrics")

	// Read and verify log contains structured data
	data, err := os.ReadFile(logPath)
	require.NoError(t, err)
	str := string(data)

	require.Contains(t, str, `"user_id":"12345"`)
	require.Contains(t, str, `"count":42`)
	require.Contains(t, str, `"active":true`)
	require.Contains(t, str, `"error":"test error"`)
	require.Contains(t, str, `"operation":"database"`)
	require.Contains(t, str, `"retry_count":3`)
	require.Contains(t, str, `"temperature":98.6`)
	require.Contains(t, str, `"port":8080`)
}

func TestStructuredLoggingWithContext(t *testing.T) {
	l, logPath := newFileLogger(t, "debug")
	t.Cleanup(func() { _ = l.Close() })

	// Create a context logger with pre-populated fields
	reqLogger := l.With().
		Str("request_id", "req-123").
		Str("user_id", "user-456").
		Logger()

	// All logs from reqLogger will include request_id and user_id
	reqLogger.InfoWith().Str("action", "start").Msg("Request started")
	reqLogger.InfoWith().Str("action", "end").Int("status", 200).Msg("Request completed")

	// Verify both logs contain the context fields
	data, err := os.ReadFile(logPath)
	require.NoError(t, err)
	str := string(data)

	// Count occurrences - should appear twice (once per log)
	requestIDCount := strings.Count(str, `"request_id":"req-123"`)
	userIDCount := strings.Count(str, `"user_id":"user-456"`)

	require.Equal(t, 2, requestIDCount, "request_id should appear in both logs")
	require.Equal(t, 2, userIDCount, "user_id should appear in both logs")
	require.Contains(t, str, `"action":"start"`)
	require.Contains(t, str, `"action":"end"`)
	require.Contains(t, str, `"status":200`)
}

func TestStructuredLoggingArraysAndDuration(t *testing.T) {
	l, logPath := newFileLogger(t, "debug")
	t.Cleanup(func() { _ = l.Close() })

	l.InfoWith().
		Strs("tags", []string{"golang", "logging", "structured"}).
		Dur("elapsed", 250*1000000). // 250ms in nanoseconds
		Msg("Tagged operation")

	data, err := os.ReadFile(logPath)
	require.NoError(t, err)
	str := string(data)

	require.Contains(t, str, `"tags":["golang","logging","structured"]`)
	require.Contains(t, str, `"elapsed":250`)
}

func TestStructuredLoggingConcurrent(t *testing.T) {
	l, _ := newFileLogger(t, "debug")
	t.Cleanup(func() { _ = l.Close() })

	const goroutines = 100
	const iterations = 50

	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			for j := 0; j < iterations; j++ {
				l.InfoWith().
					Int("goroutine_id", id).
					Int("iteration", j).
					Str("status", "running").
					Msg("Concurrent log")
			}
			done <- true
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
}

func TestStructuredLoggingWithNesting(t *testing.T) {
	l, logPath := newFileLogger(t, "debug")
	t.Cleanup(func() { _ = l.Close() })

	l.InfoWith().
		Str("event", "user_action").
		Dict("user", func(e LogEvent) {
			e.Str("id", "user-123")
			e.Int("age", 30)
		}).
		Dict("metadata", func(e LogEvent) {
			e.Str("ip", "192.168.1.1")
			e.Bool("verified", true)
		}).
		Msg("Nested structured log")

	data, err := os.ReadFile(logPath)
	require.NoError(t, err)
	str := string(data)

	require.Contains(t, str, `"user":`)
	require.Contains(t, str, `"id":"user-123"`)
	require.Contains(t, str, `"age":30`)
	require.Contains(t, str, `"metadata":`)
	require.Contains(t, str, `"ip":"192.168.1.1"`)
	require.Contains(t, str, `"verified":true`)
}

func TestUninitializedLoggerDoesNotPanic(t *testing.T) {
	// Test that a logger created without NewLogger() doesn't panic
	// This simulates dependency injection scenarios where Service is created via struct literal

	l := &Service{}

	// None of these should panic
	l.Info("test")
	l.Infof("test %d", 1)
	l.Debug("test")
	l.Debugf("test %d", 1)
	l.Warn("test")
	l.Warnf("test %d", 1)
	l.Error("test")
	l.Errorf("test %d", 1)
	l.InfoWith().Str("key", "value").Msg("test")
	l.ErrorWith().Str("key", "value").Msg("test")

	// After initialize, mutex should be created and work properly
	wd := t.TempDir()
	require.NoError(t, os.Setenv(utils.EnvSmWorkingDir, wd))
	t.Cleanup(func() { _ = os.Unsetenv(utils.EnvSmWorkingDir) })

	cfgPath := filepath.Join(wd, "config.json")
	data := []byte(`{"LoggingConfig": {"Level": "info", "ConsoleLogging": false, "FileLogging": true, "RelLogFileDir": "logs"}}`)
	require.NoError(t, os.WriteFile(cfgPath, data, 0644))

	appCfg := &config.Service{WorkDir: wd}
	require.NoError(t, appCfg.Initialize())

	l.WorkingDir = wd
	l.AppConfig = appCfg
	require.NoError(t, l.Initialize())

	// Now logging should work
	l.Info("initialized")
	l.InfoWith().Str("status", "working").Msg("Initialized successfully")
}
