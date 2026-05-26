# RabbitMQ Integration: Order Service <-> Inventory Service

This document outlines how the **Order Service** and **Inventory Service** integrate via RabbitMQ asynchronously.

## Overview

The integration involves three main flows:
1. **Order Creation**: Order Service publishes a message when an order is created. Inventory Service consumes this to reserve stock.
2. **Stock Reservation Notification**: Inventory Service publishes a message when stock is successfully reserved.
3. **Stock Reservation Failure**: Inventory Service publishes a message when stock cannot be reserved (e.g., insufficient quantity).

---

## 1. Flow: Order Creation (Order -> Inventory)

When an order is created, the Order Service must publish a message so the Inventory Service can deduct/reserve the required stock.

### RabbitMQ Configuration
- **Exchange**: `order.events` (Topic)
- **Routing Key**: `order.created`
- **Queue (Inventory Service listens on)**: `inventory.order_events`

### JSON Structure (Payload)

```json
{
  "id": "string (OrderID)",
  "items": [
    {
      "id": "string (ProductID)",
      "quantity": 2
    }
  ],
  "customerID": "string",
  "timestamp": "string"
}
```

*Note: The Inventory Service uses the `id` field as an **Idempotency Key** to ensure that duplicate messages do not result in multiple stock deductions.*

---

## 2. Flow: Stock Reservation Success (Inventory -> Order)

When stock is successfully reserved for an order, the Inventory Service will publish an event for the Order Service (or subsequent saga steps) to process.

### RabbitMQ Configuration
- **Exchange**: `inventory.events` (Topic)
- **Routing Key**: `inventory.reserved`

### JSON Structure (Payload)

```json
{
  "orderId": "string",
  "items": [
    {
      "id": "string",
      "quantity": 2
    }
  ]
}
```

---

## 3. Flow: Stock Reservation Failure (Inventory -> Order)

When stock cannot be reserved (e.g., insufficient stock or missing product), the Inventory Service will publish a failure event to trigger compensating transactions in the saga.

### RabbitMQ Configuration
- **Exchange**: `inventory.events` (Topic)
- **Routing Key**: `inventory.failed`

### JSON Structure (Payload)

```json
{
  "orderId": "string",
  "reason": "string (Error message, e.g., 'Insufficient stock')",
  "timestamp": "string (ISO 8601)"
}
```

---
