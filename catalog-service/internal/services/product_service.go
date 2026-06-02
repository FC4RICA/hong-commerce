package services

import (
	"errors"
	"fmt"

	"github.com/FC4RICA/hong-commerce/catalog-service/internal/entities"
	"github.com/FC4RICA/hong-commerce/catalog-service/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

type ProductService interface {
	CreateProduct(categoryID string, name, description string, price float64, imageURL string) (*entities.Product, error)
	GetAllProducts(categoryID string, search string) ([]entities.Product, error)
	GetProductByID(id string) (*entities.Product, error)
	UpdateProduct(id string, categoryID string, name, description string, price float64, imageURL string) (*entities.Product, error)
	DeleteProduct(id string) error
}

type productServiceImpl struct {
	productRepo  repositories.ProductRepository
	categoryRepo repositories.CategoryRepository
}

func NewProductService(productRepo repositories.ProductRepository, categoryRepo repositories.CategoryRepository) ProductService {
	return &productServiceImpl{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *productServiceImpl) CreateProduct(categoryID string, name, description string, price float64, imageURL string) (*entities.Product, error) {
	if name == "" {
		return nil, errors.New("product name is required")
	}
	if price < 0 {
		return nil, errors.New("product price cannot be negative")
	}

	catUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return nil, errors.New("invalid category_id format")
	}

	// Validate category exists
	_, err = s.categoryRepo.FindByID(categoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to validate category: %w", err)
	}

	product := &entities.Product{
		CategoryID:  catUUID,
		Name:        name,
		Description: description,
		Price:       price,
		ImageURL:    imageURL,
	}

	if err := s.productRepo.Create(product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

func (s *productServiceImpl) GetAllProducts(categoryID string, search string) ([]entities.Product, error) {
	products, err := s.productRepo.FindAll(categoryID, search)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve products: %w", err)
	}
	return products, nil
}

func (s *productServiceImpl) GetProductByID(id string) (*entities.Product, error) {
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to retrieve product: %w", err)
	}
	return product, nil
}

func (s *productServiceImpl) UpdateProduct(id string, categoryID string, name, description string, price float64, imageURL string) (*entities.Product, error) {
	product, err := s.GetProductByID(id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		product.Name = name
	}
	if description != "" {
		product.Description = description
	}
	if price >= 0 {
		product.Price = price
	}
	if imageURL != "" {
		product.ImageURL = imageURL
	}

	if categoryID != "" {
		catUUID, err := uuid.Parse(categoryID)
		if err != nil {
			return nil, errors.New("invalid category_id format")
		}

		// Validate category exists
		_, err = s.categoryRepo.FindByID(categoryID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrCategoryNotFound
			}
			return nil, fmt.Errorf("failed to validate category: %w", err)
		}

		product.CategoryID = catUUID
	}

	if err := s.productRepo.Update(product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

func (s *productServiceImpl) DeleteProduct(id string) error {
	_, err := s.GetProductByID(id)
	if err != nil {
		return err
	}

	if err := s.productRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}
