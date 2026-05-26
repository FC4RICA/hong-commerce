package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/FC4RICA/hong-commerce/payment-service/internal/entities"
	"github.com/FC4RICA/hong-commerce/payment-service/internal/repositories"
)

var (
	ErrPaymentNotFound         = errors.New("payment not found")
	ErrPaymentAlreadyProcessed = errors.New("payment is already processed")
)

type PaymentCompletedEvent struct {
	PaymentID string `json:"paymentID"`
	OrderID   string `json:"orderID"`
	Status    string `json:"status"`
	PaidAt    string `json:"paidAt"`
}

type PaymentPublisher interface {
	PublishPaymentCompleted(ctx context.Context, event PaymentCompletedEvent) error
}

type PaymentService interface {
	InitializePayment(orderID, currency string, amount float64) error
	UpdateStatus(ctx context.Context, paymentID, status, transactionRef string, amountPaid float64) (*entities.Payment, bool, error)
}

type paymentServiceImpl struct {
	repo      repositories.PaymentRepository
	publisher PaymentPublisher
}

func NewPaymentService(repo repositories.PaymentRepository, publisher PaymentPublisher) PaymentService {
	return &paymentServiceImpl{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *paymentServiceImpl) InitializePayment(orderID, currency string, amount float64) error {
	idemKey := fmt.Sprintf("order.created:%s", orderID)

	processed, err := s.repo.CheckIdempotencyKey(idemKey)
	if err != nil {
		return fmt.Errorf("failed idempotency check: %w", err)
	}
	if processed {
		// Already processed, just return gracefully
		return nil
	}

	payment := &entities.Payment{
		OrderID:  orderID,
		Amount:   amount,
		Currency: currency,
		Status:   "PENDING",
	}

	if err := s.repo.CreatePaymentWithIdempotency(payment, idemKey); err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	return nil
}

func (s *paymentServiceImpl) UpdateStatus(ctx context.Context, paymentID, status, transactionRef string, amountPaid float64) (*entities.Payment, bool, error) {
	// Idempotency check for external gateway update
	if transactionRef != "" {
		idemKey := fmt.Sprintf("payment.update:%s", transactionRef)
		processed, err := s.repo.CheckIdempotencyKey(idemKey)
		if err != nil {
			return nil, false, fmt.Errorf("failed idempotency check: %w", err)
		}
		if processed {
			// Already processed this transactionRef. Fetch payment to return to caller.
			payment, err := s.repo.GetPaymentByID(paymentID)
			if err != nil {
				return nil, true, fmt.Errorf("failed to get payment for idempotent return: %w", err)
			}
			return payment, true, nil // true = idempotent response
		}
	}

	payment, err := s.repo.GetPaymentByID(paymentID)
	if err != nil {
		return nil, false, ErrPaymentNotFound
	}

	if payment.Status != "PENDING" {
		return nil, false, ErrPaymentAlreadyProcessed
	}

	payment.Status = status
	payment.TransactionRef = transactionRef

	idemKey := ""
	if transactionRef != "" {
		idemKey = fmt.Sprintf("payment.update:%s", transactionRef)
	}

	if err := s.repo.UpdatePaymentWithIdempotency(payment, idemKey); err != nil {
		return nil, false, fmt.Errorf("failed to update payment: %w", err)
	}

	// Publish domain event
	if payment.Status == "COMPLETED" {
		event := PaymentCompletedEvent{
			PaymentID: payment.ID.String(),
			OrderID:   payment.OrderID,
			Status:    payment.Status,
			PaidAt:    time.Now().Format(time.RFC3339),
		}
		if err := s.publisher.PublishPaymentCompleted(ctx, event); err != nil {
			// We swallow or log error since DB transaction succeeded (Eventual Consistency / Outbox should handle this IRL)
			fmt.Printf("Warning: failed to publish payment completed event: %v\n", err)
		}
	}

	return payment, false, nil
}
