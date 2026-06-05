# Inventory Service API Specification

The Inventory Service manages product stock levels, handles reservations for orders, and provides a RESTful API for item management. It also integrates asynchronously via RabbitMQ to support distributed transactions (Saga pattern).

## Base URL
`http://localhost:3000` (Default)

---

## REST API (Synchronous)

All endpoints prefixed with `/items`.

### Data Models

#### Item Object
| Property | Type | Description |
| :--- | :--- | :--- |
| `id` | `string` | Unique identifier (CUID) |
| `name` | `string` | Name of the item |
| `quantity` | `integer` | Total stock quantity |
| `reserved` | `integer` | Quantity currently reserved for pending orders |
| `available` | `integer` | Available stock (`quantity - reserved`) |
| `createdAt` | `string` | ISO 8601 timestamp |
| `updatedAt` | `string` | ISO 8601 timestamp |

---

### List Items
Retrieve all items in the inventory.

- **URL:** `/items`
- **Method:** `GET`
- **Response Body:** `Item[]`

---

### Create Item
Add a new item to the inventory.

- **URL:** `/items`
- **Method:** `POST`
- **Request Body:**
  ```json
  {
    "name": "string",
    "quantity": "integer",
    "reserved": "integer (optional, default 0)"
  }
  ```
- **Success Response:** `201 Created`
- **Response Body:** `Item`

---

### Get Item by ID
Retrieve details for a specific item.

- **URL:** `/items/:id`
- **Method:** `GET`
- **Success Response:** `200 OK`
- **Error Response:** `404 Not Found` if item doesn't exist.

---

### Update Item
Modify an existing item's properties.

- **URL:** `/items/:id`
- **Method:** `PUT`
- **Request Body (Partial update supported):**
  ```json
  {
    "name": "string (optional)",
    "quantity": "integer (optional)",
    "reserved": "integer (optional)"
  }
  ```
- **Success Response:** `200 OK`
- **Error Response:** `404 Not Found` if item doesn't exist.

---

### Delete Item
Remove an item from the inventory.

- **URL:** `/items/:id`
- **Method:** `DELETE`
- **Success Response:** `200 OK`
- **Error Response:** `404 Not Found` if item doesn't exist.

---

## Messaging API (Asynchronous)

The Inventory Service uses RabbitMQ for inter-service communication, specifically for the Order-to-Inventory stock reservation flow.

### Subscribed Events

#### `order.created`
Consumed when a new order is placed. The service attempts to reserve stock for the items in the order.

- **Exchange:** `order.events` (Topic)
- **Routing Key:** `order.created`
- **Queue:** `inventory.order_events`
- **Payload:**
  ```json
  {
    "id": "string (OrderID)",
    "items": [
      {
        "id": "string (ProductID)",
        "quantity": "integer"
      }
    ]
  }
  ```

---

### Published Events

#### `inventory.reserved`
Published after successful stock reservation for an order.

- **Exchange:** `inventory.events` (Topic)
- **Routing Key:** `inventory.reserved`
- **Payload:**
  ```json
  {
    "orderId": "string",
    "items": [
      {
        "id": "string",
        "quantity": "integer"
      }
    ]
  }
  ```

#### `inventory.failed`
Published if stock reservation fails (e.g., insufficient stock, item not found). Triggers compensation in the Order Service.

- **Exchange:** `inventory.events` (Topic)
- **Routing Key:** `inventory.failed`
- **Payload:**
  ```json
  {
    "orderId": "string",
    "reason": "string (Error message)",
    "timestamp": "string (ISO 8601)"
  }
  ```

---

## Error Handling

The service uses standard HTTP status codes:
- `400 Bad Request`: Validation error (e.g., negative quantity, empty name).
- `404 Not Found`: Resource does not exist.
- `503 Service Unavailable`: Database or Message Broker connection issue.

Validation errors return the following structure:
```json
{
  "error": "Error message description"
}
```
