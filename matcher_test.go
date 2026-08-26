package jsonrpcmatcher

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/caddyserver/caddy/v2"
)

func newMatcher(methods ...string) *Matcher {
	m := &Matcher{Methods: methods}
	m.Provision(caddy.Context{})
	return m
}

func req(method, body string) *http.Request {
	r, _ := http.NewRequest(method, "/mcp", strings.NewReader(body))
	return r
}

func TestMatchesConfiguredMethod(t *testing.T) {
	m := newMatcher("tools/call")
	r := req("POST", `{"jsonrpc":"2.0","method":"tools/call","params":{}}`)
	if !m.Match(r) {
		t.Fatal("expected match on tools/call")
	}
}

func TestRejectsUnconfiguredMethod(t *testing.T) {
	m := newMatcher("tools/call")
	r := req("POST", `{"jsonrpc":"2.0","method":"initialize","params":{}}`)
	if m.Match(r) {
		t.Fatal("expected no match on initialize")
	}
}

func TestMultipleMethods(t *testing.T) {
	m := newMatcher("tools/call", "resources/read")
	for _, method := range []string{"tools/call", "resources/read"} {
		r := req("POST", `{"jsonrpc":"2.0","method":"`+method+`"}`)
		if !m.Match(r) {
			t.Fatalf("expected match on %s", method)
		}
	}
}

func TestGetRequestNeverMatches(t *testing.T) {
	m := newMatcher("tools/call")
	r, _ := http.NewRequest("GET", "/mcp", nil)
	if m.Match(r) {
		t.Fatal("GET should never match")
	}
}

func TestNilBodyNeverMatches(t *testing.T) {
	m := newMatcher("tools/call")
	r, _ := http.NewRequest("POST", "/mcp", nil)
	if m.Match(r) {
		t.Fatal("nil body should never match")
	}
}

func TestInvalidJSONNeverMatches(t *testing.T) {
	m := newMatcher("tools/call")
	r := req("POST", `not json at all`)
	if m.Match(r) {
		t.Fatal("invalid JSON should never match")
	}
}

func TestBodyIsPreservedAfterMatch(t *testing.T) {
	m := newMatcher("tools/call")
	body := `{"jsonrpc":"2.0","method":"tools/call","params":{"name":"query"}}`
	r := req("POST", body)
	m.Match(r)
	got, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("reading body after match: %v", err)
	}
	if string(got) != body {
		t.Fatalf("body not preserved: got %q", got)
	}
}

func TestEmptyBodyNeverMatches(t *testing.T) {
	m := newMatcher("tools/call")
	r := req("POST", "")
	if m.Match(r) {
		t.Fatal("empty body should never match")
	}
}
