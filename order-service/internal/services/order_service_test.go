package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/FC4RICA/hong-commerce/order-service/internal/models"
	"github.com/FC4RICA/hong-commerce/order-service/internal/services"
	"github.com/google/uuid"
)

type mockOrderRepository struct {
	CreateFunc                  func(ctx context.Context, order *models.Order) error
	GetAllFunc                  func(ctx context.Context) ([]models.Order, error)
	GetByIDFunc                 func(ctx context.Context, id uuid.UUID) (*models.Order, error)
	UpdateFunc                  func(ctx context.Context, order *models.Order) error
	GetExpiredPendingOrdersFunc func(ctx context.Context, timeout time.Duration) ([]models.Order, error)
}

func (m *mockOrderRepository) Create(ctx context.Context, order *models.Order) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, order)
	}
	return nil
}

func (m *mockOrderRepository) GetAll(ctx context.Context) ([]models.Order, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx)
	}
	return nil, nil
}

func (m *mockOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockOrderRepository) Update(ctx context.Context, order *models.Order) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, order)
	}
	return nil
}

func (m *mockOrderRepository) GetExpiredPendingOrders(ctx context.Context, timeout time.Duration) ([]models.Order, error) {
	if m.GetExpiredPendingOrdersFunc != nil {
		return m.GetExpiredPendingOrdersFunc(ctx, timeout)
	}
	return nil, nil
}

func TestCreateOrder(t *testing.T) {
	mockRepo := &mockOrderRepository{
		CreateFunc: func(ctx context.Context, order *models.Order) error {
			order.ID = uuid.New()
			return nil
		},
	}

	svc := services.NewOrderService(mockRepo, nil, 15)

	userID := uuid.New()
	productID1 := uuid.New()
	productID2 := uuid.New()

	items := []models.OrderItem{
		{
			ProductID:   productID1,
			ProductName: "Item A",
			Quantity:    2,
			UnitPrice:   15.50,
		},
		{
			ProductID:   productID2,
			ProductName: "Item B",
			Quantity:    1,
			UnitPrice:   45.00,
		},
	}

	order, err := svc.CreateOrder(context.Background(), userID, items)
	if err != nil {
		t.Fatalf("CreateOrder returned unexpected error: %v", err)
	}

	if order == nil {
		t.Fatal("expected order to be non-nil")
	}

	expectedTotal := 76.00 // (2 * 15.50) + (1 * 45.00) = 31 + 45 = 76
	if order.TotalAmount != expectedTotal {
		t.Errorf("expected total amount %f, got %f", expectedTotal, order.TotalAmount)
	}

	if order.Status != "PENDING" {
		t.Errorf("expected order status to be PENDING, got %s", order.Status)
	}

	if len(order.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(order.Items))
	}
}

func TestGetOrderByID(t *testing.T) {
	orderID := uuid.New()
	expectedOrder := &models.Order{
		ID:     orderID,
		Status: "PENDING",
	}

	mockRepo := &mockOrderRepository{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Order, error) {
			if id == orderID {
				return expectedOrder, nil
			}
			return nil, errors.New("not found")
		},
	}

	svc := services.NewOrderService(mockRepo, nil, 15)

	order, err := svc.GetOrderByID(context.Background(), orderID)
	if err != nil {
		t.Fatalf("GetOrderByID returned error: %v", err)
	}

	if order.ID != orderID {
		t.Errorf("expected order ID %s, got %s", orderID, order.ID)
	}
}

func TestHandlePaymentSucceeded(t *testing.T) {
	orderID := uuid.New()
	paymentID := uuid.New()
	existingOrder := &models.Order{
		ID:               orderID,
		Status:           "PENDING",
		PaymentConfirmed: false,
	}

	var updatedOrder *models.Order
	mockRepo := &mockOrderRepository{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Order, error) {
			return existingOrder, nil
		},
		UpdateFunc: func(ctx context.Context, order *models.Order) error {
			updatedOrder = order
			return nil
		},
	}

	svc := services.NewOrderService(mockRepo, nil, 15)

	err := svc.HandlePaymentSucceeded(context.Background(), orderID, paymentID)
	if err != nil {
		t.Fatalf("HandlePaymentSucceeded returned error: %v", err)
	}

	if updatedOrder == nil {
		t.Fatal("expected update to be called")
	}

	if !updatedOrder.PaymentConfirmed {
		t.Error("expected PaymentConfirmed to be true")
	}

	if updatedOrder.PaymentID == nil || *updatedOrder.PaymentID != paymentID {
		t.Error("expected PaymentID to match")
	}
}

