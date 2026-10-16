package config

import "path/filepath"

// ConfigDir returns the path to the configuration directory.
func ConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".config", "gnabbr")
}

// CacheDir returns the path to the cache directory.
func CacheDir(homeDir string) string {
	return filepath.Join(homeDir, ".cache", "gnabbr")
}

// LogDir returns the path to the log directory.
func LogDir(homeDir string) string {
	return filepath.Join(homeDir, ".local", "share", "gnabbr", "logs")
}

// ConfigFilePath returns the full path to config.yaml.
func ConfigFilePath(homeDir string) string {
	return filepath.Join(ConfigDir(homeDir), "config.yaml")
}
