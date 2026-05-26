# Order Service

The **Order Service** is a core microservice in the Hong Commerce platform responsible for managing the lifecycle of customer orders. It implements a **Choreography-based Saga pattern** to maintain data consistency across distributed services (Payment and Inventory).

## Features

- **Multi-item Orders**: Supports multiple products within a single order.
- **Saga Pattern**: Handles complex distributed transactions via RabbitMQ events.
- **Automated Timeouts**: Background worker automatically cancels stale `PENDING` orders.
- **Resilient Connectivity**: Built-in retry logic for Database and RabbitMQ connections.

## API Specification

### Create Order

`POST /`

**Request Body:**

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "items": [
    {
      "product_id": "a1b2c3d4-e5f6-4g7h-8i9j-k0l1m2n3o4p5",
      "product_name": "Wireless Mouse",
      "quantity": 2,
      "unit_price": 25.5
    },
    {
      "product_id": "b2c3d4e5-f6g7-h8i9-j0k1-l2m3n4o5p6q7",
      "product_name": "Mechanical Keyboard",
      "quantity": 1,
      "unit_price": 89.0
    }
  ]
}
```

**Response (201 Created):**

```json
{
  "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "items": [...],
  "total_amount": 140.00,
  "status": "PENDING",
  "payment_confirmed": false,
  "inventory_confirmed": false,
  "created_at": "2023-10-27T10:00:00Z",
  "updated_at": "2023-10-27T10:00:00Z",
  "confirmed_at": null
}
```

### List Orders

`GET /`

Returns an array of all orders including their nested items.

### Get Order by ID

`GET /:id`

Returns a single order by its UUID.

---

## Saga & Event Flow

The Order Service participates in a choreography-based Saga.

### 1. Events Published

- **`order.created`**: Triggered immediately after the order is saved to the DB with `PENDING` status.
  - **Exchange**: `order.events` (fanout/topic)
  - **Payload**: `order_id`, `user_id`, `items[]`, `total_amount`.

### 2. Events Consumed

Listening on exchange `order.events`:

- **`payment.succeeded`**: Sets `payment_confirmed` to `true` and stores `payment_id`.
- **`inventory.reserved`**: Sets `inventory_confirmed` to `true`.
- **`payment.failed`**: Transitions order status to `FAILED`.
- **`inventory.failed`**: Transitions order status to `FAILED`.

_Note: When both `payment_confirmed` and `inventory_confirmed` are true, the service automatically updates the status to `CONFIRMED` via internal hooks._

---

## Database Schema

### Orders Table

| Field                 | Type        | Description                                   |
| :-------------------- | :---------- | :-------------------------------------------- |
| `id`                  | UUID        | Primary Key                                   |
| `user_id`             | UUID        | Authenticated User ID                         |
| `status`              | VARCHAR     | `PENDING`, `CONFIRMED`, `FAILED`, `CANCELLED` |
| `total_amount`        | NUMERIC     | Grand total of all items                      |
| `payment_confirmed`   | BOOLEAN     | Payment status flag                           |
| `inventory_confirmed` | BOOLEAN     | Inventory status flag                         |
| `payment_id`          | UUID        | Nullable; ID from Payment Service             |
| `confirmed_at`        | TIMESTAMPTZ | Nullable; Set when Saga completes             |

### Order Items Table

| Field          | Type    | Description               |
| :------------- | :------ | :------------------------ |
| `id`           | UUID    | Primary Key               |
| `order_id`     | UUID    | Foreign Key to Orders     |
| `product_id`   | UUID    | Product Identifier        |
| `product_name` | VARCHAR | Denormalized product name |
| `unit_price`   | NUMERIC | Price at time of order    |
| `quantity`     | INTEGER | Amount ordered            |
| `subtotal`     | NUMERIC | `unit_price * quantity`   |

---

## Configuration (Environment Variables)

| Variable       | Default          | Description                  |
| :------------- | :--------------- | :--------------------------- |
| `PORT`         | `8083`           | Service port                 |
| `DATABASE_URL` | `postgres://...` | PostgreSQL connection string |
| `MQ_URL`       | `amqp://...`     | RabbitMQ connection string   |

## Development

### Running the service

```bash
go run cmd/main.go
```

The service will automatically run migrations and start background workers for event processing and housekeeping.
