package data

import (
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/models"
)

type LocationRepositoryImpl struct {
}

var locationRepository LocationRepository

func NewLocationRepository() LocationRepository {
	if locationRepository == nil {
		locationRepository = &LocationRepositoryImpl{}
	}
	return locationRepository
}

func (l *LocationRepositoryImpl) List(db *gorm.DB, query string, args []interface{}) ([]models.Location, error) {
	loc := models.Location{}
	var locations []models.Location
	if err := db.Table(loc.TableName()).
		Where(query, args...).
		Order("id ASC").
		Find(&locations).Error; err != nil {
		return nil, err
	}
	return locations, nil
}

func (l *LocationRepositoryImpl) UpdateByID(db *gorm.DB, loc *models.Location) error {
	if err := db.Table(loc.TableName()).
		Select("is_published, shipping_method_id, payment_method_id").
		Where("id = ?", loc.ID).
		Update(map[string]interface{}{
			"is_published": loc.IsPublished,
		}).
		Error; err != nil {
		return err
	}
	return nil
}

func (l *LocationRepositoryImpl) FindByID(db *gorm.DB, locationID int) (*models.Location, error) {
	loc := models.Location{}
	if err := db.Table(loc.TableName()).
		Where("id = ?", locationID).
		Find(&loc).Error; err != nil {
		return nil, err
	}
	return &loc, nil
}

func (l *LocationRepositoryImpl) Find(db *gorm.DB) ([]models.Location, error) {
	loc := models.Location{}
	var locations []models.Location
	if err := db.Table(loc.TableName()).
		Order("id ASC").
		Find(&locations).Error; err != nil {
		return nil, err
	}
	return locations, nil
}

func (l *LocationRepositoryImpl) AddShippingMethod(db *gorm.DB, m *models.ShippingForLocation) error {
	if err := db.FirstOrCreate(m).
		Error; err != nil {
		return err
	}
	return nil
}

func (l *LocationRepositoryImpl) RemoveShippingMethod(db *gorm.DB, m *models.ShippingForLocation) error {
	if err := db.Table(m.TableName()).
		Where("shipping_method_id = ? AND location_id = ?", m.ShippingMethodID, m.LocationID).
		Delete(m).
		Error; err != nil {
		return err
	}
	return nil
}

func (l *LocationRepositoryImpl) AddPaymentMethod(db *gorm.DB, m *models.PaymentForLocation) error {
	if err := db.FirstOrCreate(m).
		Error; err != nil {
		return err
	}
	return nil
}

func (l *LocationRepositoryImpl) RemovePaymentMethod(db *gorm.DB, m *models.PaymentForLocation) error {
	if err := db.Table(m.TableName()).
		Where("payment_method_id = ? AND location_id = ?", m.PaymentMethodID, m.LocationID).
		Delete(m).
		Error; err != nil {
		return err
	}
	return nil
}
