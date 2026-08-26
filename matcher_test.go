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
	if match, err := m.MatchWithError(r); err != nil || !match {
		t.Fatal("expected match on tools/call")
	}
}

func TestRejectsUnconfiguredMethod(t *testing.T) {
	m := newMatcher("tools/call")
	r := req("POST", `{"jsonrpc":"2.0","method":"initialize","params":{}}`)
	if match, err := m.MatchWithError(r); err != nil || match {
		t.Fatal("expected no match on initialize")
	}
}

func TestMultipleMethods(t *testing.T) {
	m := newMatcher("tools/call", "resources/read")
	for _, method := range []string{"tools/call", "resources/read"} {
		r := req("POST", `{"jsonrpc":"2.0","method":"`+method+`"}`)
		if match, err := m.MatchWithError(r); err != nil || !match {
			t.Fatalf("expected match on %s", method)
		}
	}
}

func TestGetRequestNeverMatches(t *testing.T) {
	m := newMatcher("tools/call")
	r, _ := http.NewRequest("GET", "/mcp", nil)
	if match, err := m.MatchWithError(r); err != nil || match {
		t.Fatal("GET should never match")
	}
}

func TestNilBodyNeverMatches(t *testing.T) {
	m := newMatcher("tools/call")
	r, _ := http.NewRequest("POST", "/mcp", nil)
	if match, err := m.MatchWithError(r); err != nil || match {
		t.Fatal("nil body should never match")
	}
}

func TestInvalidJSONNeverMatches(t *testing.T) {
	m := newMatcher("tools/call")
	r := req("POST", `not json at all`)
	if match, err := m.MatchWithError(r); err != nil || match {
		t.Fatal("invalid JSON should never match")
	}
}

func TestBodyIsPreservedAfterMatch(t *testing.T) {
	m := newMatcher("tools/call")
	body := `{"jsonrpc":"2.0","method":"tools/call","params":{"name":"query"}}`
	r := req("POST", body)
	m.MatchWithError(r)
	got, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("reading body after match: %v", err)
	}
	if string(got) != body {
		t.Fatalf("body not preserved: got %q", got)
	}
}

func TestBodyIsPreservedAfterNoMatch(t *testing.T) {
	m := newMatcher("tools/call")
	body := `{"jsonrpc":"2.0","method":"initialize","params":{}}`
	r := req("POST", body)
	m.MatchWithError(r)
	got, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("reading body after no-match: %v", err)
	}
	if string(got) != body {
		t.Fatalf("body not preserved on no-match: got %q", got)
	}
}

func TestEmptyBodyNeverMatches(t *testing.T) {
	m := newMatcher("tools/call")
	r := req("POST", "")
	if match, err := m.MatchWithError(r); err != nil || match {
		t.Fatal("empty body should never match")
	}
}

func TestValidateRejectsZeroMethods(t *testing.T) {
	m := &Matcher{}
	m.Provision(caddy.Context{})
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for zero methods")
	}
}

func TestValidateRejectsEmptyStringMethod(t *testing.T) {
	m := &Matcher{Methods: []string{"tools/call", ""}}
	m.Provision(caddy.Context{})
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for empty string method")
	}
}

func TestLegacyMatchInterface(t *testing.T) {
	m := newMatcher("tools/call")
	r := req("POST", `{"jsonrpc":"2.0","method":"tools/call"}`)
	if !m.Match(r) {
		t.Fatal("legacy Match should delegate to MatchWithError")
	}
}
