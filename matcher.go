package jsonrpcmatcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

const maxBodyRead = 131072

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

func (m *Matcher) Validate() error {
	if len(m.Methods) == 0 {
		return fmt.Errorf("jsonrpc_method: at least one method name is required")
	}
	for _, method := range m.Methods {
		if method == "" {
			return fmt.Errorf("jsonrpc_method: empty method name is not allowed")
		}
	}
	return nil
}

func (m *Matcher) MatchWithError(r *http.Request) (bool, error) {
	if r.Body == nil || r.Method != http.MethodPost {
		return false, nil
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyRead))
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	if len(body) == 0 {
		return false, nil
	}
	var msg struct {
		Method string `json:"method"`
	}
	if json.Unmarshal(body, &msg) != nil {
		return false, nil
	}
	_, ok := m.lookup[msg.Method]
	return ok, nil
}

func (m *Matcher) Match(r *http.Request) bool {
	match, _ := m.MatchWithError(r)
	return match
}

func (m *Matcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		m.Methods = append(m.Methods, d.RemainingArgs()...)
		if d.NextBlock(0) {
			return d.Err("jsonrpc_method does not support block syntax")
		}
	}
	return nil
}

func init() { caddy.RegisterModule(Matcher{}) }

var (
	_ caddy.Module                         = (*Matcher)(nil)
	_ caddy.Provisioner                    = (*Matcher)(nil)
	_ caddy.Validator                      = (*Matcher)(nil)
	_ caddyhttp.RequestMatcherWithError    = (*Matcher)(nil)
	_ caddyhttp.RequestMatcher             = (*Matcher)(nil)
	_ caddyfile.Unmarshaler                = (*Matcher)(nil)
)
