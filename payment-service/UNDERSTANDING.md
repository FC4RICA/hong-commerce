# Understanding of Payment Service PRD

This document captures my understanding of the `PRD.md` for the Payment Service, including identified gaps between the PRD and the current system implementation.

## 1. System Architecture & Workflow
The core workflow is event-driven combined with a synchronous REST API:
1. **Receive (Async):** Consume `order.created` (example) from RabbitMQ, containing `orderID` and `amount`.
2. **Initialize:** Create a `PENDING` payment record in the Payment DB.
3. **Update (Sync):** An external client/system calls `PATCH /api/v1/payments/{paymentID}/status` to update the status to `SUCCESS` or `FAILED`.
4. **Notify (Async):** On `SUCCESS`, publish a `payment.completed` (example) message to RabbitMQ to notify the Order Service.

## 2. API & Payload Specifications (PRD vs. Current Implementation)

### 2.1 API Endpoint
* **PRD:** `PATCH /api/v1/payments/{paymentID}/status`
* **Current Implementation:** `PATCH /api/v1/payments/{id}`
* *Gap:* The path in the code doesn't include `/status`.

### 2.2 API Request Body
* **PRD:** `{"status": "SUCCESS", "transactionRef": "TXN-000001", "amountPaid": 1500.00}`
* **Current Implementation:** `{"success": true}`
* *Gap:* The PRD requires explicit passing of `status`, `transactionRef`, and `amountPaid` for auditing and validation, whereas the current code only takes a boolean flag.

### 2.3 Status Terminology
* **PRD:** `PENDING`, `SUCCESS`, `FAILED`
* **Current Implementation:** `PENDING`, `COMPLETED`, `FAILED`, `REFUNDED`
* *Gap:* `COMPLETED` vs `SUCCESS`.

### 2.4 RabbitMQ Messages
* **Consumer (Order -> Payment):** PRD expects `orderID`, `amount`, `currency`, `timestamp`.
* **Publisher (Payment -> Order):** PRD expects `paymentID`, `orderID`, `status`, `paidAt`.
* *Gap:* The current code publishes `user_id` and `amount`, but misses `paidAt`.

## 3. Non-Functional Requirements
1. **Idempotency:** Both the RabbitMQ consumer and the API endpoint must safely handle duplicate requests without side effects.
2. **Database Isolation:** Strict separation of the Payment DB.
3. **DLQ (Dead Letter Queue):** Message retries must eventually route to a DLQ if they continually fail.

## Conclusion
The PRD introduces more rigid auditing fields (`transactionRef`, `amountPaid`) and a slightly different terminology (`SUCCESS` instead of `COMPLETED`). We need to align the current codebase with the PRD through a grilling session.
