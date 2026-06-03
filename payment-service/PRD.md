Product Requirements Document (PRD): Payment Service
1. Overview
This document outlines the product requirements for the Payment Service. The service is responsible for managing the payment processing lifecycle by integrating with an Order Service asynchronously via a message broker (RabbitMQ) and exposing an API endpoint to receive payment status updates from external sources.
2. Scope
•	In-Scope:
•	Development restricted exclusively to the Payment Service.
•	Management of the Payment Database (strictly isolated from the Order Service Database).
•	Consuming messages containing an orderID from RabbitMQ.
•	Creating a REST API endpoint to update the payment status using a paymentID.
•	Publishing messages back to RabbitMQ upon a successful payment.
•	Out-of-Scope:
•	Any internal logic or development within the Order Service.
•	Management of the Order Database.
3. System Architecture & Workflow
The operational flow of the Payment Service is as follows:
	1.	Receive: The Payment Service consumes a message from a RabbitMQ queue (e.g., order.created) containing the orderID sent by the Order Service.
	2.	Initialize: The Payment Service creates a new payment record in the Payment Database with an initial status of PENDING.
	3.	Update: An external system or client invokes the Payment Service's API endpoint, providing the paymentID to update the payment status (e.g., to SUCCESS or FAILED).
	4.	Notify: If the payment status is updated to SUCCESS, the Payment Service publishes a message to a RabbitMQ exchange/queue (e.g., payment.completed) containing the orderID to notify the Order Service.
4. Functional Requirements
4.1. RabbitMQ Consumer: Receive Order
•	Description: The system must consume messages from the Order Service indicating a new order has been created to initiate the payment process.
•	Queue Name (Example): order_to_payment_queue
•	Actions:
•	Consume the message containing the orderID (and other necessary data, such as amount).
•	Persist the data as a new record in the Payment Database.
•	Set the initial payment status to PENDING.
4.2. API: Update Payment Status
•	Description: The system must provide a REST API endpoint to receive payment results and update the status in the Payment Database.
•	Endpoint: PATCH /api/v1/payments/{paymentID}/status
•	Actions:
•	Validate the existence of the paymentID in the Payment Database.
•	Update the payment status based on the request payload (e.g., SUCCESS, FAILED).
•	Trigger: If the status updates to SUCCESS, immediately invoke the Message Publisher function (Section 4.3).
4.3. RabbitMQ Publisher: Payment Success Notification
•	Description: The system must notify the Order Service once a payment has been successfully completed.
•	Exchange / Queue (Example): payment_status_exchange / Routing Key: payment.success
•	Actions:
•	Construct a message payload containing the orderID and the payment status.
•	Publish the message to RabbitMQ for the Order Service to process.
5. Interface Specifications
5.1. RabbitMQ Message Contracts
1. Consuming Message (Received from Order Service)
{
  "orderID": "ORD-123456789",
  "amount": 1500.00,
  "currency": "THB",
  "timestamp": "2026-05-25T01:00:00Z"
}

2. Publishing Message (Sent to Order Service upon success)
{
  "paymentID": "PAY-987654321",
  "orderID": "ORD-123456789",
  "status": "SUCCESS",
  "paidAt": "2026-05-25T01:15:00Z"
}

5.2. REST API Contract
Update Payment Status Endpoint
•	Method: PATCH
•	Path: /api/v1/payments/{paymentID}/status
•	Request Body:
{
  "status": "SUCCESS",
  "transactionRef": "TXN-000001",
  "amountPaid": 1500.00
}

•	Response (200 OK):
{
  "message": "Payment status updated successfully",
  "paymentID": "PAY-987654321",
  "status": "SUCCESS"
}

6. Data Model (Payment Database)
To adhere to microservices principles, the Payment Database must only store data relevant to payment processing. It links back to the original order strictly via the orderID.
Table: payments
Column Name	Type	Description
id	UUID / String	Primary Key (Payment ID)
order_id	String	Foreign reference to the Order ID from the Order Service
amount	Decimal	The payment amount required
status	Enum / String	Current payment status (PENDING, SUCCESS, FAILED)
transaction_ref	String	External reference ID from the Payment Gateway (if applicable)
created_at	Timestamp	Record creation timestamp
updated_at	Timestamp	Last record update timestamp
7. Non-Functional Requirements
•	Idempotency: Both the RabbitMQ consumer and the API endpoint must be idempotent. If an update or a message retry occurs due to network issues (e.g., receiving a SUCCESS update for an already successful payment), the system must handle it gracefully without throwing errors or dispatching duplicate success messages.
•	Database Isolation: The Payment Service will solely connect to and perform transactions on the Payment Database. Cross-database queries to the Order Database are strictly prohibited.
•	Error Handling & Dead Letter Queue (DLQ): If the service fails to process a RabbitMQ message after a predefined number of retries, the message should be routed to a Dead Letter Queue (DLQ) for manual inspection and troubleshooting.
