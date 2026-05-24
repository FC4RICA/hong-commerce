package main

import (
	"log"

	"github.com/FC4RICA/hong-commerce/payment-service/config"
	"github.com/FC4RICA/hong-commerce/payment-service/internal/server"
)

func main() {
	cfg := config.LoadConfig()

	srv := server.NewServer(cfg)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
