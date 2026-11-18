# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A Caddy matcher module written in Go. It inspects JSON-RPC POST request bodies and matches on the `method` field. One file of logic (`matcher.go`), one test file (`matcher_test.go`).

The module was built for MCP servers where Caddy's rate limiter needs to count tool calls (`tools/call`) without counting connection handshakes (`initialize`, `tools/list`). MCP Streamable HTTP sends all JSON-RPC messages as `POST /mcp`, so Caddy can't distinguish them by path or HTTP method alone.

Source: `github.com/Shibui-Finance/caddy-jsonrpc-matcher`. Go version is constrained by Caddy (currently v2.11).

## Working conventions

- **Public repo.** Everything committed here is public, PR titles, descriptions, commit messages and comments included: no details of the deployments that use the module (hosts, rate-limit numbers, internal paths or config).
- **Pull requests.** Squash is the only merge method and the PR title becomes the commit subject, so the title is a plain imperative. CI (`make check`) runs on every PR.
- **Without a Go toolchain** (a headless job, for instance), run no checks locally: CI's `make check` on the PR is the check.
- **Releases** only through the manual Release workflow, and only on the user's go.

## Commands

```bash
make check              # build + vet + test (same as CI)
make test               # go test -v ./...
make vet                # go vet ./...
make build              # go build ./...
make clean              # clear test cache

go test -v -run TestBodyIsPreservedAfterMatch ./...   # run a single test
```

## Build with xcaddy

```bash
xcaddy build \
    --with github.com/mholt/caddy-ratelimit \
    --with github.com/Shibui-Finance/caddy-jsonrpc-matcher
```

## Integration into a Caddyfile

The matcher registers as `http.matchers.jsonrpc_method`. Use it in named matchers:

```
@toolcalls {
    jsonrpc_method tools/call
}
rate_limit @toolcalls { ... }
```

Multiple methods: `jsonrpc_method tools/call resources/write`

## Architecture

- `Matcher.MatchWithError()` reads up to 128KB of the POST body, parses the JSON-RPC `method` field, and re-attaches the body via `io.NopCloser(bytes.NewReader(...))` so the reverse proxy can still forward it.
- `Matcher.Match()` is the legacy `RequestMatcher` adapter that delegates to `MatchWithError`.
- `Matcher.Provision()` builds a lookup map from the configured method list for O(1) matching.
- `Matcher.UnmarshalCaddyfile()` parses `jsonrpc_method arg1 arg2 ...`; block syntax is rejected.
- GET requests, nil bodies, empty bodies, and unparseable JSON all return false (no match).
- Interface compile-time checks at the bottom of `matcher.go` enforce `caddy.Module`, `caddy.Provisioner`, `caddy.Validator`, `RequestMatcherWithError`, `RequestMatcher`, and `caddyfile.Unmarshaler`.

## CI / Release

- CI runs `make check` on push/PR to master (`.github/workflows/ci.yml`).
- Release is manual via `workflow_dispatch` with a bump type (patch/minor/major). It computes the next semver tag and creates a GitHub Release (`.github/workflows/release.yml`).
