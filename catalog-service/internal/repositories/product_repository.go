package repositories

import (
	"github.com/FC4RICA/hong-commerce/catalog-service/internal/entities"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(product *entities.Product) error
	FindAll(categoryID string, search string) ([]entities.Product, error)
	FindByID(id string) (*entities.Product, error)
	Update(product *entities.Product) error
	Delete(id string) error
}

type productRepositoryImpl struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepositoryImpl{db: db}
}

func (r *productRepositoryImpl) Create(product *entities.Product) error {
	if err := r.db.Create(product).Error; err != nil {
        return err
    }
    return r.db.Preload("Category").First(product, "id = ?", product.ID).Error
}

func (r *productRepositoryImpl) FindAll(categoryID string, search string) ([]entities.Product, error) {
	var products []entities.Product
	query := r.db.Model(&entities.Product{}).Preload("Category")

	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	if err := query.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productRepositoryImpl) FindByID(id string) (*entities.Product, error) {
	var product entities.Product
	if err := r.db.Preload("Category").Where("id = ?", id).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepositoryImpl) Update(product *entities.Product) error {
	if err := r.db.Save(product).Error; err != nil {
        return err
    }
    return r.db.Preload("Category").First(product, "id = ?", product.ID).Error
}

func (r *productRepositoryImpl) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&entities.Product{}).Error
}
