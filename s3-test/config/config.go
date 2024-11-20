package config

import (
	"encoding/json"
	"os"
)

type (
	Config struct {
		Storages []string `json:"storages"`
	}

	Storage struct {
		Name []string
	}
)

func NewConfig(path string) (*Config, error) {
	cfg := &Config{}
	byteValue, _ := os.ReadFile(path)
	err := json.Unmarshal(byteValue, &cfg)
	return cfg, err
}
