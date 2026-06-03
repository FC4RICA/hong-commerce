package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/FC4RICA/hong-commerce/catalog-service/config"
	"github.com/FC4RICA/hong-commerce/catalog-service/db"
	"github.com/FC4RICA/hong-commerce/catalog-service/internal/handlers"
	"github.com/FC4RICA/hong-commerce/catalog-service/internal/repositories"
	"github.com/FC4RICA/hong-commerce/catalog-service/internal/services"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func main() {
	// 1. Load configuration
	cfg := config.LoadConfig()

	// 2. Connect to database
	database, err := db.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. Initialize layers
	categoryRepo := repositories.NewCategoryRepository(database)
	productRepo := repositories.NewProductRepository(database)

	categoryService := services.NewCategoryService(categoryRepo)
	productService := services.NewProductService(productRepo, categoryRepo)

	categoryHandler := handlers.NewCategoryHandler(categoryService)
	productHandler := handlers.NewProductHandler(productService)

	// 4. Create Fiber application
	app := fiber.New()
	app.Use(logger.New())

	// Health Check
	app.Get("/health", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "ok",
			"service": "catalog-service",
		})
	})

	// 5. Register Routes (Gateway strips /api/v1/catalog)
	app.Get("/categories", categoryHandler.GetAllCategories)
	app.Get("/categories/:id", categoryHandler.GetCategoryByID)
	app.Post("/categories", categoryHandler.CreateCategory)
	app.Put("/categories/:id", categoryHandler.UpdateCategory)
	app.Delete("/categories/:id", categoryHandler.DeleteCategory)

	app.Get("/products", productHandler.GetAllProducts)
	app.Get("/products/:id", productHandler.GetProductByID)
	app.Post("/products", productHandler.CreateProduct)
	app.Put("/products/:id", productHandler.UpdateProduct)
	app.Delete("/products/:id", productHandler.DeleteProduct)

	// 6. Start server
	go func() {
		log.Printf("Starting catalog-service on port %s", cfg.Port)
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// 7. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down catalog-service gracefully...")
	if err := app.Shutdown(); err != nil {
		log.Printf("Error during graceful shutdown: %v", err)
	}
}
