package main

import (
	"log"

	"sentry53/internal/blocklist"
	"sentry53/internal/server"
)

func main() {
	blocked, err := blocklist.Load("blocklists/blocked_domains.txt")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("loaded %d blocked domains", blocked.Len())

	dnsServer := server.New(blocked)
	if err := dnsServer.ListenAndServe("127.0.0.1:1053"); err != nil {
		log.Fatal(err)
	}
}
