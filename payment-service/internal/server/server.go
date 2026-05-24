package server

import (
	"log"

	"github.com/FC4RICA/hong-commerce/payment-service/config"
	"github.com/FC4RICA/hong-commerce/payment-service/db"
	"github.com/gofiber/fiber/v3"
)

type Server struct {
	app *fiber.App
	cfg *config.Config
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		app: fiber.New(),
		cfg: cfg,
	}
}

func (s *Server) Start() error {

	_, err := db.ConnectDB(s.cfg)
	if err != nil {
		log.Fatal(err)
	}

	return s.app.Listen(":" + s.cfg.Port)
}
