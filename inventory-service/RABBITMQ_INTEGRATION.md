# RabbitMQ Integration: Order Service <-> Inventory Service

This document outlines how the **Order Service** and **Inventory Service** integrate via RabbitMQ asynchronously.

## Overview

The integration involves two main flows:
1. **Order Creation**: Order Service publishes a message when an order is created. Inventory Service consumes this to reserve stock.
2. **Stock Reservation Notification**: Inventory Service publishes a message when stock is successfully reserved. Order Service (or other services) consumes this to proceed with the order flow.

---

## 1. Flow: Order Creation (Order -> Inventory)

When an order is created, the Order Service must publish a message so the Inventory Service can deduct/reserve the required stock.

### RabbitMQ Configuration
- **Exchange**: `order.events` (Topic)
- **Routing Key**: `order.created`
- **Queue (Inventory Service listens on)**: `inventory.order_events`

### JSON Structure (Payload)

The Order Service needs to send the following JSON payload when an order is created:

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

*Note: The Inventory Service extracts `id` (as orderId) and the `items` array to perform the reservation.*

---

## 2. Flow: Stock Reservation Notification (Inventory -> Order/Others)

When stock is successfully reserved for an order, the Inventory Service will publish an event for the next service in the saga (e.g., Order Service or Payment Service) to process.

### RabbitMQ Configuration
- **Exchange**: `inventory.events` (Topic)
- **Routing Key**: `inventory.reserved`
- **Queue**: (The interested service should create and bind its own queue to the exchange above with the routing key `inventory.reserved`).

### JSON Structure (Payload)

The Inventory Service will publish the following JSON payload.

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

## 3. Flow: Stock Reservation Failure (Inventory -> Order/Others)

When stock cannot be reserved (e.g., insufficient stock or missing product), the Inventory Service will publish a failure event to trigger compensating transactions in the saga.

### RabbitMQ Configuration
- **Exchange**: `inventory.events` (Topic)
- **Routing Key**: `inventory.failed`

### JSON Structure (Payload)

```json
{
  "orderId": "string",
  "reason": "string (Error message)",
  "timestamp": "string (ISO 8601)"
}
```

## Best Practices for Implementation

1. **Idempotency**: The Inventory Service should ideally handle duplicate `order.created` messages to prevent double-deducting stock.
2. **Error Handling**: Business logic failures (like insufficient stock) trigger an `inventory.failed` event and acknowledge the original message. Technical failures (like DB connection) should ideally trigger a retry mechanism.
3. **Transactional Integrity**: Ensure that database updates and message publishing are handled reliably (e.g., using the transactional outbox pattern).
