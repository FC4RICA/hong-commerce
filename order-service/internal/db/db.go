package db

import (
	"fmt"
	"log"
	"os"

	"github.com/FC4RICA/hong-commerce/order-service/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDatabase() *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		getEnv("ORDER_DB_HOST", "localhost"),
		getEnv("ORDER_DB_USER", "postgres"),
		getEnv("ORDER_DB_PASSWORD", "postgres"),
		getEnv("ORDER_DB_NAME", "orderdb"),
		getEnv("ORDER_DB_PORT", "5432"),
		getEnv("ORDER_DB_SSLMODE", "disable"),
	)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// เรียกโมเดลที่แยกไว้มาสร้างตารางในฐานข้อมูล
	err = db.AutoMigrate(&models.Order{}, &models.OrderItem{})
	if err != nil {
		log.Fatalf("Failed to run database migration: %v", err)
	}

	log.Println("Database migration completed successfully!")
	return db
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}