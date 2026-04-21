package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadINI(t *testing.T) {
	t.Run("empty file", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "config")
		err := os.WriteFile(p, []byte(""), 0600)
		require.NoError(t, err)

		keys, err := loadINI(p)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{}, keys)
	})

	t.Run("comments only", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "config")
		content := `# comment
; another comment
  # indented comment
`
		err := os.WriteFile(p, []byte(content), 0600)
		require.NoError(t, err)

		keys, err := loadINI(p)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{}, keys)
	})

	t.Run("basic key-value pairs", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "config")
		content := `output = json
api_url = https://example.com
`
		err := os.WriteFile(p, []byte(content), 0600)
		require.NoError(t, err)

		keys, err := loadINI(p)
		require.NoError(t, err)
		assert.Equal(t, "json", keys["output"])
		assert.Equal(t, "https://example.com", keys["api_url"])
	})

	t.Run("whitespace trimming", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "config")
		content := `output   =   json  
  api_url  =  https://example.com  
`
		err := os.WriteFile(p, []byte(content), 0600)
		require.NoError(t, err)

		keys, err := loadINI(p)
		require.NoError(t, err)
		assert.Equal(t, "json", keys["output"])
		assert.Equal(t, "https://example.com", keys["api_url"])
	})

	t.Run("section headers ignored", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "config")
		content := `[default]
output = json

[other]
api_url = https://example.com
`
		err := os.WriteFile(p, []byte(content), 0600)
		require.NoError(t, err)

		keys, err := loadINI(p)
		require.NoError(t, err)
		assert.Equal(t, "json", keys["output"])
		assert.Equal(t, "https://example.com", keys["api_url"])
	})

	t.Run("empty lines ignored", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "config")
		content := `output = json

api_url = https://example.com

`
		err := os.WriteFile(p, []byte(content), 0600)
		require.NoError(t, err)

		keys, err := loadINI(p)
		require.NoError(t, err)
		assert.Equal(t, "json", keys["output"])
		assert.Equal(t, "https://example.com", keys["api_url"])
	})

	t.Run("value with equals sign", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "config")
		content := `api_url = https://example.com?foo=bar&baz=qux
`
		err := os.WriteFile(p, []byte(content), 0600)
		require.NoError(t, err)

		keys, err := loadINI(p)
		require.NoError(t, err)
		assert.Equal(t, "https://example.com?foo=bar&baz=qux", keys["api_url"])
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := loadINI("/nonexistent/path/config")
		require.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("mixed comments and values", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "config")
		content := `# header comment
output = json
; inline style comment
api_url = https://example.com
# footer
`
		err := os.WriteFile(p, []byte(content), 0600)
		require.NoError(t, err)

		keys, err := loadINI(p)
		require.NoError(t, err)
		assert.Equal(t, "json", keys["output"])
		assert.Equal(t, "https://example.com", keys["api_url"])
	})

	t.Run("invalid lines skipped", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "config")
		content := `output = json
this is not a valid line
api_url = https://example.com
`
		err := os.WriteFile(p, []byte(content), 0600)
		require.NoError(t, err)

		keys, err := loadINI(p)
		require.NoError(t, err)
		assert.Equal(t, "json", keys["output"])
		assert.Equal(t, "https://example.com", keys["api_url"])
		assert.Len(t, keys, 2)
	})
}

