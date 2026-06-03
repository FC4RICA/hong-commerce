package handlers

import (
	"errors"

	"github.com/FC4RICA/hong-commerce/payment-service/internal/services"
	"github.com/gofiber/fiber/v3"
)

type PaymentHandler struct {
	service services.PaymentService
}

func NewPaymentHandler(service services.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		service: service,
	}
}

type UpdateStatusRequest struct {
	Status         string  `json:"status"`
	TransactionRef string  `json:"transactionRef"`
	AmountPaid     float64 `json:"amountPaid"`
}

func (h *PaymentHandler) UpdatePaymentStatus(c fiber.Ctx) error {
	paymentID := c.Params("paymentID")

	var req UpdateStatusRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Status != "COMPLETED" && req.Status != "FAILED" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Status must be COMPLETED or FAILED",
		})
	}

	payment, idempotent, err := h.service.UpdateStatus(c.Context(), paymentID, req.Status, req.TransactionRef, req.AmountPaid)
	if err != nil {
		if errors.Is(err, services.ErrPaymentNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Payment not found"})
		}
		if errors.Is(err, services.ErrPaymentAlreadyProcessed) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Payment is already processed"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update payment"})
	}

	msg := "Payment status updated successfully"
	if idempotent {
		msg += " (idempotent)"
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":   msg,
		"paymentID": payment.ID.String(),
		"status":    payment.Status,
	})
}
