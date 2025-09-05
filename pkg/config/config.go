package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

type Config struct {
	Listen         string
	Routes         []Route
	ServerTimeouts ServerTimeouts
	ProxyTimeouts  ProxyTimeouts
}

type Route struct {
	Match        Match
	Upstream     string
	StripPrefix  string
	PreserveHost bool
	AddHeaders   map[string]string
}

type Match struct {
	Host   string
	Prefix string
}

type ServerTimeouts struct {
	Read  string
	Write string
	Idle  string
}

type ProxyTimeouts struct {
	Dial           string
	TLSHandshake   string
	IdleConn       string
	ResponseHeader string
}

func Load(path string) (Config, error) {
	const (
		// Server timeout defaults
		DEF_READ  = "5s"
		DEF_WRITE = "10s"
		DEF_IDLE  = "30s"

		// Proxy timeout defaults
		DEF_DIAL            = "5s"
		DEF_TLS_HANDSHAKE   = "5s"
		DEF_IDLE_CONN       = "30s"
		DEF_RESPONSE_HEADER = "10s"
	)

	var cfg Config

	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}

	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("decode config: %w", err)
	}

	cfg.Listen = strings.TrimSpace(cfg.Listen)
	if cfg.Listen == "" {
		return cfg, fmt.Errorf("Listen must not be empty")
	}
	if len(cfg.Routes) == 0 {
		return cfg, fmt.Errorf("routes must contain at least one route")
	}

	if cfg.ServerTimeouts.Read == "" {
		cfg.ServerTimeouts.Read = DEF_READ
	}
	if cfg.ServerTimeouts.Write == "" {
		cfg.ServerTimeouts.Write = DEF_WRITE
	}
	if cfg.ServerTimeouts.Idle == "" {
		cfg.ServerTimeouts.Idle = DEF_IDLE
	}
	if cfg.ProxyTimeouts.Dial == "" {
		cfg.ProxyTimeouts.Dial = DEF_DIAL
	}
	if cfg.ProxyTimeouts.TLSHandshake == "" {
		cfg.ProxyTimeouts.TLSHandshake = DEF_TLS_HANDSHAKE
	}
	if cfg.ProxyTimeouts.IdleConn == "" {
		cfg.ProxyTimeouts.IdleConn = DEF_IDLE_CONN
	}
	if cfg.ProxyTimeouts.ResponseHeader == "" {
		cfg.ProxyTimeouts.ResponseHeader = DEF_RESPONSE_HEADER
	}

	for _, pair := range [][2]string{
		{"server_timeouts.read", cfg.ServerTimeouts.Read},
		{"server_timeouts.write", cfg.ServerTimeouts.Write},
		{"server_timeouts.idle", cfg.ServerTimeouts.Idle},
		{"proxy_timeouts.dial", cfg.ProxyTimeouts.Dial},
		{"proxy_timeouts.tls_handshake", cfg.ProxyTimeouts.TLSHandshake},
		{"proxy_timeouts.idle_conn", cfg.ProxyTimeouts.IdleConn},
		{"proxy_timeouts.response_header", cfg.ProxyTimeouts.ResponseHeader},
	} {
		if _, err := time.ParseDuration(pair[1]); err != nil {
			return cfg, fmt.Errorf("%s: invalid duration %q: %v", pair[0], pair[1], err)
		}
	}

	for i := range cfg.Routes {
		r := &cfg.Routes[i]

		r.Match.Prefix = strings.TrimSpace(r.Match.Prefix)
		if r.Match.Prefix == "" || !strings.HasPrefix(r.Match.Prefix, "/") {
			return cfg, fmt.Errorf("routes[%d].match.prefix must start with \"/\"", i)
		}

		r.Match.Host = strings.TrimSpace(r.Match.Host)
		if r.Match.Host != "" {
			r.Match.Host = strings.ToLower(r.Match.Host)
			if strings.Contains(r.Match.Host, ":") {
				return cfg, fmt.Errorf("routes[%d].match.host must not include a port", i)
			}
		}

		r.StripPrefix = strings.TrimSpace(r.StripPrefix)
		if r.StripPrefix != "" && !strings.HasPrefix(r.StripPrefix, "/") {
			return cfg, fmt.Errorf("routes[%d].strip_prefix must start with \"/\"", i)
		}

		r.Upstream = strings.TrimSpace(r.Upstream)
		u, err := url.Parse(r.Upstream)
		if err != nil {
			return cfg, fmt.Errorf("routes[%d].upstream: parse error: %v", i, err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return cfg, fmt.Errorf("routes[%d].upstream: scheme must be http or https", i)
		}
		if u.Host == "" {
			return cfg, fmt.Errorf("routes[%d].upstream: host must not be empty", i)
		}

		for k, v := range r.AddHeaders {
			if strings.TrimSpace(k) == "" {
				return cfg, fmt.Errorf("routes[%d].add_headers has empty header name", i)
			}
			if strings.ContainsAny(v, "\r\n") {
				return cfg, fmt.Errorf("routes[%d].add_headers[%q] contains CR/LF", i, k)
			}
		}
	}

	return cfg, nil
}
