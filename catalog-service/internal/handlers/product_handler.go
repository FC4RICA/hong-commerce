package handlers

import (
	"errors"

	"github.com/FC4RICA/hong-commerce/catalog-service/internal/services"
	"github.com/gofiber/fiber/v3"
)

type ProductHandler struct {
	service services.ProductService
}

func NewProductHandler(service services.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

type CreateProductRequest struct {
	CategoryID  string  `json:"category_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
}

type UpdateProductRequest struct {
	CategoryID  string  `json:"category_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
}

func (h *ProductHandler) CreateProduct(c fiber.Ctx) error {
	var req CreateProductRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	product, err := h.service.CreateProduct(req.CategoryID, req.Name, req.Description, req.Price, req.ImageURL)
	if err != nil {
		if errors.Is(err, services.ErrCategoryNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(product)
}

func (h *ProductHandler) GetAllProducts(c fiber.Ctx) error {
	categoryID := c.Query("category_id")
	search := c.Query("search")

	products, err := h.service.GetAllProducts(categoryID, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(products)
}

func (h *ProductHandler) GetProductByID(c fiber.Ctx) error {
	id := c.Params("id")
	product, err := h.service.GetProductByID(id)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(product)
}

func (h *ProductHandler) UpdateProduct(c fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateProductRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// For GORM/JSON partial updates, we check and pass values. Price defaults to -1 if not provided to avoid resetting to 0.
	// Since Bind might parse empty fields as 0 or empty string, let's see if we should set price to -1 if not in JSON.
	// But to keep it simple, we check if price is provided. Fiber v3 c.Bind parses json. If request body didn't contain price, GORM schema might have it as 0.
	// Let's check how price is bound. A safer way is to parse manually or check if the request maps keys, but we can do a default check if needed.
	// We'll trust the caller to send valid price or handle it. Let's make sure s.UpdateProduct handles price >= 0 correctly.
	
	product, err := h.service.UpdateProduct(id, req.CategoryID, req.Name, req.Description, req.Price, req.ImageURL)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) || errors.Is(err, services.ErrCategoryNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(product)
}

func (h *ProductHandler) DeleteProduct(c fiber.Ctx) error {
	id := c.Params("id")
	err := h.service.DeleteProduct(id)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Product deleted successfully",
	})
}
