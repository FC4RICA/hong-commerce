package server

import (
	"log"
	"os"

	"github.com/FC4RICA/hong-commerce/payment-service/config"
	"github.com/FC4RICA/hong-commerce/payment-service/internal/handlers"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

type Server struct {
	app     *fiber.App
	cfg     *config.Config
	logFile *os.File
}

func NewServer(cfg *config.Config, paymentHandler *handlers.PaymentHandler) *Server {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Printf("Warning: failed to create logs directory: %v", err)
	}

	file, err := os.OpenFile("logs/payment.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Warning: failed to open log file: %v", err)
	} else {
		// Example: log.SetOutput(file)
	}

	app := fiber.New()

	app.Use(logger.New())

	// Health check endpoint
	app.Get("/health", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "ok",
			"service": "payment-service",
		})
	})

	// Register routes (Gateway strips /api/v1/payments, so we listen at the root)
	app.Patch("/:paymentID/status", paymentHandler.UpdatePaymentStatus)

	return &Server{
		app:     app,
		cfg:     cfg,
		logFile: file,
	}
}

func (s *Server) Start() error {
	return s.app.Listen(":" + s.cfg.Port)
}

func (s *Server) Shutdown() error {
	if s.logFile != nil {
		_ = s.logFile.Close()
	}
	return s.app.Shutdown()
}
