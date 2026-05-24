package handlers

import (
	"fmt"

	"github.com/FC4RICA/hong-commerce/order-service/internal/models"
	"github.com/FC4RICA/hong-commerce/order-service/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type OrderHandler struct {
	svc services.OrderService
}

func NewOrderHandler(svc services.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

type OrderItemRequest struct {
	ProductID   string  `json:"product_id" validate:"required"`
	ProductName string  `json:"product_name" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
}

type CreateOrderRequest struct {
	UserID string             `json:"user_id" validate:"required"`
	Items  []OrderItemRequest `json:"items" validate:"required,min=1"`
}

func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user_id format (must be uuid)",
		})
	}

	// Convert request items to models
	var items []models.OrderItem
	for _, it := range req.Items {
		productUUID, err := uuid.Parse(it.ProductID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("invalid product_id format for item: %s", it.ProductID),
			})
		}
		items = append(items, models.OrderItem{
			ProductID:   productUUID,
			ProductName: it.ProductName,
			Quantity:    it.Quantity,
			UnitPrice:   it.UnitPrice,
		})
	}

	order, err := h.svc.CreateOrder(c.UserContext(), userUUID, items)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(order)
}

func (h *OrderHandler) GetAllOrders(c *fiber.Ctx) error {
	orders, err := h.svc.GetAllOrders(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch orders",
		})
	}

	return c.Status(fiber.StatusOK).JSON(orders)
}

func (h *OrderHandler) GetOrderByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	orderID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid order_id format",
		})
	}

	order, err := h.svc.GetOrderByID(c.UserContext(), orderID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "order not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(order)
} 
