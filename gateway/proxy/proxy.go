package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/FC4RICA/hong-commerce/gateway/config"
	"go.uber.org/zap"
)

type ServiceProxy struct {
	proxy  *httputil.ReverseProxy
	target *url.URL
	logger *zap.Logger
}

func New(rawURL string, cfg *config.Config, logger *zap.Logger) (*ServiceProxy, error) {
	target, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid service URL %q: %w", rawURL, err)
	}

	rp := httputil.NewSingleHostReverseProxy(target)

	rp.Transport = &http.Transport{
		MaxIdleConns:        cfg.ProxyMaxIdleConns,
		MaxIdleConnsPerHost: cfg.ProxyMaxIdleConnsPerHost,
		IdleConnTimeout:     cfg.ProxyIdleConnTimeout,
		DisableCompression:  false,
	}

	rp.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		// Strip the gateway prefix, is handled per-route via chi's StripPrefix or manually
	}

	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error("upstream error",
			zap.String("url", r.URL.String()),
			zap.Error(err),
		)
		http.Error(w, "service unavaliable", http.StatusBadGateway)
	}

	return &ServiceProxy{
		proxy:  rp,
		target: target,
		logger: logger,
	}, nil
}

// Makes ServiceProxy implement http.Handler directly
func (s *ServiceProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

// For wildcard routes. Strips full prefix including /api/v1
func (s *ServiceProxy) StripAndForward(prefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
		s.proxy.ServeHTTP(w, r)
	}
}

// For exact routes. Rewrites path to a fixed target
func (s *ServiceProxy) ReverseWithPath(targetPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = targetPath
		s.proxy.ServeHTTP(w, r)
	}
}
