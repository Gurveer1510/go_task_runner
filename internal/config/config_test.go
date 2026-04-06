package config

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestLoadConfigUsesDefaults(t *testing.T) {
	resetConfigTestState(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.AppEnv != "development" {
		t.Fatalf("expected default app env development, got %q", cfg.AppEnv)
	}
	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %q", cfg.Port)
	}
	if cfg.Concurrency != 5 {
		t.Fatalf("expected default concurrency 5, got %d", cfg.Concurrency)
	}
	if cfg.BaseDelay != 2*time.Second {
		t.Fatalf("expected default base delay 2s, got %v", cfg.BaseDelay)
	}
	if cfg.LeaseDuration != 30*time.Second {
		t.Fatalf("expected default lease duration 30s, got %v", cfg.LeaseDuration)
	}
}

func TestLoadConfigReadsEnvironmentValues(t *testing.T) {
	resetConfigTestState(t)

	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("PORT", "9090")
	t.Setenv("WORKER_CONCURRENCY", "8")
	t.Setenv("WORKER_BASE_DELAY_MS", "500")
	t.Setenv("WORKER_LEASE_MS", "45000")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.AppEnv != "production" {
		t.Fatalf("expected production app env, got %q", cfg.AppEnv)
	}
	if cfg.DBUrl != "postgres://test" {
		t.Fatalf("expected database url from env, got %q", cfg.DBUrl)
	}
	if cfg.Port != "9090" {
		t.Fatalf("expected port 9090, got %q", cfg.Port)
	}
	if cfg.Concurrency != 8 {
		t.Fatalf("expected concurrency 8, got %d", cfg.Concurrency)
	}
	if cfg.BaseDelay != 500*time.Millisecond {
		t.Fatalf("expected base delay 500ms, got %v", cfg.BaseDelay)
	}
	if cfg.LeaseDuration != 45*time.Second {
		t.Fatalf("expected lease duration 45s, got %v", cfg.LeaseDuration)
	}
}

func TestLoadConfigRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		wantError string
	}{
		{
			name:      "invalid concurrency",
			key:       "WORKER_CONCURRENCY",
			value:     "0",
			wantError: "WORKER_CONCURRENCY must be greater than 0",
		},
		{
			name:      "invalid base delay",
			key:       "WORKER_BASE_DELAY_MS",
			value:     "0",
			wantError: "WORKER_BASE_DELAY_MS must be greater than 0",
		},
		{
			name:      "invalid lease duration",
			key:       "WORKER_LEASE_MS",
			value:     "0",
			wantError: "WORKER_LEASE_MS must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetConfigTestState(t)
			t.Setenv(tt.key, tt.value)

			_, err := LoadConfig()
			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.wantError)
			}
			if err.Error() != tt.wantError {
				t.Fatalf("expected error %q, got %q", tt.wantError, err.Error())
			}
		})
	}
}

func resetConfigTestState(t *testing.T) {
	t.Helper()

	viper.Reset()
	t.Cleanup(viper.Reset)

	keys := []string{
		"APP_ENV",
		"DATABASE_URL",
		"PORT",
		"WORKER_CONCURRENCY",
		"WORKER_BASE_DELAY_MS",
		"WORKER_LEASE_MS",
	}
	for _, key := range keys {
		unsetEnv(t, key)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWD)
	})
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	oldValue, hadValue := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("failed to unset %s: %v", key, err)
	}

	t.Cleanup(func() {
		if hadValue {
			_ = os.Setenv(key, oldValue)
			return
		}
		_ = os.Unsetenv(key)
	})
}
