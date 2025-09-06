# Parsec
A tiny, Caddy-like reverse proxy written in Go. A static JSON config, and clean defaults. Perfect for local/dev setups (front a React dev server, fan out /api to a Go backend) and small services that don’t need TLS/ACME, hot reload, or load balancing.

- Built on net/http, httputil.ReverseProxy, and a tuned http.Transport
- Route by host (optional) and path prefix (required)
- Options per route: strip_prefix, preserve_host, add_headers
- Sensible timeouts + graceful shutdown in the CLI
- Library mode for embedding inside your app

### The why
Parsec was not intended to be a competitor in the market of open suorce reverse proxies. This project was built for the purpose of learning. It is one part of a four part project I decided to build which can be seen [here](https://github.com/AJMerr/NYMToDo)

## Install
```
go install github.com/AJMerr/parsec/cmd/parsecd@latest
```

### Library (Embedded in your app)
```
import (
  "log"
  "net/http"

  "github.com/AJMerr/parsec/pkg/parsec"
)

func main() {
  h, err := parsec.HandlerFromFile("./parsec.json")
  if err != nil { log.Fatal(err) }
  log.Fatal(http.ListenAndServe(":5173", h))
}
```
**NOTE:** in library mode, the listen field in JSON is not used. You bind your own http.Server.

## Quickstart
### parsec.json
```
{
  "listen": ":5173",
  "routes": [
    {
      "match": { "prefix": "/api" },
      "upstream": "http://localhost:8080",
      "strip_prefix": "/api",
      "preserve_host": false,
      "add_headers": { "X-Parsec": "1" }
    },
    {
      "match": { "prefix": "/" },
      "upstream": "http://localhost:5174",
      "preserve_host": false
    }
  ],
  "server_timeouts": { "read": "5s", "write": "10s", "idle": "60s" },
  "proxy_timeouts":  { "dial": "5s", "tls_handshake": "5s", "idle_conn": "60s", "response_header": "10s" }
}
```

## Config Reference
### Top-level:
- `listen` (string): address for CLI mode, e.g. ":8080". Ignored in library mode.
- `routes` (array): ordered list. First matching route wins.
- `server_timeouts` (object): {read, write, idle} as Go durations.
- `proxy_timeouts` (object): {dial, tls_handshake, idle_conn, response_header} as Go durations.

### Route:
- `match.host` (string, optional): exact host match (no port). Empty = any host.
- `match.prefix` (string, required): path prefix (must start with /).
- `upstream` (string, required): target URL, e.g. http://127.0.0.1:9000.
- `strip_prefix` (string, optional): remove this prefix before proxying.
- `preserve_host` (bool, default false): keep incoming Host. Default sets Host to upstream host.
- `add_headers` (map, optional): extra headers injected to upstream requests.

### Matching Semantics:
```
( match.host is empty OR match.host == req.Host_without_port )
AND
( request path hasPrefix match.prefix )
```

## Header behavior
- Adds standard forwarding headers:
    - `X-Forwarded-For` (appends client IP)
    - `X-Forwarded-Proto` (http/https)

- `preserve_host: false` (default) sets Request.Host to upstream host.
- Any `add_headers` are set/overwritten on the upstream request.

## Timeouts (defaults)
- If not specified, Parsec applies sane defaults:
- Server: `read=5s`, `write=10s`, `idle=30s`
- Proxy: `dial=5s`, `tls_handshake=5s`, `idle_conn=30s`, `response_header=10s`

## Project Layout
```
.
├─ cmd/
│  └─ parsecd/           # CLI wrapper (reads JSON, starts server with graceful shutdown)
├─ pkg/
│  ├─ config/            # JSON loader + validation (types & Load)
│  ├─ proxy/             # Engine (http.Handler), route compilation, ServeHTTP
│  └─ parsec/            # Small facade for library use (HandlerFromFile)
└─ README.md
```
