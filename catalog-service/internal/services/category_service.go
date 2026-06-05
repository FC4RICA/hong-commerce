package services

import (
	"errors"
	"fmt"

	"github.com/FC4RICA/hong-commerce/catalog-service/internal/entities"
	"github.com/FC4RICA/hong-commerce/catalog-service/internal/repositories"
	"gorm.io/gorm"
)

var (
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryAlreadyExists = errors.New("category already exists")
)

type CategoryService interface {
	CreateCategory(name string) (*entities.Category, error)
	GetAllCategories() ([]entities.Category, error)
	GetCategoryByID(id string) (*entities.Category, error)
	UpdateCategory(id string, name string) (*entities.Category, error)
	DeleteCategory(id string) error
}

type categoryServiceImpl struct {
	repo repositories.CategoryRepository
}

func NewCategoryService(repo repositories.CategoryRepository) CategoryService {
	return &categoryServiceImpl{repo: repo}
}

func (s *categoryServiceImpl) CreateCategory(name string) (*entities.Category, error) {
	if name == "" {
		return nil, errors.New("category name is required")
	}

	category := &entities.Category{
		Name: name,
	}

	if err := s.repo.Create(category); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrCategoryAlreadyExists
		}
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

func (s *categoryServiceImpl) GetAllCategories() ([]entities.Category, error) {
	categories, err := s.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve categories: %w", err)
	}
	return categories, nil
}

func (s *categoryServiceImpl) GetCategoryByID(id string) (*entities.Category, error) {
	category, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to retrieve category: %w", err)
	}
	return category, nil
}

func (s *categoryServiceImpl) UpdateCategory(id string, name string) (*entities.Category, error) {
	if name == "" {
		return nil, errors.New("category name is required")
	}

	category, err := s.GetCategoryByID(id)
	if err != nil {
		return nil, err
	}

	category.Name = name

	if err := s.repo.Update(category); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrCategoryAlreadyExists
		}
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return category, nil
}

func (s *categoryServiceImpl) DeleteCategory(id string) error {
	_, err := s.GetCategoryByID(id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}
