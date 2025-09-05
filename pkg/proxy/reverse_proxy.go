package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func NewReverseProxy(
	target *url.URL,
	tr *http.Transport,
	preserveHost bool,
	stripPrefix string,
	addHeaders map[string]string,
	matchPrefix string,
) *httputil.ReverseProxy {
	rp := &httputil.ReverseProxy{Transport: tr}

	rp.Rewrite = func(pr *httputil.ProxyRequest) {
		pr.SetURL(target)
		pr.SetXForwarded()
		if preserveHost {
			pr.Out.Host = pr.In.Host
		}

		inPath := pr.In.URL.Path
		if stripPrefix != "" && strings.HasPrefix(inPath, matchPrefix) && strings.HasPrefix(inPath, stripPrefix) {
			stripped := strings.TrimPrefix(inPath, stripPrefix)
			if stripped == "" {
				stripped = "/"
			}

			base := strings.TrimRight(target.Path, "/")
			tail := strings.TrimLeft(stripped, "/")

			if tail == "" {
				if strings.HasSuffix(target.Path, "/") || base == "" {
					pr.Out.URL.Path = base + "/"
				} else {
					pr.Out.URL.Path = base
				}
			} else {
				if base == "" {
					pr.Out.URL.Path = "/" + tail
				} else {
					pr.Out.URL.Path = base + "/" + tail
				}
			}

			pr.Out.URL.RawPath = ""
		}

		pr.Out.URL.RawQuery = pr.In.URL.RawQuery

		for k, v := range addHeaders {
			pr.Out.Header.Set(k, v)
		}
	}

	return rp
}
