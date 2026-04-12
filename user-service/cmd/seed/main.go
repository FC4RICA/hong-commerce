package main

import (
	"context"
	"log"

	"github.com/FC4RICA/hong-commerce/user-service/internal"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := internal.LoadConfig()

	db, _ := pgxpool.New(context.Background(), cfg.DatabaseURL)
	defer db.Close()

	hash, _ := bcrypt.GenerateFromPassword([]byte(cfg.SeedAdminPassword), bcrypt.DefaultCost)

	_, err := db.Exec(context.Background(),
		`INSERT INTO users (email, password_hash, name, role)
         VALUES ($1, $2, $3, 'admin')
         ON CONFLICT (email) DO NOTHING`,
		cfg.SeedAdminEmail, string(hash), "Super Admin",
	)

	if err != nil {
		log.Fatalf("seed admin: %v", err)
	}

	log.Println("admin seeded")
}
