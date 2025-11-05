package logging

import (
	"github.com/Station-Manager/config"
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLoggingDir(t *testing.T) {
	t.Run("default logging dir", func(t *testing.T) {
		logger := &Service{
			AppConfig:     &config.Service{},
			LoggingConfig: &types.LoggingConfig{},
		}

		require.NoError(t, logger.Initialize())
	})
}
