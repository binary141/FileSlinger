package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// File mirrors the fields that can be set in ~/.config/fileSlinger/config.json.
// Pointer fields distinguish "not present" from zero values.
type File struct {
	Port     *int    `json:"port"`
	Dir      *string `json:"dir"`
	MaxFiles *int    `json:"max_files"`
	Token    *string `json:"token"`
}

func Load() (File, error) {
	var cfg File

	home, err := os.UserHomeDir()
	if err != nil {
		return cfg, err
	}

	path := filepath.Join(home, ".config", "fileSlinger", "config.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
