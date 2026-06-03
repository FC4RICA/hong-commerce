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

// CreateOrder godoc
// @Summary      Create a new order
// @Description  Submit an order with multiple checkout items. Initiates checkout saga flow.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        order  body      CreateOrderRequest  true  "Order Creation Payload"
// @Success      201    {object}  models.Order
// @Failure      400    {object}  map[string]string "invalid request payload"
// @Failure      500    {object}  map[string]string "internal database error"
// @Router       / [post]
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

// GetAllOrders godoc
// @Summary      Get all orders
// @Description  Retrieve a list of all orders, including nested checkout items.
// @Tags         orders
// @Produce      json
// @Success      200      {array}   models.Order
// @Failure      500      {object}  map[string]string "failed to fetch orders"
// @Router       / [get]
func (h *OrderHandler) GetAllOrders(c *fiber.Ctx) error {
	orders, err := h.svc.GetAllOrders(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch orders",
		})
	}

	return c.Status(fiber.StatusOK).JSON(orders)
}

// GetOrderByID godoc
// @Summary      Get order details by ID
// @Description  Retrieve order status and items details by its Order UUID.
// @Tags         orders
// @Produce      json
// @Param        id       path      string  true  "Order ID (UUID)"
// @Success      200      {object}  models.Order
// @Failure      400      {object}  map[string]string "invalid order_id format"
// @Failure      404      {object}  map[string]string "order not found"
// @Router       /{id} [get]
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

// GetOrdersByUserID godoc
// @Summary      Get orders by user ID
// @Description  Retrieve a list of all orders associated with a specific User UUID.
// @Tags         orders
// @Produce      json
// @Param        userId   path      string  true  "User ID (UUID)"
// @Success      200      {array}   models.Order
// @Failure      400      {object}  map[string]string "invalid user_id format"
// @Failure      500      {object}  map[string]string "failed to fetch user orders"
// @Router       /user/{userId} [get]
func (h *OrderHandler) GetOrdersByUserID(c *fiber.Ctx) error {
	userIDStr := c.Params("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user_id format",
		})
	}

	orders, err := h.svc.GetOrdersByUserID(c.UserContext(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch user orders",
		})
	}

	return c.Status(fiber.StatusOK).JSON(orders)
} 
