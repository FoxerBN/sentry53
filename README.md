# sentry53

sentry53 is a planned Go DNS filtering service. The project is currently a scaffold for a DNS server that will load blocked domains, forward allowed queries to upstream DNS resolvers, and expose simple runtime statistics.

## Planned Structure

```text
cmd/server/          Application entry point and dependency wiring
internal/blocklist/  Blocklist loading, storage, lookup, and reload logic
internal/resolver/   Upstream DNS forwarding
internal/server/     UDP and TCP DNS listener and query handler
internal/stats/      Counters and request logging
config/              Runtime configuration
blocklists/          Domain blocklist files
```

## Development

This repository uses Go modules.

```bash
go mod tidy
```

The DNS server implementation is not available yet. The current files only define the initial project layout.

## Configuration

Configuration will live in `config/config.yaml`. It is intended to contain upstream DNS servers, listener ports, and local file paths.

## Blocklists

Blocked domains will be stored in `blocklists/blocked_domains.txt`, with one domain per line.
