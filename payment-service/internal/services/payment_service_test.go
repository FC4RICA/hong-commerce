package services

import (
	"context"
	"errors"
	"testing"

	"github.com/FC4RICA/hong-commerce/payment-service/internal/entities"
	"github.com/google/uuid"
)

// --- Mocks ---

type mockPaymentRepository struct {
	checkIdempotencyKeyFn          func(key string) (bool, error)
	createPaymentWithIdempotencyFn func(payment *entities.Payment, idempotencyKey string) error
	updatePaymentWithIdempotencyFn func(payment *entities.Payment, idempotencyKey string) error
	getPaymentByIDFn               func(id string) (*entities.Payment, error)
}

func (m *mockPaymentRepository) CheckIdempotencyKey(key string) (bool, error) {
	if m.checkIdempotencyKeyFn != nil {
		return m.checkIdempotencyKeyFn(key)
	}
	return false, nil
}

func (m *mockPaymentRepository) CreatePaymentWithIdempotency(payment *entities.Payment, idempotencyKey string) error {
	if m.createPaymentWithIdempotencyFn != nil {
		return m.createPaymentWithIdempotencyFn(payment, idempotencyKey)
	}
	return nil
}

func (m *mockPaymentRepository) UpdatePaymentWithIdempotency(payment *entities.Payment, idempotencyKey string) error {
	if m.updatePaymentWithIdempotencyFn != nil {
		return m.updatePaymentWithIdempotencyFn(payment, idempotencyKey)
	}
	return nil
}

func (m *mockPaymentRepository) GetPaymentByID(id string) (*entities.Payment, error) {
	if m.getPaymentByIDFn != nil {
		return m.getPaymentByIDFn(id)
	}
	return nil, ErrPaymentNotFound
}

type mockPaymentPublisher struct {
	publishPaymentCompletedFn func(ctx context.Context, event PaymentCompletedEvent) error
}

func (m *mockPaymentPublisher) PublishPaymentCompleted(ctx context.Context, event PaymentCompletedEvent) error {
	if m.publishPaymentCompletedFn != nil {
		return m.publishPaymentCompletedFn(ctx, event)
	}
	return nil
}

// --- Tests ---

func TestInitializePayment_Success(t *testing.T) {
	repo := &mockPaymentRepository{
		checkIdempotencyKeyFn: func(key string) (bool, error) {
			return false, nil // Not processed yet
		},
		createPaymentWithIdempotencyFn: func(payment *entities.Payment, idempotencyKey string) error {
			if payment.Status != "PENDING" {
				t.Errorf("Expected status PENDING, got %s", payment.Status)
			}
			return nil
		},
	}
	pub := &mockPaymentPublisher{}
	service := NewPaymentService(repo, pub)

	err := service.InitializePayment("ORDER-123", "THB", 1500.00)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestInitializePayment_IdempotentAlreadyProcessed(t *testing.T) {
	repo := &mockPaymentRepository{
		checkIdempotencyKeyFn: func(key string) (bool, error) {
			return true, nil // Already processed!
		},
		createPaymentWithIdempotencyFn: func(payment *entities.Payment, idempotencyKey string) error {
			t.Fatal("CreatePaymentWithIdempotency should not be called if already processed")
			return nil
		},
	}
	service := NewPaymentService(repo, &mockPaymentPublisher{})

	err := service.InitializePayment("ORDER-123", "THB", 1500.00)
	if err != nil {
		t.Fatalf("Expected no error (graceful return), got: %v", err)
	}
}

func TestUpdateStatus_Success(t *testing.T) {
	paymentID := uuid.New()
	mockPayment := &entities.Payment{
		ID:       paymentID,
		OrderID:  "ORDER-123",
		Amount:   1500.00,
		Currency: "THB",
		Status:   "PENDING",
	}

	eventPublished := false

	repo := &mockPaymentRepository{
		checkIdempotencyKeyFn: func(key string) (bool, error) {
			return false, nil
		},
		getPaymentByIDFn: func(id string) (*entities.Payment, error) {
			if id == paymentID.String() {
				return mockPayment, nil
			}
			return nil, ErrPaymentNotFound
		},
		updatePaymentWithIdempotencyFn: func(payment *entities.Payment, idempotencyKey string) error {
			if payment.Status != "COMPLETED" {
				t.Errorf("Expected status COMPLETED, got %s", payment.Status)
			}
			return nil
		},
	}
	pub := &mockPaymentPublisher{
		publishPaymentCompletedFn: func(ctx context.Context, event PaymentCompletedEvent) error {
			eventPublished = true
			if event.PaymentID != paymentID.String() {
				t.Errorf("Expected published PaymentID %s, got %s", paymentID.String(), event.PaymentID)
			}
			if event.Status != "COMPLETED" {
				t.Errorf("Expected published status COMPLETED, got %s", event.Status)
			}
			return nil
		},
	}
	service := NewPaymentService(repo, pub)

	payment, idempotent, err := service.UpdateStatus(context.Background(), paymentID.String(), "COMPLETED", "TXN-123", 1500.00)
	
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if idempotent {
		t.Errorf("Expected idempotent to be false")
	}
	if payment.Status != "COMPLETED" {
		t.Errorf("Expected returned payment status to be COMPLETED")
	}
	if !eventPublished {
		t.Errorf("Expected PaymentCompletedEvent to be published")
	}
}

func TestUpdateStatus_AlreadyProcessedPayment(t *testing.T) {
	paymentID := uuid.New()
	mockPayment := &entities.Payment{
		ID:       paymentID,
		OrderID:  "ORDER-123",
		Amount:   1500.00,
		Status:   "COMPLETED", // Already processed!
	}

	repo := &mockPaymentRepository{
		checkIdempotencyKeyFn: func(key string) (bool, error) {
			return false, nil
		},
		getPaymentByIDFn: func(id string) (*entities.Payment, error) {
			return mockPayment, nil
		},
	}
	service := NewPaymentService(repo, &mockPaymentPublisher{})

	_, _, err := service.UpdateStatus(context.Background(), paymentID.String(), "COMPLETED", "TXN-123", 1500.00)
	
	if !errors.Is(err, ErrPaymentAlreadyProcessed) {
		t.Fatalf("Expected ErrPaymentAlreadyProcessed, got: %v", err)
	}
}

func TestUpdateStatus_IdempotentAlreadyProcessed(t *testing.T) {
	paymentID := uuid.New()
	mockPayment := &entities.Payment{
		ID:       paymentID,
		Status:   "COMPLETED",
	}

	repo := &mockPaymentRepository{
		checkIdempotencyKeyFn: func(key string) (bool, error) {
			return true, nil // Webhook already processed for this TXN!
		},
		getPaymentByIDFn: func(id string) (*entities.Payment, error) {
			return mockPayment, nil
		},
		updatePaymentWithIdempotencyFn: func(payment *entities.Payment, idempotencyKey string) error {
			t.Fatal("Should not attempt to update DB if idempotent request")
			return nil
		},
	}
	service := NewPaymentService(repo, &mockPaymentPublisher{})

	payment, idempotent, err := service.UpdateStatus(context.Background(), paymentID.String(), "COMPLETED", "TXN-123", 1500.00)
	
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !idempotent {
		t.Errorf("Expected idempotent to be true")
	}
	if payment == nil {
		t.Fatalf("Expected payment to be returned")
	}
}
