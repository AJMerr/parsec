package proxy

import (
	"net/http/httputil"
	"net/url"
)

type Engine struct {
	compiled []compiledRoute
}

type compiledRoute struct {
	host         string
	prefix       string
	stripPrefix  string
	preserveHost string
	addHeaders   string
	targetURL    *url.URL
	rp           *httputil.ReverseProxy
}
