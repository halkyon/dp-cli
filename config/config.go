package config

import (
	"os"

	"gopkg.in/ini.v1"
)

type Config struct {
	APIKey string
}

func getConfigPath() string {
	home := os.Getenv("HOME")
	if home == "" {
		return ""
	}
	return home + "/.config/dp/credentials"
}

func Load() (*Config, error) {
	var cfg Config

	path := getConfigPath()
	if path == "" {
		return &cfg, nil
	}

	data, err := ini.Load(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &cfg, nil
		}
		return nil, err
	}

	cfg.APIKey = data.Section("").Key("api_key").String()

	return &cfg, nil
}

func GetAPIKey() (string, error) {
	if key := os.Getenv("DATAPACKET_API_KEY"); key != "" {
		return key, nil
	}

	cfg, err := Load()
	if err != nil {
		return "", err
	}

	if cfg.APIKey == "" {
		return "", nil
	}

	return cfg.APIKey, nil
}
