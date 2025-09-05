package parsec

import (
	"net/http"

	"github.com/AJMerr/parsec/pkg/config"
	"github.com/AJMerr/parsec/pkg/proxy"
)

// HandlerFromFile loads a Parsec JSON config from disk and returns
// an http.Handler (the proxy Engine) ready to mount in your app.
func HandlerFromFile(path string) (http.Handler, error) {
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	eng, err := proxy.NewEngine(cfg)
	if err != nil {
		return nil, err
	}
	return eng, nil
}
