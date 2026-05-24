package repositories

import (
	"testing"

	"github.com/FC4RICA/hong-commerce/payment-service/internal/entities"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to in-memory database: %v", err)
	}

	err = db.AutoMigrate(&entities.Payment{}, &entities.IdempotencyKey{})
	if err != nil {
		t.Fatalf("Failed to auto-migrate: %v", err)
	}

	return db
}

func TestCheckIdempotencyKey(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPaymentRepository(db)

	key := "test-idem-key-1"

	// 1. Check a non-existent key
	exists, err := repo.CheckIdempotencyKey(key)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if exists {
		t.Errorf("Expected key to not exist")
	}

	// 2. Insert the key manually
	err = db.Create(&entities.IdempotencyKey{
		ID:  uuid.New(),
		Key: key,
	}).Error
	if err != nil {
		t.Fatalf("Failed to insert mock idempotency key: %v", err)
	}

	// 3. Check again, it should exist now
	exists, err = repo.CheckIdempotencyKey(key)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !exists {
		t.Errorf("Expected key to exist")
	}
}

func TestCreatePaymentWithIdempotency(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPaymentRepository(db)

	payment := &entities.Payment{
		OrderID:  "ORD-123",
		Amount:   150.50,
		Currency: "USD",
		Status:   "PENDING",
	}
	idemKey := "order.created:ORD-123"

	err := repo.CreatePaymentWithIdempotency(payment, idemKey)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify payment was created with a populated UUID
	if payment.ID == uuid.Nil {
		t.Errorf("Expected Payment ID to be populated by GORM hook")
	}

	// Verify idempotency key was inserted
	var keyRecord entities.IdempotencyKey
	err = db.Where("key = ?", idemKey).First(&keyRecord).Error
	if err != nil {
		t.Errorf("Expected idempotency key to be found, got error: %v", err)
	}
}

func TestUpdatePaymentWithIdempotency(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPaymentRepository(db)

	// Seed an existing payment
	payment := &entities.Payment{
		OrderID: "ORD-999",
		Amount:  100.00,
		Status:  "PENDING",
	}
	if err := db.Create(payment).Error; err != nil {
		t.Fatalf("Failed to seed payment: %v", err)
	}

	// Perform update
	payment.Status = "COMPLETED"
	payment.TransactionRef = "TXN-999"
	idemKey := "payment.update:TXN-999"

	err := repo.UpdatePaymentWithIdempotency(payment, idemKey)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify the update in DB
	var updatedPayment entities.Payment
	db.First(&updatedPayment, payment.ID)
	if updatedPayment.Status != "COMPLETED" {
		t.Errorf("Expected status COMPLETED, got %s", updatedPayment.Status)
	}
	if updatedPayment.TransactionRef != "TXN-999" {
		t.Errorf("Expected tx ref TXN-999, got %s", updatedPayment.TransactionRef)
	}

	// Verify idempotency key was inserted
	exists, _ := repo.CheckIdempotencyKey(idemKey)
	if !exists {
		t.Errorf("Expected idempotency key to be inserted during update")
	}
}

func TestGetPaymentByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPaymentRepository(db)

	payment := &entities.Payment{
		OrderID: "ORD-888",
		Status:  "PENDING",
	}
	db.Create(payment)

	fetched, err := repo.GetPaymentByID(payment.ID.String())
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if fetched.OrderID != "ORD-888" {
		t.Errorf("Expected fetched OrderID to be ORD-888, got %s", fetched.OrderID)
	}

	// Test Not Found
	_, err = repo.GetPaymentByID(uuid.New().String())
	if err == nil {
		t.Errorf("Expected an error for non-existent payment, got nil")
	}
}
