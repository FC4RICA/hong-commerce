package main

import (
	"log"

	"github.com/FC4RICA/hong-commerce/order-service/internal/config"
	"github.com/FC4RICA/hong-commerce/order-service/internal/handlers"
	"github.com/FC4RICA/hong-commerce/order-service/internal/models"
	"github.com/FC4RICA/hong-commerce/order-service/internal/repositories"
	"github.com/FC4RICA/hong-commerce/order-service/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	// 1. Database Setup (GORM + Postgres)
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	
	// Auto Migrate
	if err := db.AutoMigrate(&models.Order{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// 2. Redis Setup
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	// 3. RabbitMQ Setup
	mqConn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer mqConn.Close()

	mqChan, err := mqConn.Channel()
	if err != nil {
		log.Fatalf("failed to open a channel: %v", err)
	}
	defer mqChan.Close()

	// Declare queue
	_, err = mqChan.QueueDeclare(
		"order.created", // name
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		log.Fatalf("failed to declare a queue: %v", err)
	}

	// 4. Dependency Injection
	repo := repositories.NewOrderRepository(db)
	svc := services.NewOrderService(repo, rdb, mqChan)
	h := handlers.NewOrderHandler(svc)

	// 5. Fiber App Setup
	app := fiber.New()
	app.Use(logger.New())
	app.Use(recover.New())

	// Routes
	// Note: Gateway strips "/api/v1/orders", so we listen on "/"
	app.Post("/", h.CreateOrder)
	app.Get("/", h.GetAllOrders)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	log.Printf("Order Service starting on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
