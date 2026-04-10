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
