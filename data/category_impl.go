package data

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/helpers"
	"github.com/shopicano/shopicano-backend/models"
	"strings"
)

type CategoryRepositoryImpl struct {
}

var categoryRepository CategoryRepository

func NewCategoryRepository() CategoryRepository {
	if categoryRepository == nil {
		categoryRepository = &CategoryRepositoryImpl{}
	}

	return categoryRepository
}

func (cu *CategoryRepositoryImpl) Create(db *gorm.DB, c *models.Category) error {
	if err := db.Table(c.TableName()).Create(c).Error; err != nil {
		return err
	}
	return nil
}

func (cu *CategoryRepositoryImpl) List(db *gorm.DB, from, limit int) ([]models.ResCategorySearch, error) {
	var cols []models.ResCategorySearch
	col := models.Category{}
	if err := db.Table(fmt.Sprintf("%s AS c", col.TableName())).
		Select("COUNT(p.id) AS count, c.id, c.name, c.description, c.image, c.store_id").
		Joins("LEFT JOIN products AS p ON p.category_id = c.id").
		Group("c.name, c.description, c.image, c.id, c.updated_at, c.store_id").
		Where("c.is_published = ?", true).
		Offset(from).Limit(limit).
		Order("c.updated_at DESC").Find(&cols).Error; err != nil {
		return nil, err
	}
	return cols, nil
}

func (cu *CategoryRepositoryImpl) ListAsStoreStuff(db *gorm.DB, storeID string, from, limit int) ([]models.ResCategorySearchInternal, error) {
	var cols []models.ResCategorySearchInternal
	col := models.Category{}
	if err := db.Table(fmt.Sprintf("%s AS c", col.TableName())).
		Select("COUNT(p.id) AS count, c.id, c.name, c.description, c.image, c.store_id, c.created_at, c.is_published, c.updated_at").
		Joins("LEFT JOIN products AS p ON p.category_id = c.id").
		Group("c.name, c.description, c.image, c.id, c.updated_at, c.store_id, c.is_published, c.created_at").
		Where("c.store_id = ?", storeID).
		Offset(from).Limit(limit).
		Order("c.updated_at DESC").Find(&cols).Error; err != nil {
		return nil, err
	}
	return cols, nil
}

func (cu *CategoryRepositoryImpl) Search(db *gorm.DB, query string, from, limit int) ([]models.ResCategorySearch, error) {
	var cols []models.ResCategorySearch
	col := models.Category{}
	if err := db.Table(fmt.Sprintf("%s AS c", col.TableName())).
		Select("COUNT(p.id) AS count, c.id, c.name, c.description, c.image, c.store_id").
		Joins("LEFT JOIN products AS p ON p.category_id = c.id").
		Group("c.name, c.description, c.image, c.id, c.updated_at, c.store_id").
		Where("c.is_published = ? AND LOWER(c.name) LIKE ?", true, "%"+strings.ToLower(query)+"%").
		Offset(from).Limit(limit).
		Order("c.updated_at DESC").Find(&cols).Error; err != nil {
		return nil, err
	}
	return cols, nil
}

func (cu *CategoryRepositoryImpl) SearchAsStoreStuff(db *gorm.DB, storeID, query string, from, limit int) ([]models.ResCategorySearchInternal, error) {
	var cols []models.ResCategorySearchInternal
	col := models.Category{}
	if err := db.Table(fmt.Sprintf("%s AS c", col.TableName())).
		Select("COUNT(p.id) AS count, c.id, c.name, c.description, c.image, c.store_id, c.created_at, c.is_published, c.updated_at").
		Joins("LEFT JOIN products AS p ON p.category_id = c.id").
		Group("c.name, c.description, c.image, c.id, c.updated_at, c.store_id, c.is_published, c.created_at").
		Where("c.store_id = ? AND LOWER(c.name) LIKE ?", storeID, "%"+strings.ToLower(query)+"%").
		Offset(from).Limit(limit).
		Order("c.updated_at DESC").Find(&cols).Error; err != nil {
		return nil, err
	}
	return cols, nil
}

func (cu *CategoryRepositoryImpl) Delete(db *gorm.DB, storeID, categoryID string) error {
	col := models.Category{}
	if err := db.Table(col.TableName()).
		Where("store_id = ? AND id = ?", storeID, categoryID).
		Delete(&col).Error; err != nil {
		return err
	}
	return nil
}

func (cu *CategoryRepositoryImpl) Get(db *gorm.DB, categoryID string) (*models.Category, error) {
	col := models.Category{}
	if err := db.Table(col.TableName()).
		Where("id = ? AND is_published = ?", categoryID, true).
		First(&col).Error; err != nil {
		return nil, err
	}
	return &col, nil
}

func (cu *CategoryRepositoryImpl) GetAsStoreOwner(db *gorm.DB, storeID, categoryID string) (*models.Category, error) {
	col := models.Category{}
	if err := db.Table(col.TableName()).
		Where("store_id = ? AND id = ?", storeID, categoryID).
		First(&col).Error; err != nil {
		return nil, err
	}
	return &col, nil
}

func (cu *CategoryRepositoryImpl) Update(db *gorm.DB, c *models.Category) error {
	col := models.Category{}
	if err := db.Table(col.TableName()).
		Where("store_id = ? AND id = ?", c.StoreID, c.ID).
		Select("name", "description", "image", "is_published", "updated_at").
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

func (cu *CategoryRepositoryImpl) Stats(db *gorm.DB, from, limit int) ([]helpers.CategoryStats, error) {
	var stats []helpers.CategoryStats

	c := models.Category{}
	p := models.Product{}

	if err := db.Table(fmt.Sprintf("%s AS c", c.TableName())).
		Select("c.id AS id, c.name AS name, c.image AS image, c.description AS description, COALESCE(COUNT(p.id), 0) AS count").
		Joins(fmt.Sprintf("LEFT JOIN %s AS p ON c.id = p.category_id", p.TableName())).
		Group("c.id, c.name, c.image, c.description").
		Order("count DESC").
		Offset(from).
		Limit(limit).
		Find(&stats).Error; err != nil {
		return nil, err
	}

	if stats == nil {
		stats = []helpers.CategoryStats{}
	}

	return stats, nil
}

func (cu *CategoryRepositoryImpl) StatsAsStoreStuff(db *gorm.DB, storeID string, from, limit int) ([]helpers.CategoryStats, error) {
	var stats []helpers.CategoryStats

	p := models.Product{}
	oi := models.OrderedItem{}

	if err := db.Table(fmt.Sprintf("%s AS p", p.TableName())).
		Select("c.id AS id, c.name AS name, c.image AS image, c.description AS description, COALESCE(SUM(oi.quantity), 0) AS count").
		Joins(fmt.Sprintf("LEFT JOIN %s AS p ON c.id = p.category_id", p.TableName())).
		Joins(fmt.Sprintf("LEFT JOIN %s AS oi ON oi.product_id = p.id", oi.TableName())).
		Group("c.id, c.name, c.image, c.description").
		Order("count DESC").
		Offset(from).
		Limit(limit).
		Find(&stats, "p.store_id = ?", storeID).Error; err != nil {
		return nil, err
	}

	if stats == nil {
		stats = []helpers.CategoryStats{}
	}

	return stats, nil
}
