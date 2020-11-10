package data

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/models"
)

type MarketplaceRepositoryImpl struct {
}

var marketplaceRepository MarketplaceRepository

func NewMarketplaceRepository() MarketplaceRepository {
	if marketplaceRepository == nil {
		marketplaceRepository = &MarketplaceRepositoryImpl{}
	}
	return marketplaceRepository
}

func (au *MarketplaceRepositoryImpl) CreateShippingMethod(db *gorm.DB, sm *models.ShippingMethod) error {
	if err := db.Table(sm.TableName()).Create(sm).Error; err != nil {
		return err
	}
	return nil
}

func (au *MarketplaceRepositoryImpl) UpdateShippingMethod(db *gorm.DB, sm *models.ShippingMethod) error {
	if err := db.Table(sm.TableName()).
		Where("id = ?", sm.ID).Save(sm).Error; err != nil {
		return err
	}
	return nil
}

func (au *MarketplaceRepositoryImpl) ListShippingMethods(db *gorm.DB, from, limit int) ([]models.ShippingMethod, error) {
	var data []models.ShippingMethod
	m := models.ShippingMethod{}
	if err := db.Table(m.TableName()).
		Offset(from).Limit(limit).
		Order("created_at DESC").Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (au *MarketplaceRepositoryImpl) ListActiveShippingMethods(db *gorm.DB, from, limit int) ([]models.ShippingMethod, error) {
	var data []models.ShippingMethod
	m := models.ShippingMethod{}
	if err := db.Table(m.TableName()).
		Where("is_published = ?", true).
		Offset(from).Limit(limit).
		Order("created_at DESC").Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (au *MarketplaceRepositoryImpl) DeleteShippingMethod(db *gorm.DB, ID string) error {
	m := models.ShippingMethod{}
	if err := db.Table(m.TableName()).
		Where("id = ?", ID).
		Delete(&m).Error; err != nil {
		return err
	}
	return nil
}

func (au *MarketplaceRepositoryImpl) GetShippingMethod(db *gorm.DB, ID string) (*models.ShippingMethod, error) {
	m := models.ShippingMethod{}
	if err := db.Table(m.TableName()).
		Where("id = ?", ID).
		First(&m).Error; err != nil {
		return &m, err
	}
	return &m, nil
}

func (au *MarketplaceRepositoryImpl) GetShippingMethodForUser(db *gorm.DB, ID string) (*models.ShippingMethod, error) {
	m := models.ShippingMethod{}
	if err := db.Table(m.TableName()).
		Where("id = ? AND is_published = ?", ID, true).
		First(&m).Error; err != nil {
		return &m, err
	}
	return &m, nil
}

func (au *MarketplaceRepositoryImpl) CreatePaymentMethod(db *gorm.DB, pm *models.PaymentMethod) error {
	if err := db.Table(pm.TableName()).Create(pm).Error; err != nil {
		return err
	}
	return nil
}

func (au *MarketplaceRepositoryImpl) UpdatePaymentMethod(db *gorm.DB, pm *models.PaymentMethod) error {
	if err := db.Table(pm.TableName()).
		Where("id = ?", pm.ID).Save(pm).Error; err != nil {
		return err
	}
	return nil
}

func (au *MarketplaceRepositoryImpl) ListPaymentMethods(db *gorm.DB, from, limit int) ([]models.PaymentMethod, error) {
	var data []models.PaymentMethod
	m := models.PaymentMethod{}
	if err := db.Table(m.TableName()).
		Offset(from).Limit(limit).
		Order("created_at DESC").Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (au *MarketplaceRepositoryImpl) ListActivePaymentMethods(db *gorm.DB, from, limit int) ([]models.PaymentMethod, error) {
	var data []models.PaymentMethod
	m := models.PaymentMethod{}
	if err := db.Table(m.TableName()).
		Where("is_published = ?", true).
		Offset(from).Limit(limit).
		Order("created_at DESC").Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (au *MarketplaceRepositoryImpl) DeletePaymentMethod(db *gorm.DB, ID string) error {
	m := models.PaymentMethod{}
	if err := db.Table(m.TableName()).
		Where("id = ?", ID).
		Delete(&m).Error; err != nil {
		return err
	}
	return nil
}

func (au *MarketplaceRepositoryImpl) GetPaymentMethod(db *gorm.DB, ID string) (*models.PaymentMethod, error) {
	m := models.PaymentMethod{}
	if err := db.Table(m.TableName()).
		Where("id = ?", ID).
		First(&m).Error; err != nil {
		return &m, err
	}
	return &m, nil
}

func (au *MarketplaceRepositoryImpl) GetPaymentMethodForUser(db *gorm.DB, ID string) (*models.PaymentMethod, error) {
	m := models.PaymentMethod{}
	if err := db.Table(m.TableName()).
		Where("id = ? AND is_published = ?", ID, true).
		First(&m).Error; err != nil {
		return &m, err
	}
	return &m, nil
}

func (au *MarketplaceRepositoryImpl) GetSettingsDetails(db *gorm.DB) (*models.SettingsDetails, error) {
	settings := models.Settings{}
	settingsDetails := models.SettingsDetails{}
	a := models.Address{}
	if err := db.Table(fmt.Sprintf("%s AS s", settings.TableName())).
		Where("s.id = ?", "1").
		Select("s.id AS id, s.name AS name, s.website AS website, s.status AS status, a.address AS address, a.city AS city, a.country AS country, a.postcode AS postcode, a.email AS email, a.phone AS phone, s.is_sign_up_enabled AS is_sign_up_enabled, s.enabled_auto_store_confirmation AS enabled_auto_store_confirmation, s.is_store_creation_enabled AS is_store_creation_enabled, s.default_commission_rate AS default_commission_rate, s.tag_line AS tag_line, s.created_at AS created_at, s.updated_at AS updated_at").
		Joins(fmt.Sprintf("LEFT JOIN %s AS a ON s.company_address_id = a.id", a.TableName())).
		Find(&settingsDetails).Error; err != nil {
		return nil, err
	}
	return &settingsDetails, nil
}

func (au *MarketplaceRepositoryImpl) GetSettings(db *gorm.DB) (*models.Settings, error) {
	settings := models.Settings{}
	if err := db.Table(settings.TableName()).
		Where("id = ?", "1").
		Find(&settings).Error; err != nil {
		return nil, err
	}
	return &settings, nil
}

func (au *MarketplaceRepositoryImpl) UpdateSettings(db *gorm.DB, s *models.Settings) error {
	if err := db.Table(s.TableName()).
		Select("name, status, website, company_address_id, default_commission_rate, enabled_auto_store_confirmation, tag_line, is_sign_up_enabled, is_store_creation_enabled, updated_at").
		Where("id = ?", "1").
		Update(map[string]interface{}{
			"name":                            s.Name,
			"status":                          s.Status,
			"website":                         s.Website,
			"company_address_id":              s.CompanyAddressID,
			"default_commission_rate":         s.DefaultCommissionRate,
			"enabled_auto_store_confirmation": s.EnabledAutoStoreConfirmation,
			"tag_line":                        s.TagLine,
			"is_sign_up_enabled":              s.IsSignUpEnabled,
			"is_store_creation_enabled":       s.IsStoreCreationEnabled,
			"updated_at":                      s.UpdatedAt,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (au *MarketplaceRepositoryImpl) ListPaymentMethodsByLocation(db *gorm.DB, locationID int64) ([]models.PaymentMethod, error) {
	var data []models.PaymentMethod
	m := models.PaymentMethod{}
	pol := models.PaymentForLocation{}
	if err := db.Table(fmt.Sprintf("%s AS pm", m.TableName())).
		Select("pm.id AS id, pm.name AS name, pm.processing_fee AS processing_fee, pm.min_processing_fee AS min_processing_fee, pm.max_processing_fee AS max_processing_fee, pm.is_published AS is_published, pm.is_offline_payment AS is_offline_payment, pm.is_flat AS is_flat, pm.created_at AS created_at, pm.updated_at AS updated_at").
		Joins(fmt.Sprintf("JOIN %s AS pol ON pm.id = pol.payment_method_id AND pol.location_id = %d", pol.TableName(), locationID)).
		Order("pm.created_at DESC").
		Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (au *MarketplaceRepositoryImpl) ListPaymentMethodsByLocationForUser(db *gorm.DB, locationID int64) ([]models.PaymentMethod, error) {
	var data []models.PaymentMethod
	m := models.PaymentMethod{}
	pol := models.PaymentForLocation{}
	loc := models.Location{}
	if err := db.Table(fmt.Sprintf("%s AS pm", m.TableName())).
		Select("pm.id AS id, pm.name AS name, pm.processing_fee AS processing_fee, pm.min_processing_fee AS min_processing_fee, pm.max_processing_fee AS max_processing_fee, pm.is_published AS is_published, pm.is_offline_payment AS is_offline_payment, pm.is_flat AS is_flat, pm.created_at AS created_at, pm.updated_at AS updated_at").
		Joins(fmt.Sprintf("JOIN %s AS pol ON pm.id = pol.payment_method_id AND pol.location_id = %d AND pm.is_published = %v", pol.TableName(), locationID, true)).
		Joins(fmt.Sprintf("JOIN %s AS loc ON loc.id = pol.location_id AND loc.is_published = %d", loc.TableName(), 1)).
		Order("pm.created_at DESC").
		Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (au *MarketplaceRepositoryImpl) ListShippingMethodsByLocation(db *gorm.DB, locationID int64) ([]models.ShippingMethod, error) {
	var data []models.ShippingMethod
	m := models.ShippingMethod{}
	sol := models.ShippingForLocation{}
	if err := db.Table(fmt.Sprintf("%s AS sm", m.TableName())).
		Select("sm.id AS id, sm.name AS name, sm.approximate_delivery_time AS approximate_delivery_time, sm.delivery_charge AS delivery_charge, sm.weight_unit AS weight_unit, sm.is_flat AS is_flat, sm.is_published AS is_published, sm.created_at AS created_at, sm.updated_at AS updated_at").
		Joins(fmt.Sprintf("JOIN %s AS sol ON sm.id = sol.shipping_method_id AND sol.location_id = %d", sol.TableName(), locationID)).
		Order("sm.created_at DESC").
		Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (au *MarketplaceRepositoryImpl) ListShippingMethodsByLocationForUser(db *gorm.DB, locationID int64) ([]models.ShippingMethod, error) {
	var data []models.ShippingMethod
	m := models.ShippingMethod{}
	sol := models.ShippingForLocation{}
	loc := models.Location{}
	if err := db.Table(fmt.Sprintf("%s AS sm", m.TableName())).
		Select("sm.id AS id, sm.name AS name, sm.approximate_delivery_time AS approximate_delivery_time, sm.delivery_charge AS delivery_charge, sm.weight_unit AS weight_unit, sm.is_flat AS is_flat, sm.is_published AS is_published, sm.created_at AS created_at, sm.updated_at AS updated_at").
		Joins(fmt.Sprintf("JOIN %s AS sol ON sm.id = sol.shipping_method_id AND sol.location_id = %d AND sm.is_published = %v", sol.TableName(), locationID, true)).
		Joins(fmt.Sprintf("JOIN %s AS loc ON loc.id = sol.location_id AND loc.is_published = %d", loc.TableName(), 1)).
		Order("sm.created_at DESC").
		Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}
