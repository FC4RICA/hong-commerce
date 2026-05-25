package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FC4RICA/hong-commerce/order-service/internal/handlers"
	"github.com/FC4RICA/hong-commerce/order-service/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type mockOrderService struct {
	CreateOrderFunc            func(ctx context.Context, userID uuid.UUID, items []models.OrderItem) (*models.Order, error)
	GetAllOrdersFunc           func(ctx context.Context) ([]models.Order, error)
	GetOrderByIDFunc           func(ctx context.Context, orderID uuid.UUID) (*models.Order, error)
	GetOrdersByUserIDFunc      func(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
	HandlePaymentSucceededFunc func(ctx context.Context, orderID uuid.UUID, paymentID uuid.UUID) error
	HandleInventoryReservedFunc func(ctx context.Context, orderID uuid.UUID) error
	HandlePaymentFailedFunc    func(ctx context.Context, orderID uuid.UUID, reason string) error
	HandleInventoryFailedFunc  func(ctx context.Context, orderID uuid.UUID, reason string) error
	ProcessTimeoutOrdersFunc   func(ctx context.Context) error
}

func (m *mockOrderService) CreateOrder(ctx context.Context, userID uuid.UUID, items []models.OrderItem) (*models.Order, error) {
	if m.CreateOrderFunc != nil {
		return m.CreateOrderFunc(ctx, userID, items)
	}
	return nil, nil
}

func (m *mockOrderService) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	if m.GetAllOrdersFunc != nil {
		return m.GetAllOrdersFunc(ctx)
	}
	return nil, nil
}

func (m *mockOrderService) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*models.Order, error) {
	if m.GetOrderByIDFunc != nil {
		return m.GetOrderByIDFunc(ctx, orderID)
	}
	return nil, nil
}

func (m *mockOrderService) GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	if m.GetOrdersByUserIDFunc != nil {
		return m.GetOrdersByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockOrderService) HandlePaymentSucceeded(ctx context.Context, orderID uuid.UUID, paymentID uuid.UUID) error {
	if m.HandlePaymentSucceededFunc != nil {
		return m.HandlePaymentSucceededFunc(ctx, orderID, paymentID)
	}
	return nil
}

func (m *mockOrderService) HandleInventoryReserved(ctx context.Context, orderID uuid.UUID) error {
	if m.HandleInventoryReservedFunc != nil {
		return m.HandleInventoryReservedFunc(ctx, orderID)
	}
	return nil
}

func (m *mockOrderService) HandlePaymentFailed(ctx context.Context, orderID uuid.UUID, reason string) error {
	if m.HandlePaymentFailedFunc != nil {
		return m.HandlePaymentFailedFunc(ctx, orderID, reason)
	}
	return nil
}

func (m *mockOrderService) HandleInventoryFailed(ctx context.Context, orderID uuid.UUID, reason string) error {
	if m.HandleInventoryFailedFunc != nil {
		return m.HandleInventoryFailedFunc(ctx, orderID, reason)
	}
	return nil
}

func (m *mockOrderService) ProcessTimeoutOrders(ctx context.Context) error {
	if m.ProcessTimeoutOrdersFunc != nil {
		return m.ProcessTimeoutOrdersFunc(ctx)
	}
	return nil
}

