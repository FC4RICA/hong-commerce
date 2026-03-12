# Hong Commerce
A project demonstrating a simple ecommerce built with a microservice architecture.

## Architecture Overview
```
                  Frontend
                      |
                 API Gateway
                      |
       ┌──────────────┼────────────────┐
       ▼              ▼                ▼
Users Service   Catalog Service   Order Service
                                       |
                                       |
                              ┌────────┴────────┐
                              ▼                 ▼
                      Inventory Service   Payment Service
                              |
                              ▼
                        Message Queue
```

## Repository Structure
```
hong-commerce/
│
├── services/
│   ├── gateway/
│   ├── user-service/
│   ├── catalog-service/
│   ├── inventory-service/
│   ├── order-service/
│   └── payment-service/
│
├── pkg/                # shared Go packages
│
├── docker-compose.yml  # local development environment
│
└── README.md
```
