package handlers

import (
	"github.com/FC4RICA/hong-commerce/order-service/internal/services"
	"github.com/gofiber/fiber/v2"
)

type OrderHandler struct {
	svc services.OrderService
}

func NewOrderHandler(svc services.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

type CreateOrderRequest struct {
	UserID     string  `json:"user_id" validate:"required"`
	ItemID     string  `json:"item_id" validate:"required"`
	Quantity   int     `json:"quantity" validate:"required,min=1"`
	TotalPrice float64 `json:"total_price" validate:"required,gt=0"`
}

func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Basic validation (could use go-playground/validator for more complex cases)
	if req.UserID == "" || req.ItemID == "" || req.Quantity <= 0 || req.TotalPrice <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing required fields or invalid values",
		})
	}

	order, err := h.svc.CreateOrder(c.Context(), req.UserID, req.ItemID, req.Quantity, req.TotalPrice)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(order)
}

func (h *OrderHandler) GetAllOrders(c *fiber.Ctx) error {
    orders, err := h.svc.GetAllOrders(c.Context())
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
        "error": "failed to fetch orders",
        })
    }
	
    return c.Status(fiber.StatusOK).JSON(orders)
} 
