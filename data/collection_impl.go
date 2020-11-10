package data

import (
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/models"
	"strings"
)

type CollectionRepositoryImpl struct {
}

var collectionRepository CollectionRepository

func NewCollectionRepository() CollectionRepository {
	if collectionRepository == nil {
		collectionRepository = &CollectionRepositoryImpl{}
	}

	return collectionRepository
}

func (cu *CollectionRepositoryImpl) Create(db *gorm.DB, c *models.Collection) error {
	if err := db.Table(c.TableName()).Create(c).Error; err != nil {
		return err
	}
	return nil
}

func (cu *CollectionRepositoryImpl) List(db *gorm.DB, from, limit int) ([]models.Collection, error) {
	var cols []models.Collection
	col := models.Collection{}
	if err := db.Table(col.TableName()).
		Select("DISTINCT(name), id, description, image, is_published, created_at, updated_at").
		Where("is_published = ?", true).
		Offset(from - limit).Limit(limit).
		Order("updated_at DESC").Find(&cols).Error; err != nil {
		return nil, err
	}
	return cols, nil
}

func (cu *CollectionRepositoryImpl) ListAsStoreStuff(db *gorm.DB, storeID string, from, limit int) ([]models.Collection, error) {
	var cols []models.Collection
	col := models.Collection{}
	if err := db.Table(col.TableName()).
		Where("store_id = ?", storeID).
		Offset(from - limit).Limit(limit).
		Order("updated_at DESC").Find(&cols).Error; err != nil {
		return nil, err
	}
	return cols, nil
}

func (cu *CollectionRepositoryImpl) Search(db *gorm.DB, query string, from, limit int) ([]models.Collection, error) {
	var cols []models.Collection
	col := models.Collection{}
	if err := db.Table(col.TableName()).
		Select("DISTINCT(name), id, description, image, is_published, created_at, updated_at").
		Where("is_published = ? AND LOWER(name) LIKE ?", true, "%"+strings.ToLower(query)+"%").
		Offset(from - limit).Limit(limit).
		Order("updated_at DESC").Find(&cols).Error; err != nil {
		return nil, err
	}
	return cols, nil
}

func (cu *CollectionRepositoryImpl) SearchAsStoreStuff(db *gorm.DB, storeID, query string, from, limit int) ([]models.Collection, error) {
	var cols []models.Collection
	col := models.Collection{}
	if err := db.Table(col.TableName()).
		Where("store_id = ? AND LOWER(name) LIKE ?", storeID, "%"+strings.ToLower(query)+"%").
		Offset(from - limit).Limit(limit).
		Order("updated_at DESC").Find(&cols).Error; err != nil {
		return nil, err
	}
	return cols, nil
}

func (cu *CollectionRepositoryImpl) Delete(db *gorm.DB, storeID, collectionID string) error {
	col := models.Collection{}
	if err := db.Table(col.TableName()).
		Where("store_id = ? AND id = ?", storeID, collectionID).
		Delete(&col).Error; err != nil {
		return err
	}
	return nil
}

func (cu *CollectionRepositoryImpl) Get(db *gorm.DB, collectionID string) (*models.Collection, error) {
	col := models.Collection{}
	if err := db.Table(col.TableName()).
		Where("id = ? AND is_published = ?", collectionID, true).
		Find(&col).Error; err != nil {
		return nil, err
	}
	return &col, nil
}

func (cu *CollectionRepositoryImpl) GetAsStoreOwner(db *gorm.DB, storeID, collectionID string) (*models.Collection, error) {
	col := models.Collection{}
	if err := db.Table(col.TableName()).
		Where("store_id = ? AND id = ?", storeID, collectionID).
		First(&col).Error; err != nil {
		return nil, err
	}
	return &col, nil
}

func (cu *CollectionRepositoryImpl) Update(db *gorm.DB, c *models.Collection) error {
	if err := db.Table(c.TableName()).
		Where("store_id = ? AND id = ?", c.StoreID, c.ID).
		Select("name", "description", "is_published", "image", "updated_at").
		Updates(map[string]interface{}{
			"name":         c.Name,
			"description":  c.Description,
			"is_published": c.IsPublished,
			"image":        c.Image,
			"updated_at":   c.UpdatedAt,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (cu *CollectionRepositoryImpl) AddProducts(db *gorm.DB, cop *models.CollectionOfProduct) error {
	if err := db.Table(cop.TableName()).Create(cop).Error; err != nil {
		return err
	}
	return nil
}

func (cu *CollectionRepositoryImpl) RemoveProducts(db *gorm.DB, cop *models.CollectionOfProduct) error {
	if err := db.Table(cop.TableName()).
		Delete(cop, "collection_id = ? AND product_id = ?", cop.CollectionID, cop.ProductID).Error; err != nil {
		return err
	}
	return nil
}
