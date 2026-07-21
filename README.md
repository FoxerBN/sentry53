# sentry53

sentry53 is a small DNS filtering forwarder written in Go. It blocks domains
from a local list, forwards allowed DNS queries to upstream resolvers, and can
cache successful answers in memory.

It listens on both UDP and TCP. The default configuration is safe for local
development: `127.0.0.1:1053`.

## What it does

- Blocks exact domains and their subdomains with an `NXDOMAIN` response.
- Forwards allowed queries to the configured upstream DNS resolvers.
- Tries the next upstream only after a network-level failure, such as a
  timeout or refused connection.
- Retries a truncated upstream UDP response over TCP when enabled.
- Caches positive responses by DNS name, record type, and class.
- Honors DNS TTLs, returns decreasing TTLs on cache hits, and bounds cache
  memory with `max_entries`.

## Requirements

- Go 1.25 or newer
- Network access to at least one configured upstream DNS resolver

## Quick start

1. Review [config/config.yaml](config/config.yaml). The included defaults
   listen only on your machine at `127.0.0.1:1053`.
2. Start the server from the repository root:

   ```bash
   go run ./cmd/server
   ```

3. In another terminal, query it:

   ```bash
   dig @127.0.0.1 -p 1053 example.com A
   ```

4. Test a blocked domain from [blocklists/blocked_domains.txt](blocklists/blocked_domains.txt):

   ```bash
   dig @127.0.0.1 -p 1053 ads.example.com A
   ```

   The response should use the `NXDOMAIN` status.

## Build and run

To create a binary:

```bash
go build -o bin/sentry53 ./cmd/server
./bin/sentry53
```

Use another configuration file with `-config`:

```bash
./bin/sentry53 -config /etc/sentry53/config.yaml
```

Paths such as `blocklist.path` are currently resolved relative to the process
working directory. If you run the binary outside this repository, use an
absolute blocklist path in the configuration.

## Configuration

All runtime settings are read from YAML at startup. Unknown YAML fields and
invalid values cause startup to fail, rather than silently using a bad
configuration.

```yaml
server:
  listen_address: "127.0.0.1:1053"

blocklist:
  path: "blocklists/blocked_domains.txt"

resolver:
  upstreams:
    - "1.1.1.1:53"
    - "1.0.0.1:53"
  timeout: "3s"
  tcp_fallback_on_truncation: true

cache:
  enabled: true
  max_entries: 10000
  min_ttl: "30s"
  max_ttl: "24h"
  cleanup_interval: "5m"
```

| Setting | Meaning |
| --- | --- |
| `server.listen_address` | Address and port for both UDP and TCP listeners. |
| `blocklist.path` | File with one domain per line; comments start with `#`. |
| `resolver.upstreams` | Ordered upstream `host:port` addresses. |
| `resolver.timeout` | Per-attempt UDP or TCP timeout; Go duration syntax. |
| `tcp_fallback_on_truncation` | Retry the same upstream over TCP after a truncated UDP response. |
| `cache.max_entries` | Hard maximum number of cached questions. New entries are skipped once full. |
| `cache.min_ttl` | Do not cache answers whose smallest answer TTL is below this value. |
| `cache.max_ttl` | Maximum time that an otherwise valid response stays in the cache. |
| `cache.cleanup_interval` | Interval for removing expired entries that are not queried again. |

Durations use Go syntax: `500ms`, `3s`, `5m`, `24h`. Set
`cache.enabled: false` to bypass caching; cache-specific values are then not
used.

## How resolver fallback works

The resolver uses the first upstream unless a transport error occurs. A normal
DNS result — including `NXDOMAIN` or `SERVFAIL` — is returned directly and does
not trigger a query to another resolver. This avoids inconsistent answers.

If a UDP response is marked truncated, sentry53 retries the original request
over TCP to that same upstream. Only a TCP connection failure then advances to
the next configured upstream.

## Cache behavior

The cache stores only responses with answer records. Its key contains the
fully-qualified lower-case DNS name, query type, and query class. Cached DNS
messages are copied on input and output, so requests cannot mutate shared
memory.

Expiration is handled twice: every lookup rejects and removes expired entries
(lazy cleanup), while a background ticker removes entries that are never
looked up again (periodic cleanup). The server stops the ticker when it exits.

## Development checks

```bash
go test ./...
go vet ./...
```

## Notes for deployment

Binding to port `53` usually requires elevated privileges. Change
`listen_address` intentionally and do not expose this development DNS forwarder
to an untrusted network without adding appropriate access controls and
operational monitoring.
