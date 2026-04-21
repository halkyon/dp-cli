package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Load(t *testing.T) {
	t.Run("Empty config", func(t *testing.T) {
		// Assuming no config file exists in the environment where tests run
		cfg, err := Load()
		require.NoError(t, err)
		assert.Empty(t, cfg.APIKey)
	})

	t.Run("Environment variable", func(t *testing.T) {
		t.Setenv("DATAPACKET_API_KEY", "test-key")

		key, err := GetAPIKey()
		require.NoError(t, err)
		assert.Equal(t, "test-key", key)
	})

	t.Run("GetAPIKey error", func(t *testing.T) {
		t.Setenv("DATAPACKET_API_KEY", "")
		key, err := GetAPIKey()
		// If no config file, it returns "" and nil error.
		require.NoError(t, err)
		assert.Empty(t, key)
	})
}
