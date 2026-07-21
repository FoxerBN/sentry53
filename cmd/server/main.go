package main

import (
	"flag"
	"log"

	"sentry53/internal/blocklist"
	"sentry53/internal/cache"
	"sentry53/internal/config"
	"sentry53/internal/resolver"
	"sentry53/internal/server"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "path to the YAML configuration file")
	flag.Parse()

	appConfig, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	blocked, err := blocklist.Load(appConfig.Blocklist.Path)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("loaded %d blocked domains", blocked.Len())

	upstream := resolver.New(appConfig.Resolver.Upstreams, appConfig.Resolver.Timeout, appConfig.Resolver.TCPFallback)
	var responses *cache.Cache
	if appConfig.Cache.Enabled {
		responses = cache.New(cache.Config{
			MaxEntries:      appConfig.Cache.MaxEntries,
			MinTTL:          appConfig.Cache.MinTTL,
			MaxTTL:          appConfig.Cache.MaxTTL,
			CleanupInterval: appConfig.Cache.CleanupInterval,
		})
		defer responses.Close()
	}

	dnsServer := server.New(blocked, responses, upstream)
	log.Printf("listening on %s", appConfig.Server.ListenAddress)
	if err := dnsServer.ListenAndServe(appConfig.Server.ListenAddress); err != nil {
		log.Fatal(err)
	}
}
