package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Load(t *testing.T) {
	t.Run("Empty config", func(t *testing.T) {
		// Assuming no config file exists in the environment where tests run
		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "", cfg.APIKey)
		assert.Equal(t, "", cfg.Output)
	})

	t.Run("Environment variable", func(t *testing.T) {
		t.Setenv("DATAPACKET_API_KEY", "test-key")

		key, err := GetAPIKey()
		assert.NoError(t, err)
		assert.Equal(t, "test-key", key)
	})

	t.Run("GetAPIKey error", func(t *testing.T) {
		t.Setenv("DATAPACKET_API_KEY", "")
		key, err := GetAPIKey()
		// If no config file, it returns "" and nil error.
		assert.NoError(t, err)
		assert.Equal(t, "", key)
	})
}
