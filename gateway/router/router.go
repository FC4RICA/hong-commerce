package router

import (
	"net/http"

	"github.com/FC4RICA/hong-commerce/gateway/config"
	"github.com/FC4RICA/hong-commerce/gateway/middleware"
	"github.com/FC4RICA/hong-commerce/gateway/proxy"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func New(cfg *config.Config, logger *zap.Logger) (http.Handler, error) {
	r := chi.NewRouter()

	// Build proxies per service
	userProxy, err := proxy.New(cfg.UserServiceURL, cfg, logger)
	if err != nil {
		return nil, err
	}
	catalogProxy, err := proxy.New(cfg.CatalogServiceURL, cfg, logger)
	if err != nil {
		return nil, err
	}
	inventoryProxy, err := proxy.New(cfg.InventoryServiceURL, cfg, logger)
	if err != nil {
		return nil, err
	}
	orderProxy, err := proxy.New(cfg.OrderServiceURL, cfg, logger)
	if err != nil {
		return nil, err
	}
	paymentProxy, err := proxy.New(cfg.PaymentServiceURL, cfg, logger)
	if err != nil {
		return nil, err
	}

	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(logger))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})

		// Public routes
		r.Group(func(r chi.Router) {
			r.Mount("/users/login", userProxy.StripAndServe("/users"))
			r.Mount("/users/register", userProxy.StripAndServe("/users"))
			r.Mount("/catalog", catalogProxy.StripAndServe("/catalog"))
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.JWTSecret))

			r.Mount("/users", userProxy.StripAndServe("/users"))
			r.Mount("/inventories", inventoryProxy.StripAndServe("/inventories"))
			r.Mount("/orders", orderProxy.StripAndServe("/orders"))
			r.Mount("/payments", paymentProxy.StripAndServe("/payments"))
		})
	})

	return r, nil
}
