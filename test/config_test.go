package test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"sentry53/internal/config"
)

func TestLoadConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	contents := []byte(`server:
  listen_address: "127.0.0.1:1053"
blocklist:
  path: "blocklist.txt"
resolver:
  upstreams: ["1.1.1.1:53"]
  timeout: "2s"
  tcp_fallback_on_truncation: true
cache:
  enabled: true
  max_entries: 10
  min_ttl: "5s"
  max_ttl: "1h"
  cleanup_interval: "1m"
`)
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}

	appConfig, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if appConfig.Resolver.Timeout != 2*time.Second {
		t.Fatalf("timeout = %s, want 2s", appConfig.Resolver.Timeout)
	}
	if !appConfig.Resolver.TCPFallback || appConfig.Cache.MaxEntries != 10 {
		t.Fatalf("unexpected config: %#v", appConfig)
	}
}

func TestLoadConfigRejectsInvalidCacheRange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	contents := []byte(`server:
  listen_address: "127.0.0.1:1053"
blocklist:
  path: "blocklist.txt"
resolver:
  upstreams: ["1.1.1.1:53"]
  timeout: "2s"
cache:
  enabled: true
  max_entries: 10
  min_ttl: "2h"
  max_ttl: "1h"
  cleanup_interval: "1m"
`)
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(path); err == nil {
		t.Fatal("Load() succeeded for an invalid cache TTL range")
	}
}
