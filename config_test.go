package main

import (
	"os"
	"testing"
	"time"
)

// unsetEnv removes key for the duration of the test. t.Setenv is used first so
// that the testing package restores the original value during cleanup.
func unsetEnv(t *testing.T, key string) {
	t.Helper()

	t.Setenv(key, "")
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unsetting %s: %v", key, err)
	}
}

func TestNewConfigDefaults(t *testing.T) {
	t.Setenv("HETZNER_TOKEN", "some-token")
	unsetEnv(t, "SCRAPE_INTERVAL")

	config, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() returned error: %v", err)
	}

	if config.Token != "some-token" {
		t.Errorf("Token = %q, want %q", config.Token, "some-token")
	}
	if want := 30 * time.Minute; config.ScrapeInterval != want {
		t.Errorf("ScrapeInterval = %v, want %v", config.ScrapeInterval, want)
	}
}

func TestNewConfigEnvOverridesDefault(t *testing.T) {
	t.Setenv("HETZNER_TOKEN", "some-token")
	t.Setenv("SCRAPE_INTERVAL", "90s")

	config, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() returned error: %v", err)
	}

	if want := 90 * time.Second; config.ScrapeInterval != want {
		t.Errorf("ScrapeInterval = %v, want %v", config.ScrapeInterval, want)
	}
}

// A field without a default tag is required, and an unset variable is an error.
func TestNewConfigMissingRequiredEnv(t *testing.T) {
	unsetEnv(t, "HETZNER_TOKEN")
	t.Setenv("SCRAPE_INTERVAL", "30m")

	config, err := NewConfig()
	if err == nil {
		t.Fatalf("NewConfig() = %+v, want error for unset HETZNER_TOKEN", config)
	}
}

// An explicitly empty variable counts as set, so it is accepted verbatim
// instead of falling back to the default or erroring out.
func TestNewConfigEmptyEnvIsUsedVerbatim(t *testing.T) {
	t.Setenv("HETZNER_TOKEN", "")
	t.Setenv("SCRAPE_INTERVAL", "1h")

	config, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() returned error: %v", err)
	}

	if config.Token != "" {
		t.Errorf("Token = %q, want empty string", config.Token)
	}
}

func TestNewConfigInvalidDuration(t *testing.T) {
	t.Setenv("HETZNER_TOKEN", "some-token")
	t.Setenv("SCRAPE_INTERVAL", "every-monday")

	config, err := NewConfig()
	if err == nil {
		t.Fatalf("NewConfig() = %+v, want error for unparseable SCRAPE_INTERVAL", config)
	}
}
