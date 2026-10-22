package config

import (
	"log/slog"
	"slices"
	"strings"
)

// Option is a function that modifies Config.
// This is the functional options pattern.
type Option func(*Config)

// OptLogFormat sets the log format.
func OptLogFormat(s string) Option {
	return func(c *Config) {
		if isValidEnum("Log Format", s, []string{"json", "text", "tint"}) {
			c.Log.Format = s
		}
	}
}

// OptLogLevel sets the log level.
func OptLogLevel(s string) Option {
	return func(c *Config) {
		if isValidEnum("Log Level", s, []string{"debug", "info", "warn", "error"}) {
			c.Log.Level = s
		}
	}
}

// OptLogDestination sets the log destination.
func OptLogDestination(s string) Option {
	return func(c *Config) {
		if isValidEnum("Log Destination", s, []string{"file", "stdout", "stderr"}) {
			c.Log.Destination = s
		}
	}
}

// OptGnamesURL sets the gnames REST API URL.
func OptGnamesURL(s string) Option {
	s = strings.TrimSpace(s)
	return func(c *Config) {
		if isValidString("Gnames URL", s) {
			c.GnamesURL = s
		}
	}
}

// OptGapChars sets the gap in runes that splits groups.
func OptGapChars(i int) Option {
	return func(c *Config) {
		if isValidInt("Gap Chars", i) {
			c.GapChars = i
		}
	}
}

// OptSpanChars sets the largest number of runes in a group.
func OptSpanChars(i int) Option {
	return func(c *Config) {
		if isValidInt("Span Chars", i) {
			c.SpanChars = i
		}
	}
}

// OptHintWindow sets how many runes back genus hints come from.
func OptHintWindow(i int) Option {
	return func(c *Config) {
		if isValidInt("Hint Window", i) {
			c.HintWindow = i
		}
	}
}

// OptHintsLimit sets the largest number of genus hints per name.
func OptHintsLimit(i int) Option {
	return func(c *Config) {
		if isValidInt("Hints Limit", i) {
			c.HintsLimit = i
		}
	}
}

// OptGeneraLimit sets the largest number of candidate genera per group.
func OptGeneraLimit(i int) Option {
	return func(c *Config) {
		if isValidInt("Genera Limit", i) {
			c.GeneraLimit = i
		}
	}
}

// OptThreshold sets the lowest score of an accepted genus.
func OptThreshold(f float64) Option {
	return func(c *Config) {
		c.Threshold = f
	}
}

// OptWithFeatures asks abbr_names for raw score features.
// Runtime-only field - not in ToOptions().
func OptWithFeatures(b bool) Option {
	return func(c *Config) {
		c.WithFeatures = b
	}
}

// OptWithDetails adds raw abbr_names input and output to the groups.
// Runtime-only field - not in ToOptions().
func OptWithDetails(b bool) Option {
	return func(c *Config) {
		c.WithDetails = b
	}
}

// OptHomeDir sets the home directory for config, cache, and log locations.
// Set once at startup from os.UserHomeDir().
// Runtime-only field - not in ToOptions().
func OptHomeDir(s string) Option {
	s = strings.TrimSpace(s)
	return func(c *Config) {
		if isValidString("Home Directory", s) {
			c.HomeDir = s
		}
	}
}

// Validation helpers

func isValidString(name, value string) bool {
	if value == "" {
		slog.Warn("Option rejected: value cannot be empty", "option", name)
		return false
	}
	return true
}

func isValidInt(name string, value int) bool {
	if value < 0 {
		slog.Warn("Option rejected: value must be non-negative",
			"option", name, "value", value)
		return false
	}
	return true
}

func isValidEnum(name, value string, allowed []string) bool {
	if slices.Contains(allowed, value) {
		return true
	}
	slog.Warn("Option rejected: value is not allowed",
		"option", name, "value", value, "allowed", allowed)
	return false
}
