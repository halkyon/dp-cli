package config

import (
	"bufio"
	"os"
	"strings"
	"time"
)

const (
	defaultConfigDir      = "/.config/dp"
	defaultAliasesCache   = 1 * time.Hour
	defaultLocationsCache = 7 * 24 * time.Hour
	defaultRegionsCache   = 7 * 24 * time.Hour
)

type Config struct {
	configDir      string
	APIKey         string
	Output         string
	APIURL         string
	TestAPI        bool
	AliasesCache   time.Duration
	LocationsCache time.Duration
	RegionsCache   time.Duration
}

type ConfigOption func(*Config)

func WithConfigDir(dir string) ConfigOption {
	return func(c *Config) {
		c.configDir = dir
	}
}

func Load(opts ...ConfigOption) (*Config, error) {
	cfg := new(Config)
	for _, opt := range opts {
		opt(cfg)
	}

	configPath, credsPath := cfg.configPaths()

	var data map[string]string
	if configPath != "" {
		var err error
		data, err = loadINI(configPath)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}

	if envOutput := os.Getenv("DATAPACKET_OUTPUT"); envOutput != "" {
		cfg.Output = envOutput
	} else if data != nil && data["output"] != "" {
		cfg.Output = data["output"]
	}

	if envAPIURL := os.Getenv("DATAPACKET_API_URL"); envAPIURL != "" {
		cfg.APIURL = envAPIURL
	} else if data != nil && data["api_url"] != "" {
		cfg.APIURL = data["api_url"]
	}

	if envTestAPI := os.Getenv("DATAPACKET_TEST_API"); envTestAPI != "" {
		cfg.TestAPI = strings.EqualFold(envTestAPI, "true")
	} else if data != nil {
		cfg.TestAPI = strings.EqualFold(data["test_api"], "true")
	}

	if envAliases := os.Getenv("DATAPACKET_ALIASES_CACHE"); envAliases != "" {
		d, err := time.ParseDuration(envAliases)
		if err != nil {
			return nil, err
		}
		cfg.AliasesCache = d
	} else if data != nil && data["aliases_cache"] != "" {
		d, err := time.ParseDuration(data["aliases_cache"])
		if err != nil {
			return nil, err
		}
		cfg.AliasesCache = d
	} else {
		cfg.AliasesCache = defaultAliasesCache
	}

	if envLocations := os.Getenv("DATAPACKET_LOCATIONS_CACHE"); envLocations != "" {
		d, err := time.ParseDuration(envLocations)
		if err != nil {
			return nil, err
		}
		cfg.LocationsCache = d
	} else if data != nil && data["locations_cache"] != "" {
		d, err := time.ParseDuration(data["locations_cache"])
		if err != nil {
			return nil, err
		}
		cfg.LocationsCache = d
	} else {
		cfg.LocationsCache = defaultLocationsCache
	}

	if envRegions := os.Getenv("DATAPACKET_REGIONS_CACHE"); envRegions != "" {
		d, err := time.ParseDuration(envRegions)
		if err != nil {
			return nil, err
		}
		cfg.RegionsCache = d
	} else if data != nil && data["regions_cache"] != "" {
		d, err := time.ParseDuration(data["regions_cache"])
		if err != nil {
			return nil, err
		}
		cfg.RegionsCache = d
	} else {
		cfg.RegionsCache = defaultRegionsCache
	}

	if credsPath != "" {
		credsData, err := loadINI(credsPath)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if credsData != nil {
			if cfg.APIKey == "" {
				cfg.APIKey = credsData["api_key"]
			}
		}
	}

	if envAPIKey := os.Getenv("DATAPACKET_API_KEY"); envAPIKey != "" {
		cfg.APIKey = envAPIKey
	}

	return cfg, nil
}

func (c *Config) configPaths() (config, credentials string) {
	if c.configDir != "" {
		return c.configDir + "/config", c.configDir + "/credentials"
	}
	home := os.Getenv("HOME")
	if home == "" {
		return "", ""
	}
	return home + defaultConfigDir + "/config", home + defaultConfigDir + "/credentials"
}

func loadINI(path string) (keys map[string]string, err error) {
	f, openErr := os.Open(path) //nolint:gosec // path is from internal function
	if openErr != nil {
		return nil, openErr
	}
	defer func() {
		closeErr := f.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	keys = make(map[string]string)
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "[") {
			continue
		}
		if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			keys[k] = v
		}
	}
	return keys, s.Err()
}
