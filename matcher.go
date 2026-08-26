package jsonrpcmatcher

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

const maxBodyRead = 65536

type Matcher struct {
	Methods []string `json:"methods"`
	lookup  map[string]struct{}
}

func (Matcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.matchers.jsonrpc_method",
		New: func() caddy.Module { return new(Matcher) },
	}
}

func (m *Matcher) Provision(_ caddy.Context) error {
	m.lookup = make(map[string]struct{}, len(m.Methods))
	for _, method := range m.Methods {
		m.lookup[method] = struct{}{}
	}
	return nil
}

func (m *Matcher) Match(r *http.Request) bool {
	if r.Body == nil || r.Method != http.MethodPost {
		return false
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyRead))
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(body))
	if err != nil || len(body) == 0 {
		return false
	}
	var msg struct {
		Method string `json:"method"`
	}
	if json.Unmarshal(body, &msg) != nil {
		return false
	}
	_, ok := m.lookup[msg.Method]
	return ok
}

func (m *Matcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	m.Methods = d.RemainingArgs()
	return nil
}

func init() { caddy.RegisterModule(Matcher{}) }

var (
	_ caddy.Module             = (*Matcher)(nil)
	_ caddy.Provisioner        = (*Matcher)(nil)
	_ caddyhttp.RequestMatcher = (*Matcher)(nil)
	_ caddyfile.Unmarshaler    = (*Matcher)(nil)
)
