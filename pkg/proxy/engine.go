package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/AJMerr/parsec/pkg/config"
)

type Engine struct {
	compiled []compiledRoute
}

type compiledRoute struct {
	host         string
	prefix       string
	stripPrefix  string
	preserveHost bool
	addHeaders   map[string]string
	targetURL    *url.URL
	rp           *httputil.ReverseProxy
}

func NewEngine(cfg config.Config) (*Engine, error) {
	//Parse durations
	dialDuration, err := time.ParseDuration(cfg.ProxyTimeouts.Dial)
	if err != nil {
		return nil, fmt.Errorf("proxy_timeouts.dial: %w", err)
	}
	tlsDuration, err := time.ParseDuration(cfg.ProxyTimeouts.TLSHandshake)
	if err != nil {
		return nil, fmt.Errorf("proxy_timeouts.tls_handshake: %w", err)
	}
	idleDuration, err := time.ParseDuration(cfg.ProxyTimeouts.IdleConn)
	if err != nil {
		return nil, fmt.Errorf("proxy_timeouts.idle_conn: %w", err)
	}
	respDuration, err := time.ParseDuration(cfg.ProxyTimeouts.ResponseHeader)
	if err != nil {
		return nil, fmt.Errorf("proxy_timeouts.response_header: %w", err)
	}

	tr := NewTransport(dialDuration, tlsDuration, idleDuration, respDuration)

	compiled := make([]compiledRoute, 0, len(cfg.Routes))

	for i, r := range cfg.Routes {
		u, err := url.Parse(r.Upstream)
		if err != nil {
			return nil, fmt.Errorf("routes[%d].upstream parse: %w", i, err)
		}

		cr := compiledRoute{
			host:         strings.ToLower(r.Match.Host),
			prefix:       r.Match.Prefix,
			stripPrefix:  r.StripPrefix,
			preserveHost: r.PreserveHost,
			addHeaders:   shallowCopy(r.AddHeaders),
			targetURL:    u,
		}

		cr.rp = NewReverseProxy(
			u,
			tr,
			cr.preserveHost,
			cr.stripPrefix,
			cr.addHeaders,
			cr.prefix,
		)
		compiled = append(compiled, cr)
	}
	return &Engine{compiled: compiled}, nil
}

func shallowCopy(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))

	for k, v := range in {
		out[k] = v
	}
	return out
}

func normalizeRequestHost(h string) string {
	if h == "" {
		return h
	}

	if strings.HasPrefix(h, "[") {
		if i := strings.LastIndex(h, "]"); i != -1 && i+1 < len(h) && h[i+1] == ':' {
			if host, _, err := net.SplitHostPort(h); err == nil {
				return strings.ToLower(strings.Trim(host, "[]"))
			}
		}
		return strings.ToLower(strings.Trim(h, "[]"))
	}
	if host, _, err := net.SplitHostPort(h); err == nil {
		return strings.ToLower(host)
	}
	return strings.ToLower(h)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e == nil || len(e.compiled) == 0 {
		http.NotFound(w, r)
		return
	}

	reqHost := normalizeRequestHost(r.Host)
	path := r.URL.Path

	for _, cr := range e.compiled {
		hostOK := cr.host == "" || cr.host == reqHost
		if hostOK && strings.HasPrefix(path, cr.prefix) {
			cr.rp.ServeHTTP(w, r)
			return
		}
	}

	http.NotFound(w, r)
}
