# RabbitMQ Integration: Order Service <-> Payment Service

This document outlines how the **Order Service** and **Payment Service** integrate via RabbitMQ asynchronously.

## Overview

The integration involves two main flows:
1. **Order Creation**: Order Service publishes a message when an order is created. Payment Service consumes this to initialize the payment record.
2. **Payment Notification**: Payment Service publishes a message when a payment is successful. Order Service consumes this to update the order status.

---

## 1. Flow: Order Creation (Order -> Payment)

When an order is created, the Order Service must publish a message so the Payment Service can prepare the payment record.

### RabbitMQ Configuration
- **Exchange**: (Depends on Order Service configuration, e.g., `order_exchange`)
- **Queue (Payment Service listens on)**: `order_to_payment_queue`

### JSON Structure (Payload)

The Order Service needs to send the following JSON payload when an order is created:

```json
{
  "orderID": "string",
  "amount": 123.45,
  "currency": "string",
  "timestamp": "string (RFC3339 format, e.g., 2026-05-26T13:55:19+07:00)"
}
```

*Note: The Payment Service uses an idempotency key `order.created:<orderID>` to prevent processing duplicate messages.*

---

## 2. Flow: Payment Notification (Payment -> Order)

When the payment status is updated to `COMPLETED` (e.g., via a webhook from an external payment gateway), the Payment Service will publish an event for the Order Service to process.

### RabbitMQ Configuration
- **Exchange**: `payment_status_exchange`
- **Routing Key**: `payment.success`
- **Queue**: (The Order Service should create and bind its own queue to the exchange above with the routing key `payment.success`).

### JSON Structure (Payload)

The Payment Service will publish the following JSON payload. The Order Service should parse this to update the order.

```json
{
  "paymentID": "string (UUID)",
  "orderID": "string",
  "status": "COMPLETED",
  "paidAt": "string (RFC3339 format)"
}
```

## Best Practices for Implementation

1. **Idempotency**: Both services must safely handle duplicate messages. Implement idempotency keys (like `orderID` and `paymentID`) to prevent side effects on retries.
2. **Error Handling & DLQ**: If a message fails to process after retries, route it to a Dead Letter Queue (DLQ) rather than dropping it.
3. **Eventual Consistency**: Expect a slight delay between an order being paid and the order status updating in the Order Service database.
