# Payment Success Flow (Order Service)

This document explains how the **Order Service** receives and processes payment success notifications from the **Payment Service** asynchronously using RabbitMQ.

## Architecture & Infrastructure

When a payment is successfully completed, the Payment Service publishes an event. The Order Service acts as a **Consumer** for this event to update the order status accordingly.

* **Message Broker:** RabbitMQ
* **Exchange:** `payment_status_exchange` (Type: `direct`)
* **Routing Key:** `payment.success`
* **Queue:** `order.payment.succeeded`

## Event Consumption

The Order Service sets up an `EventWorker` (located in `internal/workers/event_worker.go`) that declares the necessary RabbitMQ infrastructure on startup:

1. **Exchange Declaration:** Ensures `payment_status_exchange` exists.
2. **Queue Declaration:** Creates the `order.payment.succeeded` queue.
3. **Queue Binding:** Binds the queue to the exchange using the `payment.success` routing key.

### JSON Payload Structure

The `payment-service` publishes a `PaymentCompletedEvent` JSON payload. The Order Service expects the following structure:

```json
{
  "paymentID": "123e4567-e89b-12d3-a456-426614174000",
  "orderID": "987fcdeb-51a2-43d7-9012-3456789abcde",
  "status": "COMPLETED",
  "paidAt": "2026-05-26T12:00:00Z"
}
```

## Processing Logic (`handlePaymentSucceeded`)

When a message is received in the `order.payment.succeeded` queue, the following steps happen:

1. **Unmarshalling:** The JSON body is parsed into the `PaymentCompletedEvent` struct.
2. **Validation:** The `orderID` and `paymentID` strings are parsed into Go `uuid.UUID` types. If the format is invalid, the message processing fails.
3. **Service Layer:** The worker calls `svc.HandlePaymentSucceeded(ctx, orderUUID, paymentUUID)`.
4. **Database Update:** The service layer updates the order record in the `orderdb` (typically changing the order status to `PAID` or triggering the next step in the saga pattern).
5. **Acknowledgement (ACK):** If no errors occur, the worker sends an `ACK` to RabbitMQ, removing the message from the queue. If an error occurs, it sends a `NACK` (with requeue) to try again.

## Error Handling & Idempotency

* **Requeueing:** If a transient database error occurs while updating the order status, the event is `NACK`ed and RabbitMQ will requeue it.
* **Idempotency:** The `HandlePaymentSucceeded` service method must be idempotent. If the order is already marked as paid, it should safely ignore the duplicate event to prevent side effects.
