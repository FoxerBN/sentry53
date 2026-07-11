# sentry53

sentry53 is a small DNS filtering service written in Go. It loads a local domain
blocklist into memory and serves DNS over UDP and TCP. Blocked queries receive
an `NXDOMAIN` response; allowed-query forwarding to an upstream resolver is the
next implementation milestone.

## Structure

```text
cmd/server/          Application entry point and dependency wiring
internal/blocklist/  Blocklist loading, normalization, and lookup
internal/resolver/   Upstream DNS forwarding (planned)
internal/server/     UDP/TCP listeners and DNS request handler
internal/stats/      Runtime counters and logging (planned)
config/              Runtime configuration
blocklists/          Domain blocklist files
```

## Current Behavior

- The blocklist is loaded once during startup.
- DNS names are normalized before exact matching.
- Blocked domains return `NXDOMAIN`.
- Allowed domains currently return `SERVFAIL` until forwarding is implemented.

Run the development server:

```bash
go run ./cmd/server
```

The current listener is intended for local development on port `1053`. Runtime
configuration and upstream forwarding will be added in later milestones.
