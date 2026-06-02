package handlers

import (
	"errors"

	"github.com/FC4RICA/hong-commerce/catalog-service/internal/services"
	"github.com/gofiber/fiber/v3"
)

type CategoryHandler struct {
	service services.CategoryService
}

func NewCategoryHandler(service services.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

type CreateCategoryRequest struct {
	Name string `json:"name"`
}

type UpdateCategoryRequest struct {
	Name string `json:"name"`
}

func (h *CategoryHandler) CreateCategory(c fiber.Ctx) error {
	var req CreateCategoryRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	category, err := h.service.CreateCategory(req.Name)
	if err != nil {
		if errors.Is(err, services.ErrCategoryAlreadyExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(category)
}

func (h *CategoryHandler) GetAllCategories(c fiber.Ctx) error {
	categories, err := h.service.GetAllCategories()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(categories)
}

func (h *CategoryHandler) GetCategoryByID(c fiber.Ctx) error {
	id := c.Params("id")
	category, err := h.service.GetCategoryByID(id)
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

	return c.Status(fiber.StatusOK).JSON(category)
}

func (h *CategoryHandler) UpdateCategory(c fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateCategoryRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	category, err := h.service.UpdateCategory(id, req.Name)
	if err != nil {
		if errors.Is(err, services.ErrCategoryNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if errors.Is(err, services.ErrCategoryAlreadyExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(category)
}

func (h *CategoryHandler) DeleteCategory(c fiber.Ctx) error {
	id := c.Params("id")
	err := h.service.DeleteCategory(id)
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

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Category deleted successfully",
	})
}