func TestConfig_Load(t *testing.T) {
	t.Run("no config file returns defaults", func(t *testing.T) {
		dir := t.TempDir()
		cfg, err := Load(WithConfigDir(dir))
		require.NoError(t, err)
		assert.Empty(t, cfg.APIKey)
		assert.Empty(t, cfg.Output)
		assert.Empty(t, cfg.APIURL)
		assert.False(t, cfg.TestAPI)
		assert.Equal(t, defaultAliasesCache, cfg.AliasesCache)
		assert.Equal(t, defaultLocationsCache, cfg.LocationsCache)
		assert.Equal(t, defaultRegionsCache, cfg.RegionsCache)
	})

	t.Run("with config file", func(t *testing.T) {
		dir := t.TempDir()

		cfgContent := `output = table
api_url = https://test.example.com
test_api = true
aliases_cache = 2h
locations_cache = 24h
regions_cache = 48h
`
		err := os.WriteFile(filepath.Join(dir, "config"), []byte(cfgContent), 0600)
		require.NoError(t, err)

		cfg, err := Load(WithConfigDir(dir))
		require.NoError(t, err)
		assert.Equal(t, "table", cfg.Output)
		assert.Equal(t, "https://test.example.com", cfg.APIURL)
		assert.True(t, cfg.TestAPI)
		assert.Equal(t, 2*time.Hour, cfg.AliasesCache)
		assert.Equal(t, 24*time.Hour, cfg.LocationsCache)
		assert.Equal(t, 48*time.Hour, cfg.RegionsCache)
	})

	t.Run("credentials file", func(t *testing.T) {
		dir := t.TempDir()

		//nolint:gosec // test fixture
		credContent := `api_key = my-secret-key
`
		err := os.WriteFile(filepath.Join(dir, "credentials"), []byte(credContent), 0600)
		require.NoError(t, err)

		cfg, err := Load(WithConfigDir(dir))
		require.NoError(t, err)
		assert.Equal(t, "my-secret-key", cfg.APIKey)
	})

	t.Run("test_api case insensitive", func(t *testing.T) {
		dir := t.TempDir()

		cfgContent := `test_api = TRUE
`
		err := os.WriteFile(filepath.Join(dir, "config"), []byte(cfgContent), 0600)
		require.NoError(t, err)

		cfg, err := Load(WithConfigDir(dir))
		require.NoError(t, err)
		assert.True(t, cfg.TestAPI)
	})

	t.Run("invalid duration returns error", func(t *testing.T) {
		dir := t.TempDir()

		cfgContent := `aliases_cache = not-a-duration
`
		err := os.WriteFile(filepath.Join(dir, "config"), []byte(cfgContent), 0600)
		require.NoError(t, err)

		_, err = Load(WithConfigDir(dir))
		assert.Error(t, err)
	})

	t.Run("config file read error", func(t *testing.T) {
		dir := t.TempDir()

		err := os.WriteFile(filepath.Join(dir, "config"), []byte("output = json"), 0000)
		require.NoError(t, err)

		_, err = Load(WithConfigDir(dir))
		assert.Error(t, err)
	})

	t.Run("env vars override config file", func(t *testing.T) {
		dir := t.TempDir()

		cfgContent := `output = json
api_url = https://file.example.com
test_api = false
aliases_cache = 1h
locations_cache = 1h
regions_cache = 1h
`
		err := os.WriteFile(filepath.Join(dir, "config"), []byte(cfgContent), 0600)
		require.NoError(t, err)

		t.Setenv("DATAPACKET_OUTPUT", "csv")
		t.Setenv("DATAPACKET_API_URL", "https://env.example.com")
		t.Setenv("DATAPACKET_TEST_API", "true")
		t.Setenv("DATAPACKET_ALIASES_CACHE", "3h")
		t.Setenv("DATAPACKET_LOCATIONS_CACHE", "5h")
		t.Setenv("DATAPACKET_REGIONS_CACHE", "7h")

		cfg, err := Load(WithConfigDir(dir))
		require.NoError(t, err)
		assert.Equal(t, "csv", cfg.Output)
		assert.Equal(t, "https://env.example.com", cfg.APIURL)
		assert.True(t, cfg.TestAPI)
		assert.Equal(t, 3*time.Hour, cfg.AliasesCache)
		assert.Equal(t, 5*time.Hour, cfg.LocationsCache)
		assert.Equal(t, 7*time.Hour, cfg.RegionsCache)
	})

	t.Run("env var invalid duration returns error", func(t *testing.T) {
		dir := t.TempDir()

		t.Setenv("DATAPACKET_ALIASES_CACHE", "not-a-duration")

		_, err := Load(WithConfigDir(dir))
		assert.Error(t, err)
	})

	t.Run("env var overrides credentials file for api_key", func(t *testing.T) {
		dir := t.TempDir()

		//nolint:gosec // test fixture
		credContent := `api_key = file-key
`
		err := os.WriteFile(filepath.Join(dir, "credentials"), []byte(credContent), 0600)
		require.NoError(t, err)

		t.Setenv("DATAPACKET_API_KEY", "env-key")

		cfg, err := Load(WithConfigDir(dir))
		require.NoError(t, err)
		assert.Equal(t, "env-key", cfg.APIKey)
	})

	t.Run("credentials file used when no env var", func(t *testing.T) {
		dir := t.TempDir()

		//nolint:gosec // test fixture
		credContent := `api_key = file-key
`
		err := os.WriteFile(filepath.Join(dir, "credentials"), []byte(credContent), 0600)
		require.NoError(t, err)

		cfg, err := Load(WithConfigDir(dir))
		require.NoError(t, err)
		assert.Equal(t, "file-key", cfg.APIKey)
	})
}