func TestCreateOrder_Handler(t *testing.T) {
	app := fiber.New()
	mockSvc := &mockOrderService{}
	handler := handlers.NewOrderHandler(mockSvc)
	app.Post("/", handler.CreateOrder)

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		productID := uuid.New()

		mockSvc.CreateOrderFunc = func(ctx context.Context, uID uuid.UUID, items []models.OrderItem) (*models.Order, error) {
			if uID != userID {
				return nil, errors.New("mismatched user id")
			}
			return &models.Order{
				ID:          uuid.New(),
				UserID:      uID,
				TotalAmount: 50.00,
				Status:      "PENDING",
			}, nil
		}

		reqBody, _ := json.Marshal(map[string]interface{}{
			"user_id": userID.String(),
			"items": []map[string]interface{}{
				{
					"product_id":   productID.String(),
					"product_name": "Keyboard",
					"quantity":     1,
					"unit_price":   50.00,
				},
			},
		})

		req := httptest.NewRequest("POST", "/", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed to run HTTP request: %v", err)
		}

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("expected 201 Created, got %d", resp.StatusCode)
		}
	})

	t.Run("Invalid UserID Format", func(t *testing.T) {
		reqBody := []byte(`{"user_id": "invalid-uuid", "items": []}`)
		req := httptest.NewRequest("POST", "/", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed to run HTTP request: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request, got %d", resp.StatusCode)
		}
	})

	t.Run("Invalid ProductID Format", func(t *testing.T) {
		userID := uuid.New()
		reqBody, _ := json.Marshal(map[string]interface{}{
			"user_id": userID.String(),
			"items": []map[string]interface{}{
				{
					"product_id":   "not-a-uuid",
					"product_name": "Keyboard",
					"quantity":     1,
					"unit_price":   50.00,
				},
			},
		})
		req := httptest.NewRequest("POST", "/", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed to run HTTP request: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request, got %d", resp.StatusCode)
		}
	})
}

func TestGetAllOrders_Handler(t *testing.T) {
	app := fiber.New()
	mockSvc := &mockOrderService{}
	handler := handlers.NewOrderHandler(mockSvc)
	app.Get("/", handler.GetAllOrders)

	t.Run("Success", func(t *testing.T) {
		mockSvc.GetAllOrdersFunc = func(ctx context.Context) ([]models.Order, error) {
			return []models.Order{
				{ID: uuid.New(), Status: "CONFIRMED"},
			}, nil
		}

		req := httptest.NewRequest("GET", "/", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed to run HTTP request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", resp.StatusCode)
		}
	})

	t.Run("Service Error", func(t *testing.T) {
		mockSvc.GetAllOrdersFunc = func(ctx context.Context) ([]models.Order, error) {
			return nil, errors.New("db error")
		}

		req := httptest.NewRequest("GET", "/", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed to run HTTP request: %v", err)
		}

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected 500 Internal Server Error, got %d", resp.StatusCode)
		}
	})
}

func TestGetOrderByID_Handler(t *testing.T) {
	app := fiber.New()
	mockSvc := &mockOrderService{}
	handler := handlers.NewOrderHandler(mockSvc)
	app.Get("/:id", handler.GetOrderByID)

	t.Run("Success", func(t *testing.T) {
		orderID := uuid.New()
		mockSvc.GetOrderByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Order, error) {
			if id != orderID {
				return nil, errors.New("not found")
			}
			return &models.Order{ID: orderID, Status: "PENDING"}, nil
		}

		req := httptest.NewRequest("GET", "/"+orderID.String(), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed to run HTTP request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", resp.StatusCode)
		}
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/not-a-uuid", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed to run HTTP request: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request, got %d", resp.StatusCode)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		orderID := uuid.New()
		mockSvc.GetOrderByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Order, error) {
			return nil, errors.New("gorm: record not found")
		}

		req := httptest.NewRequest("GET", "/"+orderID.String(), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed to run HTTP request: %v", err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404 Not Found, got %d", resp.StatusCode)
		}
	})
}

func TestGetOrdersByUserID_Handler(t *testing.T) {
	app := fiber.New()
	mockSvc := &mockOrderService{}
	handler := handlers.NewOrderHandler(mockSvc)
	app.Get("/user/:userId", handler.GetOrdersByUserID)

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		mockSvc.GetOrdersByUserIDFunc = func(ctx context.Context, id uuid.UUID) ([]models.Order, error) {
			if id != userID {
				return nil, errors.New("mismatched id")
			}
			return []models.Order{
				{ID: uuid.New(), UserID: userID, Status: "PENDING"},
			}, nil
		}

		req := httptest.NewRequest("GET", "/user/"+userID.String(), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed to run HTTP request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", resp.StatusCode)
		}
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/user/not-a-uuid", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed to run HTTP request: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request, got %d", resp.StatusCode)
		}
	})
}
