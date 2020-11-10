package data

import (
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/models"
)

type MarketplaceRepository interface {
	CreateShippingMethod(db *gorm.DB, sm *models.ShippingMethod) error
	UpdateShippingMethod(db *gorm.DB, sm *models.ShippingMethod) error
	ListShippingMethods(db *gorm.DB, from, limit int) ([]models.ShippingMethod, error)
	ListActiveShippingMethods(db *gorm.DB, from, limit int) ([]models.ShippingMethod, error)
	ListShippingMethodsByLocation(db *gorm.DB, locationID int64) ([]models.ShippingMethod, error)
	ListShippingMethodsByLocationForUser(db *gorm.DB, locationID int64) ([]models.ShippingMethod, error)
	DeleteShippingMethod(db *gorm.DB, ID string) error
	GetShippingMethod(db *gorm.DB, ID string) (*models.ShippingMethod, error)
	GetShippingMethodForUser(db *gorm.DB, ID string) (*models.ShippingMethod, error)

	CreatePaymentMethod(db *gorm.DB, pm *models.PaymentMethod) error
	UpdatePaymentMethod(db *gorm.DB, pm *models.PaymentMethod) error
	ListPaymentMethods(db *gorm.DB, from, limit int) ([]models.PaymentMethod, error)
	ListActivePaymentMethods(db *gorm.DB, from, limit int) ([]models.PaymentMethod, error)
	ListPaymentMethodsByLocation(db *gorm.DB, locationID int64) ([]models.PaymentMethod, error)
	ListPaymentMethodsByLocationForUser(db *gorm.DB, locationID int64) ([]models.PaymentMethod, error)
	DeletePaymentMethod(db *gorm.DB, ID string) error
	GetPaymentMethod(db *gorm.DB, ID string) (*models.PaymentMethod, error)
	GetPaymentMethodForUser(db *gorm.DB, ID string) (*models.PaymentMethod, error)

	CreateBusinessAccountType(db *gorm.DB, m *models.BusinessAccountType) error
	UpdateBusinessAccountType(db *gorm.DB, m *models.BusinessAccountType) error
	ListBusinessAccountTypes(db *gorm.DB, from, limit int) ([]models.BusinessAccountType, error)
	ListBusinessAccountTypesForUser(db *gorm.DB, from, limit int) ([]models.BusinessAccountType, error)
	DeleteBusinessAccountType(db *gorm.DB, ID string) error
	GetBusinessAccountType(db *gorm.DB, ID string) (*models.BusinessAccountType, error)
	GetBusinessAccountTypeForUser(db *gorm.DB, ID string) (*models.BusinessAccountType, error)

	CreatePayoutMethod(db *gorm.DB, m *models.PayoutMethod) error
	UpdatePayoutMethod(db *gorm.DB, m *models.PayoutMethod) error
	ListPayoutMethods(db *gorm.DB, from, limit int) ([]models.PayoutMethod, error)
	ListPayoutMethodForUser(db *gorm.DB, from, limit int) ([]models.PayoutMethod, error)
	DeletePayoutMethod(db *gorm.DB, ID string) error
	GetPayoutMethod(db *gorm.DB, ID string) (*models.PayoutMethod, error)
	GetPayoutMethodForUser(db *gorm.DB, ID string) (*models.PayoutMethod, error)

	CreatePayoutSettings(db *gorm.DB, m *models.PayoutSettings) error
	UpdatePayoutSettings(db *gorm.DB, m *models.PayoutSettings) error
	GetPayoutSettings(db *gorm.DB, storeID string) (*models.PayoutSettings, error)
	GetPayoutSettingsDetails(db *gorm.DB, storeID string) (*models.PayoutSettingsDetails, error)

	CreatePayoutEntry(db *gorm.DB, m *models.PayoutSend) error
	ListPayoutEntries(db *gorm.DB, storeID string, from, limit int) ([]models.PayoutSend, error)
	GetPayoutEntry(db *gorm.DB, storeID, entryID string) (*models.PayoutSend, error)
	GetPayoutEntryDetails(db *gorm.DB, storeID, entryID string) (*models.PayoutSendDetails, error)
	UpdatePayoutEntry(db *gorm.DB, ps *models.PayoutSend) error

	GetSettings(db *gorm.DB) (*models.Settings, error)
	GetSettingsDetails(db *gorm.DB) (*models.SettingsDetails, error)
	UpdateSettings(db *gorm.DB, s *models.Settings) error
}
