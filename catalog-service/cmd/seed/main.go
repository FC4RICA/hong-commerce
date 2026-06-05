package main

import (
	"log"
	"time"

	"github.com/FC4RICA/hong-commerce/catalog-service/config"
	"github.com/FC4RICA/hong-commerce/catalog-service/db"
	"github.com/FC4RICA/hong-commerce/catalog-service/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	// Connect to database (with Retries)
	var database *gorm.DB
	var err error
	maxRetries := 5
	for i := 1; i <= maxRetries; i++ {
		log.Printf("Connecting to database (attempt %d/%d)...", i, maxRetries)
		database, err = db.ConnectDB(cfg)
		if err == nil {
			break
		}
		if i == maxRetries {
			log.Fatalf("failed to connect database after %d attempts: %v", maxRetries, err)
		}
		time.Sleep(time.Duration(i) * 2 * time.Second)
	}

	seed(database)
}

func seed(db *gorm.DB) {
	log.Println("Seeding catalog data...")

	// 1. Create Categories
	categories := []entities.Category{
		{Name: "Gaming"},
		{Name: "Mobile Phones"},
		{Name: "Audio"},
		{Name: "Computers"},
	}

	for i := range categories {
		if err := db.Where("name = ?", categories[i].Name).FirstOrCreate(&categories[i]).Error; err != nil {
			log.Printf("Failed to seed category %s: %v", categories[i].Name, err)
		} else {
			log.Printf("Seeded category: %s", categories[i].Name)
		}
	}

	// 2. Create Products
	products := []entities.Product{
		{
			Name:        "Sony PlayStation 5",
			Description: "Next-gen gaming console with ultra-high speed SSD",
			Price:       16900.00,
			ImageURL:    "https://upload.wikimedia.org/wikipedia/commons/q/q2/PlayStation_5_Console.png",
			CategoryID:  getCategoryID(db, "Gaming"),
		},
		{
			Name:        "Xbox Series X",
			Description: "Power your dreams with the fastest, most powerful Xbox ever",
			Price:       15900.00,
			ImageURL:    "https://upload.wikimedia.org/wikipedia/commons/4/43/Xbox-Series-X_Isolated.png",
			CategoryID:  getCategoryID(db, "Gaming"),
		},
		{
			Name:        "iPhone 15 Pro",
			Description: "Forged in titanium and featuring the groundbreaking A17 Pro chip",
			Price:       41900.00,
			ImageURL:    "https://upload.wikimedia.org/wikipedia/commons/thumb/c/c1/IPhone_15_Pro_Max_Natural_Titanium_Vertical.png/800px-IPhone_15_Pro_Max_Natural_Titanium_Vertical.png",
			CategoryID:  getCategoryID(db, "Mobile Phones"),
		},
		{
			Name:        "Sony WH-1000XM5",
			Description: "Industry-leading noise canceling headphones",
			Price:       12900.00,
			ImageURL:    "https://upload.wikimedia.org/wikipedia/commons/e/ee/Sony_WH-1000XM5.png",
			CategoryID:  getCategoryID(db, "Audio"),
		},
	}

	for i := range products {
		if err := db.Where("name = ?", products[i].Name).FirstOrCreate(&products[i]).Error; err != nil {
			log.Printf("Failed to seed product %s: %v", products[i].Name, err)
		} else {
			log.Printf("Seeded product: %s", products[i].Name)
		}
	}

	log.Println("Catalog seeding completed!")
}

func getCategoryID(db *gorm.DB, name string) uuid.UUID {
	var cat entities.Category
	db.Where("name = ?", name).First(&cat)
	return cat.ID
}
