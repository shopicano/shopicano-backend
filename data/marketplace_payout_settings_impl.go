package data

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/models"
)

func (au *MarketplaceRepositoryImpl) CreatePayoutSettings(db *gorm.DB, m *models.PayoutSettings) error {
	if err := db.Create(m).Error; err != nil {
		return err
	}
	return nil
}

func (au *MarketplaceRepositoryImpl) UpdatePayoutSettings(db *gorm.DB, m *models.PayoutSettings) error {
	if err := db.Table(m.TableName()).
		Where("id = ?", m.ID).
		Select("country_id, account_type_id, business_name, business_address_id, vat_number, payout_method_id, payout_method_details, payout_minimum_threshold, updated_at").
		Updates(map[string]interface{}{
			"country_id":               m.CountryID,
			"account_type_id":          m.AccountTypeID,
			"business_name":            m.BusinessName,
			"business_address_id":      m.BusinessAddressID,
			"vat_number":               m.VatNumber,
			"payout_method_id":         m.PayoutMethodID,
			"payout_method_details":    m.PayoutMethodDetails,
			"payout_minimum_threshold": m.PayoutMinimumThreshold,
			"updated_at":               m.UpdatedAt,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (au *MarketplaceRepositoryImpl) GetPayoutSettings(db *gorm.DB, storeID string) (*models.PayoutSettings, error) {
	pos := models.PayoutSettings{}
	if err := db.Table(pos.TableName()).Find(&pos, "store_id = ?", storeID).Error; err != nil {
		return nil, err
	}
	return &pos, nil
}

func (au *MarketplaceRepositoryImpl) GetPayoutSettingsDetails(db *gorm.DB, storeID string) (*models.PayoutSettingsDetails, error) {
	psd := models.PayoutSettingsDetails{}
	if err := db.Table(fmt.Sprintf("%s AS ps", psd.TableName())).
		Select("ps.id AS id, ps.country_id AS country_id, l.name AS country_name, ps.business_name AS business_name, "+
			"ps.payout_minimum_threshold AS payout_minimum_threshold, ps.payout_method_details AS payout_method_details, "+
			"ps.vat_number AS vat_number, ps.payout_method_id AS payout_method_id, pom.name AS payout_method_name, "+
			"pom.inputs AS payout_method_inputs, ps.updated_at AS updated_at, ps.created_at AS created_at, "+
			"ps.store_id AS store_id, ps.business_address_id AS business_address_id, a.address AS business_address_address, "+
			"a.city AS business_address_city, a.state AS business_address_state, a.postcode AS business_address_postcode, "+
			"ps.account_type_id AS business_account_type_id, bat.name AS business_account_type_name").
		Joins("LEFT JOIN locations AS l ON ps.country_id = l.id").
		Joins("LEFT JOIN business_account_types AS bat ON ps.account_type_id = bat.id").
		Joins("LEFT JOIN addresses AS a ON ps.business_address_id = a.id").
		Joins("LEFT JOIN payout_methods AS pom ON ps.payout_method_id = pom.id").
		Find(&psd, "ps.store_id = ?", storeID).Error; err != nil {
		return nil, err
	}
	return &psd, nil
}
