package proxy

import (
	"net"
	"net/http"
	"time"
)

// NewTransport returns a shared *http.Transport configured with sane timeouts and pooling.
func NewTransport(
	dial, tlsHandshake, idleConn, responseHeader time.Duration,
) *http.Transport {
	d := &net.Dialer{
		Timeout:   dial,
		KeepAlive: 30 * time.Second,
	}

	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,

		// Core timeouts
		DialContext:           d.DialContext,
		TLSHandshakeTimeout:   tlsHandshake,
		IdleConnTimeout:       idleConn,
		ResponseHeaderTimeout: responseHeader,
		ExpectContinueTimeout: 1 * time.Second,

		// Connection pooling
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,

		// Try HTTP/2 when available (ALPN).
		ForceAttemptHTTP2: true,
	}
}
