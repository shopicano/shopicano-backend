package data

import (
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/models"
)

type LocationRepository interface {
	List(db *gorm.DB, query string, args []interface{}) ([]models.Location, error)
	UpdateByID(db *gorm.DB, loc *models.Location) error
	FindByID(db *gorm.DB, locationID int) (*models.Location, error)
	Find(db *gorm.DB) ([]models.Location, error)
	AddShippingMethod(db *gorm.DB, m *models.ShippingForLocation) error
	RemoveShippingMethod(db *gorm.DB, m *models.ShippingForLocation) error
	AddPaymentMethod(db *gorm.DB, m *models.PaymentForLocation) error
	RemovePaymentMethod(db *gorm.DB, m *models.PaymentForLocation) error
}
