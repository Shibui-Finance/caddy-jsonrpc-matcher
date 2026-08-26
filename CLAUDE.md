# CLAUDE.md

## What this is

A Caddy matcher module written in Go. It inspects JSON-RPC POST request bodies and matches on the `method` field. One file of logic (`matcher.go`), one test file (`matcher_test.go`).

The module was built for MCP servers where Caddy's rate limiter needs to count tool calls (`tools/call`) without counting connection handshakes (`initialize`, `tools/list`). MCP Streamable HTTP sends all JSON-RPC messages as `POST /mcp`, so Caddy can't distinguish them by path or HTTP method alone.

## Commands

```bash
go test -v ./...    # run tests
go vet ./...        # static analysis
```

## Build with xcaddy

```bash
# standalone
xcaddy build --with github.com/Shibui-Finance/caddy-jsonrpc-matcher

# with rate limiting (the typical combination)
xcaddy build \
    --with github.com/mholt/caddy-ratelimit \
    --with github.com/Shibui-Finance/caddy-jsonrpc-matcher
```

For Docker builds, see `README.md`.

## Integration into a Caddyfile

The matcher registers as `http.matchers.jsonrpc_method`. Use it in named matchers:

```
@toolcalls {
    jsonrpc_method tools/call
}
rate_limit @toolcalls { ... }
```

Multiple methods: `jsonrpc_method tools/call resources/write`

Combines with standard Caddy matchers (`not remote_ip`, `not header`, etc.) inside the same named matcher block.

## Architecture

- `Matcher.Match()` reads up to 64KB of the POST body, parses the JSON-RPC `method` field, and re-attaches the body via `io.NopCloser(bytes.NewReader(...))` so the reverse proxy can still forward it.
- `Matcher.Provision()` builds a lookup map from the configured method list for O(1) matching.
- GET requests, nil bodies, empty bodies, and unparseable JSON all return false (no match).

## Deployed at

Used by the Shibui Finance MCP server (`mcp.shibui.finance`). The Caddy build is in `gamble2sql/deploy/Dockerfile.caddy`, the rate limit config in `gamble2sql/deploy/Caddyfile`.
