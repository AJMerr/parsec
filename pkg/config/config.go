package config

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
	PreserveHost string
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