func TestHandleInventoryReserved(t *testing.T) {
	orderID := uuid.New()
	existingOrder := &models.Order{
		ID:                 orderID,
		Status:             "PENDING",
		InventoryConfirmed: false,
	}

	var updatedOrder *models.Order
	mockRepo := &mockOrderRepository{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Order, error) {
			return existingOrder, nil
		},
		UpdateFunc: func(ctx context.Context, order *models.Order) error {
			updatedOrder = order
			return nil
		},
	}

	svc := services.NewOrderService(mockRepo, nil, 15)

	err := svc.HandleInventoryReserved(context.Background(), orderID)
	if err != nil {
		t.Fatalf("HandleInventoryReserved returned error: %v", err)
	}

	if updatedOrder == nil {
		t.Fatal("expected update to be called")
	}

	if !updatedOrder.InventoryConfirmed {
		t.Error("expected InventoryConfirmed to be true")
	}
}

func TestHandlePaymentFailed(t *testing.T) {
	orderID := uuid.New()
	existingOrder := &models.Order{
		ID:     orderID,
		Status: "PENDING",
	}

	var updatedOrder *models.Order
	mockRepo := &mockOrderRepository{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Order, error) {
			return existingOrder, nil
		},
		UpdateFunc: func(ctx context.Context, order *models.Order) error {
			updatedOrder = order
			return nil
		},
	}

	svc := services.NewOrderService(mockRepo, nil, 15)

	err := svc.HandlePaymentFailed(context.Background(), orderID, "declined")
	if err != nil {
		t.Fatalf("HandlePaymentFailed returned error: %v", err)
	}

	if updatedOrder.Status != "FAILED" {
		t.Errorf("expected status to be FAILED, got %s", updatedOrder.Status)
	}
}

func TestHandleInventoryFailed(t *testing.T) {
	orderID := uuid.New()
	existingOrder := &models.Order{
		ID:     orderID,
		Status: "PENDING",
	}

	var updatedOrder *models.Order
	mockRepo := &mockOrderRepository{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Order, error) {
			return existingOrder, nil
		},
		UpdateFunc: func(ctx context.Context, order *models.Order) error {
			updatedOrder = order
			return nil
		},
	}

	svc := services.NewOrderService(mockRepo, nil, 15)

	err := svc.HandleInventoryFailed(context.Background(), orderID, "out of stock")
	if err != nil {
		t.Fatalf("HandleInventoryFailed returned error: %v", err)
	}

	if updatedOrder.Status != "FAILED" {
		t.Errorf("expected status to be FAILED, got %s", updatedOrder.Status)
	}
}

func TestProcessTimeoutOrders(t *testing.T) {
	expiredOrder := models.Order{
		ID:     uuid.New(),
		Status: "PENDING",
	}

	var updatedOrder *models.Order
	mockRepo := &mockOrderRepository{
		GetExpiredPendingOrdersFunc: func(ctx context.Context, timeout time.Duration) ([]models.Order, error) {
			return []models.Order{expiredOrder}, nil
		},
		UpdateFunc: func(ctx context.Context, order *models.Order) error {
			updatedOrder = order
			return nil
		},
	}

	svc := services.NewOrderService(mockRepo, nil, 15)

	err := svc.ProcessTimeoutOrders(context.Background())
	if err != nil {
		t.Fatalf("ProcessTimeoutOrders returned error: %v", err)
	}

	if updatedOrder == nil {
		t.Fatal("expected update to be called for expired order")
	}

	if updatedOrder.Status != "CANCELLED" {
		t.Errorf("expected timeout order status to be CANCELLED, got %s", updatedOrder.Status)
	}
}
