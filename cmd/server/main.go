package main

import (
	"log"

	"sentry53/internal/blocklist"
	"sentry53/internal/resolver"
	"sentry53/internal/server"
)

func main() {
	blocked, err := blocklist.Load("blocklists/blocked_domains.txt")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("loaded %d blocked domains", blocked.Len())

	upstream := resolver.New("1.1.1.1:53")
	dnsServer := server.New(blocked, upstream)
	if err := dnsServer.ListenAndServe("192.168.1.174:53"); err != nil {
		log.Fatal(err)
	}
}
