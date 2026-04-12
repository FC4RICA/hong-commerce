# API Gateway

The entry point for all client requests in Hong Commerce. It handles routing, authentication, and proxying to the appropriate downstream service — clients never talk to services directly.


## How It Works

The gateway sits in front of all services. When a request comes in, it:

1. Validates the JWT (on protected routes) and injects `X-User-ID` / `X-User-Role` headers
2. Strips the route prefix and forwards the request to the target service 
3. Streams the response back to the client


## Configuration

All configuration is via environment variables.

| Variable | Description | Default |
|---|---|---|
| `JWT_SECRET` | Secret used to validate JWT tokens | — |
| `USER_SERVICE_URL` | user-service base URL | `http://user-service:8081` |
| `CATALOG_SERVICE_URL` | catalog-service base URL | `http://catalog-service:8082` |
| `ORDER_SERVICE_URL` | order-service base URL | `http://order-service:8083` |
| `INVENTORY_SERVICE_URL` | inventory-service base URL | `http://inventory-service:8084` |
| `PAYMENT_SERVICE_URL` | payment-service base URL | `http://payment-service:8085` |


## Adding New Routes

The gateway uses two proxy methods depending on the route type. Here's when to use each and how to add them.

### Proxy Methods

**`StripAndForward(prefix string)`** — for wildcard routes that forward any sub-path to a service. Strips the full `/api/v1/<service>` prefix before forwarding.

```
Client: GET /api/v1/orders/123/items
Strip:  /api/v1/orders
Result: GET /123/items → order-service
```

**`ReverseWithPath(targetPath string)`** — for exact public endpoints where you want full control over what the downstream service receives, regardless of what the client sent.

```
Client: POST /api/v1/users/login
Target: /login
Result: POST /login → user-service
```

### Adding a Public Exact Route

Use this when exposing a specific endpoint without auth (e.g. login, register, password reset).

In `router/router.go`, add inside the public `r.Group`:

```go
r.Group(func(r chi.Router) {
    // Example: add a forgot-password route
    r.Post("/users/forgot-password", userProxy.ReverseWithPath("/forgot-password"))
})
```

The first argument is the path the client calls. The `ReverseWithPath` argument is the path the downstream service receives.

### Adding a Public Wildcard Route

Use this when an entire service or resource is publicly accessible (e.g. catalog browsing).

```go
r.Group(func(r chi.Router) {
    // Example: add a public promotions service
    r.Handle("/promotions", promotionsProxy.StripAndForward("/api/v1/promotions"))
})
```

The `StripAndForward` argument must match the full prefix including `/api/v1`.

### Adding a Protected Wildcard Route

Use this for any authenticated resource. Add inside the protected `r.Group`:

```go
r.Group(func(r chi.Router) {
    r.Use(middleware.Auth(cfg.JWTSecret))

    // Example: add a notifications service
    r.Handle("/notifications", notificationsProxy.StripAndForward("/api/v1/notifications"))
})
```

### Current Route Table

| Method | Path | Downstream | Auth |
|---|---|---|---|
| `GET` | `/api/v1/health` | gateway | No |
| `POST` | `/api/v1/users/login` | `POST /login` → user-service | No |
| `POST` | `/api/v1/users/register` | `POST /register` → user-service | No |
| `*` | `/api/v1/catalog/*` | `/*` → catalog-service | No |
| `*` | `/api/v1/users/*` | `/*` → user-service | Yes |
| `*` | `/api/v1/inventories/*` | `/*` → inventory-service | Yes |
| `*` | `/api/v1/orders/*` | `/*` → order-service | Yes |
| `*` | `/api/v1/payments/*` | `/*` → payment-service | Yes |