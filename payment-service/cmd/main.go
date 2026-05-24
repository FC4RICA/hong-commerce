package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/FC4RICA/hong-commerce/payment-service/config"
	"github.com/FC4RICA/hong-commerce/payment-service/db"
	"github.com/FC4RICA/hong-commerce/payment-service/internal/entities"
	"github.com/FC4RICA/hong-commerce/payment-service/internal/handlers"
	"github.com/FC4RICA/hong-commerce/payment-service/internal/rabbitmq"
	"github.com/FC4RICA/hong-commerce/payment-service/internal/repositories"
	"github.com/FC4RICA/hong-commerce/payment-service/internal/server"
	"github.com/FC4RICA/hong-commerce/payment-service/internal/services"
)

func main() {
	cfg := config.LoadConfig()

	database, err := db.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// AutoMigrate the entities
	err = database.AutoMigrate(&entities.Payment{}, &entities.IdempotencyKey{})
	if err != nil {
		log.Fatalf("Failed to automigrate: %v", err)
	}

	rabbitURL := config.GetEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	rabbitClient, err := rabbitmq.NewClient(rabbitURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitClient.Close()

	publisher := rabbitmq.NewPublisher(rabbitClient)
	
	repo := repositories.NewPaymentRepository(database)
	svc := services.NewPaymentService(repo, publisher)

	paymentHandler := handlers.NewPaymentHandler(svc)
	consumer := rabbitmq.NewConsumer(rabbitClient, svc)
	if err := consumer.Start(); err != nil {
		log.Fatalf("Failed to start RabbitMQ consumer: %v", err)
	}

	srv := server.NewServer(cfg, paymentHandler)

	go func() {
		log.Printf("Starting payment-service on port %s", cfg.Port)
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")
	srv.Shutdown()
}
