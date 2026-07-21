package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig
	Blocklist BlocklistConfig
	Resolver  ResolverConfig
	Cache     CacheConfig
}

type ServerConfig struct{ ListenAddress string }
type BlocklistConfig struct{ Path string }
type ResolverConfig struct {
	Upstreams   []string
	Timeout     time.Duration
	TCPFallback bool
}
type CacheConfig struct {
	Enabled         bool
	MaxEntries      int
	MinTTL          time.Duration
	MaxTTL          time.Duration
	CleanupInterval time.Duration
}

type fileConfig struct {
	Server struct {
		ListenAddress string `yaml:"listen_address"`
	} `yaml:"server"`
	Blocklist struct {
		Path string `yaml:"path"`
	} `yaml:"blocklist"`
	Resolver struct {
		Upstreams   []string `yaml:"upstreams"`
		Timeout     string   `yaml:"timeout"`
		TCPFallback bool     `yaml:"tcp_fallback_on_truncation"`
	} `yaml:"resolver"`
	Cache struct {
		Enabled         bool   `yaml:"enabled"`
		MaxEntries      int    `yaml:"max_entries"`
		MinTTL          string `yaml:"min_ttl"`
		MaxTTL          string `yaml:"max_ttl"`
		CleanupInterval string `yaml:"cleanup_interval"`
	} `yaml:"cache"`
}

func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer file.Close()

	var raw fileConfig
	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	if err := decoder.Decode(&raw); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}

	timeout, err := parsePositiveDuration("resolver.timeout", raw.Resolver.Timeout)
	if err != nil {
		return Config{}, err
	}
	config := Config{
		Server:    ServerConfig{ListenAddress: raw.Server.ListenAddress},
		Blocklist: BlocklistConfig{Path: raw.Blocklist.Path},
		Resolver:  ResolverConfig{Upstreams: raw.Resolver.Upstreams, Timeout: timeout, TCPFallback: raw.Resolver.TCPFallback},
		Cache:     CacheConfig{Enabled: raw.Cache.Enabled},
	}
	if config.Server.ListenAddress == "" {
		return Config{}, fmt.Errorf("server.listen_address is required")
	}
	if err := validateAddress("server.listen_address", config.Server.ListenAddress); err != nil {
		return Config{}, err
	}
	if config.Blocklist.Path == "" {
		return Config{}, fmt.Errorf("blocklist.path is required")
	}
	if len(config.Resolver.Upstreams) == 0 {
		return Config{}, fmt.Errorf("resolver.upstreams must contain at least one address")
	}
	for _, upstream := range config.Resolver.Upstreams {
		if err := validateAddress("resolver.upstreams", upstream); err != nil {
			return Config{}, err
		}
	}
	if !config.Cache.Enabled {
		return config, nil
	}
	if raw.Cache.MaxEntries <= 0 {
		return Config{}, fmt.Errorf("cache.max_entries must be greater than zero")
	}
	minTTL, err := parseNonNegativeDuration("cache.min_ttl", raw.Cache.MinTTL)
	if err != nil {
		return Config{}, err
	}
	maxTTL, err := parsePositiveDuration("cache.max_ttl", raw.Cache.MaxTTL)
	if err != nil {
		return Config{}, err
	}
	cleanupInterval, err := parsePositiveDuration("cache.cleanup_interval", raw.Cache.CleanupInterval)
	if err != nil {
		return Config{}, err
	}
	if minTTL > maxTTL {
		return Config{}, fmt.Errorf("cache.min_ttl must not exceed cache.max_ttl")
	}
	config.Cache.MaxEntries = raw.Cache.MaxEntries
	config.Cache.MinTTL = minTTL
	config.Cache.MaxTTL = maxTTL
	config.Cache.CleanupInterval = cleanupInterval
	return config, nil
}

func parsePositiveDuration(name, value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", name)
	}
	return duration, nil
}

func parseNonNegativeDuration(name, value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil || duration < 0 {
		return 0, fmt.Errorf("%s must be a non-negative duration", name)
	}
	return duration, nil
}

func validateAddress(name, address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil || strings.TrimSpace(host) == "" {
		return fmt.Errorf("%s must be a host:port address", name)
	}
	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil || portNumber == 0 {
		return fmt.Errorf("%s must use a port from 1 to 65535", name)
	}
	return nil
}
