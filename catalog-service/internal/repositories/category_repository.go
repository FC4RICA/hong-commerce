package repositories

import (
	"github.com/FC4RICA/hong-commerce/catalog-service/internal/entities"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(category *entities.Category) error
	FindAll() ([]entities.Category, error)
	FindByID(id string) (*entities.Category, error)
	Update(category *entities.Category) error
	Delete(id string) error
}

type categoryRepositoryImpl struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepositoryImpl{db: db}
}

func (r *categoryRepositoryImpl) Create(category *entities.Category) error {
	return r.db.Create(category).Error
}

func (r *categoryRepositoryImpl) FindAll() ([]entities.Category, error) {
	var categories []entities.Category
	if err := r.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *categoryRepositoryImpl) FindByID(id string) (*entities.Category, error) {
	var category entities.Category
	if err := r.db.Where("id = ?", id).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepositoryImpl) Update(category *entities.Category) error {
	return r.db.Save(category).Error
}

func (r *categoryRepositoryImpl) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&entities.Category{}).Error
}
