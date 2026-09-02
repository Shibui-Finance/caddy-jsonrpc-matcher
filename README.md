# caddy-jsonrpc-matcher

A Caddy module that matches HTTP requests by their JSON-RPC method field. Built for MCP servers where rate limiting should target tool calls (`tools/call`) without counting connection handshakes (`initialize`, `tools/list`).

## How it works

The matcher reads the `method` field from JSON-RPC POST request bodies and matches against a configured list. The body is buffered and re-attached so downstream handlers can still read it.

- Only matches POST requests
- Reads up to 128KB of the body (MCP requests are typically a few KB)
- GET requests, nil bodies, and invalid JSON never match
- Body is preserved after matching for the reverse proxy

## Build

Requires Go 1.25+ and xcaddy (constrained by Caddy v2.11).

```bash
# Run tests
go test -v ./...

# Build Caddy with this module
xcaddy build \
    --with github.com/Shibui-Finance/caddy-jsonrpc-matcher

# Combine with other modules (e.g. rate limiting)
xcaddy build \
    --with github.com/mholt/caddy-ratelimit \
    --with github.com/Shibui-Finance/caddy-jsonrpc-matcher
```

### Docker

```dockerfile
FROM caddy:2-builder-alpine AS builder
RUN xcaddy build \
    --with github.com/mholt/caddy-ratelimit \
    --with github.com/Shibui-Finance/caddy-jsonrpc-matcher

FROM caddy:2-alpine
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
```

## Caddyfile usage

The matcher directive is `jsonrpc_method` followed by one or more method names:

```
jsonrpc_method <method1> [method2] [method3] ...
```

### Rate-limit only tool calls on an MCP server

```
handle /mcp* {
    @toolcalls {
        jsonrpc_method tools/call
    }
    rate_limit @toolcalls {
        zone mcp_per_day {
            key {remote_host}
            events 500
            window 24h
        }
    }
    reverse_proxy backend:8000
}
```

Connection handshakes (`initialize`, `notifications/initialized`, `tools/list`) pass through unmetered.

### Combine with IP and header exemptions

```
handle /mcp* {
    @ratelimited {
        jsonrpc_method tools/call
        not remote_ip 10.0.0.0/8
        not header X-Bypass-Ratelimit *
    }
    rate_limit @ratelimited {
        zone mcp_per_hour {
            key {remote_host}
            events 250
            window 1h
        }
        zone mcp_per_day {
            key {remote_host}
            events 500
            window 24h
        }
    }
    reverse_proxy backend:8000
}
```

### Match multiple methods

```
@mutations {
    jsonrpc_method tools/call resources/write
}
```

## JSON format (API config)

```json
{
    "match": [{
        "jsonrpc_method": {
            "methods": ["tools/call"]
        }
    }]
}
```

## In production

This module powers the rate limiting layer at [Shibui Finance](https://shibui.finance), a Claude-native MCP server for financial data.
