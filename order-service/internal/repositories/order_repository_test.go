package repositories_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/FC4RICA/hong-commerce/order-service/internal/models"
	"github.com/FC4RICA/hong-commerce/order-service/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestOrderRepository(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		// Default to standard local compose DSN
		dsn = "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping repository integration tests (database connection failed): %v", err)
	}

	// Clean up / Setup tables
	_ = db.AutoMigrate(&models.Order{}, &models.OrderItem{})

	// Always clear records before testing to ensure isolation
	db.Exec("TRUNCATE TABLE order_items CASCADE")
	db.Exec("TRUNCATE TABLE orders CASCADE")

	repo := repositories.NewOrderRepository(db)
	ctx := context.Background()

	t.Run("Create and Get Order", func(t *testing.T) {
		userID := uuid.New()
		order := &models.Order{
			UserID:      userID,
			Status:      "PENDING",
			TotalAmount: 120.50,
			Items: []models.OrderItem{
				{
					ProductID:   uuid.New(),
					ProductName: "Mouse",
					Quantity:    1,
					UnitPrice:   20.50,
				},
				{
					ProductID:   uuid.New(),
					ProductName: "Keyboard",
					Quantity:    1,
					UnitPrice:   100.00,
				},
			},
		}

		err := repo.Create(ctx, order)
		if err != nil {
			t.Fatalf("failed to create order: %v", err)
		}

		if order.ID == uuid.Nil {
			t.Fatal("expected order ID to be populated by GORM")
		}

		fetched, err := repo.GetByID(ctx, order.ID)
		if err != nil {
			t.Fatalf("failed to get order by ID: %v", err)
		}

		if fetched.UserID != userID {
			t.Errorf("expected UserID %s, got %s", userID, fetched.UserID)
		}

		if len(fetched.Items) != 2 {
			t.Errorf("expected 2 preloaded items, got %d", len(fetched.Items))
		}
	})

	t.Run("GetAll Orders", func(t *testing.T) {
		orders, err := repo.GetAll(ctx)
		if err != nil {
			t.Fatalf("failed to get all orders: %v", err)
		}

		if len(orders) == 0 {
			t.Error("expected at least 1 order, got 0")
		}
	})

	t.Run("Update Order Status", func(t *testing.T) {
		userID := uuid.New()
		order := &models.Order{
			UserID:      userID,
			Status:      "PENDING",
			TotalAmount: 50.00,
		}

		_ = repo.Create(ctx, order)

		order.Status = "CONFIRMED"
		err := repo.Update(ctx, order)
		if err != nil {
			t.Fatalf("failed to update order: %v", err)
		}

		fetched, _ := repo.GetByID(ctx, order.ID)
		if fetched.Status != "CONFIRMED" {
			t.Errorf("expected status CONFIRMED, got %s", fetched.Status)
		}
	})

	t.Run("GetExpiredPendingOrders", func(t *testing.T) {
		// Truncate first to isolate this test
		db.Exec("TRUNCATE TABLE order_items CASCADE")
		db.Exec("TRUNCATE TABLE orders CASCADE")

		// Create one order that is expired (created 20 mins ago)
		expiredOrder := &models.Order{
			UserID:      uuid.New(),
			Status:      "PENDING",
			TotalAmount: 10.00,
		}
		_ = repo.Create(ctx, expiredOrder)
		// Manually backdate created_at
		db.Model(expiredOrder).Update("created_at", time.Now().Add(-20*time.Minute))

		// Create one order that is fresh (created now)
		freshOrder := &models.Order{
			UserID:      uuid.New(),
			Status:      "PENDING",
			TotalAmount: 20.00,
		}
		_ = repo.Create(ctx, freshOrder)

		// Create one order that is confirmed (but old)
		confirmedOldOrder := &models.Order{
			UserID:      uuid.New(),
			Status:      "CONFIRMED",
			TotalAmount: 30.00,
		}
		_ = repo.Create(ctx, confirmedOldOrder)
		db.Model(confirmedOldOrder).Update("created_at", time.Now().Add(-20*time.Minute))

		expired, err := repo.GetExpiredPendingOrders(ctx, 15*time.Minute)
		if err != nil {
			t.Fatalf("failed to get expired pending orders: %v", err)
		}

		if len(expired) != 1 {
			t.Fatalf("expected exactly 1 expired pending order, got %d", len(expired))
		}

		if expired[0].ID != expiredOrder.ID {
			t.Errorf("expected expired order ID %s, got %s", expiredOrder.ID, expired[0].ID)
		}
	})
}
