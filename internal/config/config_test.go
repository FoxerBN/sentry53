package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
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

	config, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if config.Resolver.Timeout != 2*time.Second {
		t.Fatalf("timeout = %s, want 2s", config.Resolver.Timeout)
	}
	if !config.Resolver.TCPFallback || config.Cache.MaxEntries != 10 {
		t.Fatalf("unexpected config: %#v", config)
	}
}

func TestLoadRejectsInvalidCacheRange(t *testing.T) {
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
	if _, err := Load(path); err == nil {
		t.Fatal("Load() succeeded for an invalid cache TTL range")
	}
}
