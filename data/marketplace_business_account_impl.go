package data

import (
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/models"
)

func (au *MarketplaceRepositoryImpl) CreateBusinessAccountType(db *gorm.DB, m *models.BusinessAccountType) error {
	if err := db.Create(m).Error; err != nil {
		return err
	}
	return nil
}

func (au *MarketplaceRepositoryImpl) UpdateBusinessAccountType(db *gorm.DB, m *models.BusinessAccountType) error {
	if err := db.Table(m.TableName()).
		Where("id = ?", m.ID).
		Select("name, is_published, updated_at").
		Updates(map[string]interface{}{
			"name":         m.Name,
			"is_published": m.IsPublished,
			"updated_at":   m.UpdatedAt,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (au *MarketplaceRepositoryImpl) ListBusinessAccountTypes(db *gorm.DB, from, limit int) ([]models.BusinessAccountType, error) {
	bat := models.BusinessAccountType{}
	var data []models.BusinessAccountType
	if err := db.Table(bat.TableName()).
		Order("created_at DESC").
		Offset(from).
		Limit(limit).
		Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (au *MarketplaceRepositoryImpl) ListBusinessAccountTypesForUser(db *gorm.DB, from, limit int) ([]models.BusinessAccountType, error) {
	bat := models.BusinessAccountType{}
	var data []models.BusinessAccountType
	if err := db.Table(bat.TableName()).
		Order("created_at DESC").
		Offset(from).
		Limit(limit).
		Find(&data, "is_published = ?", true).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (au *MarketplaceRepositoryImpl) DeleteBusinessAccountType(db *gorm.DB, ID string) error {
	bat := models.BusinessAccountType{}
	if err := db.Table(bat.TableName()).Delete(&bat, "id = ?", ID).Error; err != nil {
		return err
	}
	return nil
}

func (au *MarketplaceRepositoryImpl) GetBusinessAccountType(db *gorm.DB, ID string) (*models.BusinessAccountType, error) {
	bat := models.BusinessAccountType{}
	if err := db.Table(bat.TableName()).Find(&bat, "id = ?", ID).Error; err != nil {
		return nil, err
	}
	return &bat, nil
}

func (au *MarketplaceRepositoryImpl) GetBusinessAccountTypeForUser(db *gorm.DB, ID string) (*models.BusinessAccountType, error) {
	bat := models.BusinessAccountType{}
	if err := db.Table(bat.TableName()).Find(&bat, "id = ? AND is_published = ?", ID, true).
		Error; err != nil {
		return nil, err
	}
	return &bat, nil
}
