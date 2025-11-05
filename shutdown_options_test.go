package logging

import (
	"testing"
	"time"

	"github.com/Station-Manager/config"
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cfgWithDefaults() *types.LoggingConfig {
	return &types.LoggingConfig{
		Level:                  "debug",
		WithTimestamp:          true,
		ConsoleLogging:         true,
		FileLogging:            false,
		RelLogFileDir:          ".",
		LogFileMaxBackups:      1,
		LogFileMaxAgeDays:      1,
		LogFileMaxSizeMB:       10,
		ShutdownTimeoutMS:      20,
		ShutdownTimeoutWarning: true,
	}
}

func newCfgService(cfg *types.LoggingConfig) *config.Service {
	svc := &config.Service{AppConfig: types.AppConfig{LoggingConfig: *cfg}}
	_ = svc.Initialize()
	return svc
}

// Verifies Close() waits up to timeout and returns without hanging when an event is never sent.
func TestCloseTimeoutWaitGroup(t *testing.T) {
	tmp := t.TempDir()
	cfg := cfgWithDefaults()
	cfg.ConsoleLogging = false
	cfg.FileLogging = true

	svc := &Service{WorkingDir: tmp, AppConfig: newCfgService(cfg)}
	require.NoError(t, svc.Initialize())

	// Start an event and never call Msg/Send to keep wg non-zero
	_ = svc.InfoWith()

	start := time.Now()
	require.NoError(t, svc.Close())
	elapsed := time.Since(start)
	assert.GreaterOrEqual(t, int64(elapsed/time.Millisecond), int64(cfg.ShutdownTimeoutMS))
}

// Verifies writer options (compression and console formatting) are plumbed.
func TestWriterOptions(t *testing.T) {
	tmp := t.TempDir()
	cfg := cfgWithDefaults()
	cfg.ConsoleLogging = true
	cfg.FileLogging = true
	cfg.LogFileCompress = true
	cfg.ConsoleNoColor = true
	cfg.ConsoleTimeFormat = time.RFC3339

	svc := &Service{WorkingDir: tmp, AppConfig: newCfgService(cfg)}
	require.NoError(t, svc.Initialize())

	defer svc.Close()

	if svc.fileWriter == nil {
		t.Fatalf("fileWriter must be initialized")
	}
	assert.True(t, svc.fileWriter.Compress)

	// Emit a log to ensure console writer configured doesn't panic
	svc.InfoWith().Msg("hello world")
}

// Verifies RelLogFileDir safety validation rejects absolute path.
func TestRelLogFileDirSafety(t *testing.T) {
	tmp := t.TempDir()
	cfg := cfgWithDefaults()
	cfg.RelLogFileDir = "/not/relative"

	svc := &Service{WorkingDir: tmp, AppConfig: newCfgService(cfg)}
	err := svc.Initialize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "RelLogFileDir")
}

// Basic race-ish scenario: concurrently build scoped loggers while closing.
func TestConcurrentWithDuringClose(t *testing.T) {
	tmp := t.TempDir()
	cfg := cfgWithDefaults()
	cfg.ShutdownTimeoutMS = 50

	svc := &Service{WorkingDir: tmp, AppConfig: newCfgService(cfg)}
	require.NoError(t, svc.Initialize())

	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			_ = svc.With().Str("i", "x").Logger().InfoWith()
		}
		close(done)
	}()

	_ = svc.Close()
	<-done
}
