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
			r.Post("/users/login", userProxy.ReverseWithPath("/login"))
			r.Post("/users/register", userProxy.ReverseWithPath("/register"))
			r.Mount("/catalog", catalogProxy.StripAndForward("/api/v1/catalog"))
			r.Mount("/orders/swagger", orderProxy.StripAndForward("/api/v1/orders"))
			r.Mount("/inventories/swagger", inventoryProxy.StripAndForward("/api/v1/inventories"))

			// Unified API Portal Dashboard
			r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write([]byte(swaggerHTML))
			})
			r.Get("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(masterSwaggerJSON))
			})
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.JWTSecret))

			r.Mount("/users", userProxy.StripAndForward("/api/v1/users"))
			r.Mount("/inventories", inventoryProxy.StripAndForward("/api/v1/inventories"))
			r.Mount("/orders", orderProxy.StripAndForward("/api/v1/orders"))
			r.Mount("/payments", paymentProxy.StripAndForward("/api/v1/payments"))

			// Admin routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))

				r.Post("/users/admin/register", userProxy.ReverseWithPath("/admin/register"))
			})
		})
	})

	return r, nil
}

const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Hong Commerce - Unified API Portal</title>
  <link rel="stylesheet" type="text/css" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.11.0/swagger-ui.min.css" />
  <style>
    html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
    .topbar { background-color: #1b1b1b !important; padding: 10px 0 !important; }
    .topbar .download-url-wrapper { display: none !important; } /* Hide URL wrapper */
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.11.0/swagger-ui-bundle.min.js"></script>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.11.0/swagger-ui-standalone-preset.min.js"></script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: "/api/v1/swagger.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      });
      window.ui = ui;
    };
  </script>
</body>
</html>`

const masterSwaggerJSON = `{
  "openapi": "3.0.0",
  "info": {
    "title": "Hong Commerce - Unified API Portal",
    "description": "Unified interactive documentation and playground for all microservices in the Hong Commerce platform. Access all endpoints (User, Inventory, Order, Payment) through the API Gateway with built-in JWT authorization.",
    "version": "1.0.0"
  },
  "servers": [
    {
      "url": "http://localhost:8080",
      "description": "API Gateway (Local)"
    }
  ],
  "components": {
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT",
        "description": "Enter your JWT token in the format: <token_value>."
      }
    }
  },
  "security": [
    {
      "BearerAuth": []
    }
  ],
  "paths": {
    "/api/v1/users/register": {
      "post": {
        "summary": "Register a new user",
        "tags": ["User Service"],
        "security": [],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "email": { "type": "string", "example": "user@example.com" },
                  "password": { "type": "string", "example": "password123" },
                  "name": { "type": "string", "example": "John Doe" }
                },
                "required": ["email", "password", "name"]
              }
            }
          }
        },
        "responses": {
          "201": { "description": "User successfully registered" },
          "400": { "description": "Invalid input" }
        }
      }
    },
    "/api/v1/users/login": {
      "post": {
        "summary": "Log in and obtain a JWT token",
        "tags": ["User Service"],
        "security": [],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "email": { "type": "string", "example": "user@example.com" },
                  "password": { "type": "string", "example": "password123" }
                },
                "required": ["email", "password"]
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Login successful",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "token": { "type": "string", "example": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." }
                  }
                }
              }
            }
          },
          "401": { "description": "Invalid credentials" }
        }
      }
    },
    "/api/v1/users/admin/register": {
      "post": {
        "summary": "Register a new Admin user",
        "tags": ["User Service"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "email": { "type": "string", "example": "admin@example.com" },
                  "password": { "type": "string", "example": "password123" },
                  "name": { "type": "string", "example": "Admin User" }
                },
                "required": ["email", "password", "name"]
              }
            }
          }
        },
        "responses": {
          "201": { "description": "Admin user successfully registered" },
          "403": { "description": "Forbidden - requires Admin privileges" }
        }
      }
    },
    "/api/v1/users/me": {
      "get": {
        "summary": "Get authenticated user profile",
        "tags": ["User Service"],
        "responses": {
          "200": { "description": "Successfully retrieved profile" },
          "401": { "description": "Unauthorized" }
        }
      }
    },
    "/api/v1/inventories/items": {
      "get": {
        "summary": "List all inventory items",
        "tags": ["Inventory Service"],
        "responses": {
          "200": { "description": "List of items retrieved" }
        }
      },
      "post": {
        "summary": "Create a new inventory item",
        "tags": ["Inventory Service"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "name": { "type": "string", "example": "Sony PlayStation 5" },
                  "quantity": { "type": "integer", "example": 10 }
                },
                "required": ["name", "quantity"]
              }
            }
          }
        },
        "responses": {
          "201": { "description": "Item created successfully" }
        }
      }
    },
    "/api/v1/inventories/items/{id}": {
      "get": {
        "summary": "Get inventory item by ID",
        "tags": ["Inventory Service"],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": { "description": "Item details retrieved" },
          "404": { "description": "Item not found" }
        }
      },
      "put": {
        "summary": "Update an inventory item",
        "tags": ["Inventory Service"],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "name": { "type": "string" },
                  "quantity": { "type": "integer" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Item updated successfully" }
        }
      },
      "delete": {
        "summary": "Delete an inventory item",
        "tags": ["Inventory Service"],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": { "description": "Item deleted successfully" }
        }
      }
    },
    "/api/v1/orders": {
      "get": {
        "summary": "Get all orders",
        "tags": ["Order Service"],
        "responses": {
          "200": { "description": "List of all orders retrieved" }
        }
      },
      "post": {
        "summary": "Place a new order (Initiates Saga Flow)",
        "tags": ["Order Service"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "items": {
                    "type": "array",
                    "items": {
                      "type": "object",
                      "properties": {
                        "product_id": { "type": "string", "example": "cm0y3b9u10000..." },
                        "product_name": { "type": "string", "example": "Sony PlayStation 5" },
                        "unit_price": { "type": "number", "example": 16900 },
                        "quantity": { "type": "integer", "example": 2 }
                      },
                      "required": ["product_id", "product_name", "unit_price", "quantity"]
                    }
                  }
                },
                "required": ["items"]
              }
            }
          }
        },
        "responses": {
          "201": { "description": "Order successfully created" }
        }
      }
    },
    "/api/v1/orders/{id}": {
      "get": {
        "summary": "Get order details by ID",
        "tags": ["Order Service"],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": { "description": "Order details retrieved" },
          "404": { "description": "Order not found" }
        }
      }
    },
    "/api/v1/orders/user/{userId}": {
      "get": {
        "summary": "Get orders by User ID",
        "tags": ["Order Service"],
        "parameters": [
          { "name": "userId", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": { "description": "List of user orders retrieved" }
        }
      }
    },
    "/api/v1/payments/{paymentID}/status": {
      "patch": {
        "summary": "Update payment status (Triggers Payment Completion)",
        "tags": ["Payment Service"],
        "parameters": [
          { "name": "paymentID", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "status": { "type": "string", "example": "COMPLETED", "enum": ["COMPLETED", "FAILED"] },
                  "transactionRef": { "type": "string", "example": "TXN-102938" },
                  "amountPaid": { "type": "number", "example": 33800 }
                },
                "required": ["status", "transactionRef", "amountPaid"]
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Payment status updated successfully" }
        }
      }
    }
  }
}`
