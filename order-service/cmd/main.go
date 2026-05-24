package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/FC4RICA/hong-commerce/order-service/internal/config"
	"github.com/FC4RICA/hong-commerce/order-service/internal/handlers"
	"github.com/FC4RICA/hong-commerce/order-service/internal/models"
	"github.com/FC4RICA/hong-commerce/order-service/internal/repositories"
	"github.com/FC4RICA/hong-commerce/order-service/internal/services"
	"github.com/FC4RICA/hong-commerce/order-service/internal/workers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	// 1. Database Setup (with Retries)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)

	var db *gorm.DB
	var err error
	maxRetries := 5
	for i := 1; i <= maxRetries; i++ {
		log.Printf("Connecting to database (attempt %d/%d)...", i, maxRetries)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		if i == maxRetries {
			log.Fatalf("failed to connect database after %d attempts: %v", maxRetries, err)
		}
		time.Sleep(time.Duration(i) * 2 * time.Second)
	}

	// Auto Migrate
	if err := db.AutoMigrate(&models.Order{}, &models.OrderItem{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// 2. RabbitMQ Setup (with Retries)
	var mqConn *amqp.Connection
	for i := 1; i <= maxRetries; i++ {
		log.Printf("Connecting to RabbitMQ (attempt %d/%d)...", i, maxRetries)
		mqConn, err = amqp.Dial(cfg.RabbitMQURL)
		if err == nil {
			break
		}
		if i == maxRetries {
			log.Fatalf("failed to connect to RabbitMQ after %d attempts: %v", maxRetries, err)
		}
		time.Sleep(time.Duration(i) * 2 * time.Second)
	}
	defer mqConn.Close()

	mqChan, err := mqConn.Channel()
	if err != nil {
		log.Fatalf("failed to open a channel: %v", err)
	}
	defer mqChan.Close()

	// 3. Dependency Injection
	repo := repositories.NewOrderRepository(db)
	svc := services.NewOrderService(repo, mqChan, cfg.TimeoutMinutes)
	h := handlers.NewOrderHandler(svc)

	// 4. Start Background Workers
	eventWorker := workers.NewEventWorker(mqChan, svc)
	if err := eventWorker.Start(context.Background()); err != nil {
		log.Fatalf("failed to start event worker: %v", err)
	}

	housekeeping := workers.NewHousekeepingWorker(svc, 1*time.Minute)
	go housekeeping.Start(context.Background())

	// 5. Fiber App Setup
	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	})

	app.Use(logger.New())
	app.Use(recover.New())

	// Request Timeout Middleware (Custom safe implementation)
	orderTimeout := 5 * time.Second
	timeoutMiddleware := func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), orderTimeout)
		defer cancel()
		c.SetUserContext(ctx)
		return c.Next()
	}

	// Routes
	// Note: Gateway strips "/api/v1/orders", so we listen on "/"
	app.Post("/", timeoutMiddleware, h.CreateOrder)
	app.Get("/", timeoutMiddleware, h.GetAllOrders)
	app.Get("/:id", timeoutMiddleware, h.GetOrderByID)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	log.Printf("Order Service starting on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
