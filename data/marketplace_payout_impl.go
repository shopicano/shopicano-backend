package data

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/models"
)

func (au *MarketplaceRepositoryImpl) CreatePayoutMethod(db *gorm.DB, m *models.PayoutMethod) error {
	if err := db.Create(m).Error; err != nil {
		return err
	}
	return nil
}

func (au *MarketplaceRepositoryImpl) UpdatePayoutMethod(db *gorm.DB, m *models.PayoutMethod) error {
	if err := db.Table(m.TableName()).
		Select("name, inputs, is_published, updated_at").
		Updates(map[string]interface{}{
			"name":         m.Name,
			"inputs":       m.Inputs,
			"is_published": m.IsPublished,
			"updated_at":   m.UpdatedAt,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (au *MarketplaceRepositoryImpl) ListPayoutMethods(db *gorm.DB, from, limit int) ([]models.PayoutMethod, error) {
	pom := models.PayoutMethod{}
	var data []models.PayoutMethod
	if err := db.Table(pom.TableName()).
		Order("created_at DESC").
		Offset(from).
		Limit(limit).
		Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (au *MarketplaceRepositoryImpl) ListPayoutMethodForUser(db *gorm.DB, from, limit int) ([]models.PayoutMethod, error) {
	pom := models.PayoutMethod{}
	var data []models.PayoutMethod
	if err := db.Table(pom.TableName()).
		Order("created_at DESC").
		Offset(from).
		Limit(limit).
		Find(&data, "is_published = ?", true).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (au *MarketplaceRepositoryImpl) DeletePayoutMethod(db *gorm.DB, ID string) error {
	pom := models.PayoutMethod{}
	if err := db.Table(pom.TableName()).
		Delete(&pom, "id = ?", ID).Error; err != nil {
		return err
	}
	return nil
}

func (au *MarketplaceRepositoryImpl) GetPayoutMethod(db *gorm.DB, ID string) (*models.PayoutMethod, error) {
	pom := models.PayoutMethod{}
	if err := db.Table(pom.TableName()).
		Find(&pom, "id = ?", ID).Error; err != nil {
		return nil, err
	}
	return &pom, nil
}

func (au *MarketplaceRepositoryImpl) GetPayoutMethodForUser(db *gorm.DB, ID string) (*models.PayoutMethod, error) {
	pom := models.PayoutMethod{}
	if err := db.Table(pom.TableName()).
		Find(&pom, "id = ? AND is_published = ?", ID, true).Error; err != nil {
		return nil, err
	}
	return &pom, nil
}

func (au *MarketplaceRepositoryImpl) CreatePayoutEntry(db *gorm.DB, m *models.PayoutSend) error {
	if err := db.Create(m).Error; err != nil {
		return err
	}
	return nil
}

func (au *MarketplaceRepositoryImpl) ListPayoutEntries(db *gorm.DB, storeID string, from, limit int) ([]models.PayoutSend, error) {
	m := models.PayoutSend{}
	var entries []models.PayoutSend
	if err := db.Table(m.TableName()).
		Where("store_id = ?", storeID).
		Order("created_at DESC").
		Offset(from).
		Limit(limit).
		Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func (au *MarketplaceRepositoryImpl) GetPayoutEntry(db *gorm.DB, storeID, entryID string) (*models.PayoutSend, error) {
	ps := models.PayoutSend{}
	if err := db.Table(ps.TableName()).Find(&ps, "id = ? AND store_id = ?", entryID, storeID).Error; err != nil {
		return nil, err
	}
	return &ps, nil
}

func (au *MarketplaceRepositoryImpl) GetPayoutEntryDetails(db *gorm.DB, storeID, entryID string) (*models.PayoutSendDetails, error) {
	pom := models.PayoutMethod{}
	ps := models.PayoutSendDetails{}
	if err := db.Table(fmt.Sprintf("%s AS ps", ps.TableName())).
		Select("ps.id AS id, ps.store_id AS store_id, ps.initiated_by_user_id AS initiated_by_user_id, ps.is_marketplace_initiated AS is_marketplace_initiated,"+
			"ps.status AS status, ps.amount AS amount, ps.failure_reason AS failure_reason, ps.note AS note, ps.highlights AS highlights,"+
			"pom.id AS payout_method_id, pom.name AS payout_method_name, pom.inputs AS payout_method_inputs, ps.payout_method_details AS payout_method_details,"+
			"ps.created_at AS created_at, ps.updated_at AS updated_at").
		Joins(fmt.Sprintf("LEFT JOIN %s AS pom ON ps.payout_method_id = pom.id", pom.TableName())).
		Find(&ps, "ps.id = ? AND ps.store_id = ?", entryID, storeID).
		Error; err != nil {
		return nil, err
	}
	return &ps, nil
}

func (au *MarketplaceRepositoryImpl) UpdatePayoutEntry(db *gorm.DB, ps *models.PayoutSend) error {
	if err := db.Table(ps.TableName()).
		Where("id = ? AND store_id = ?", ps.ID, ps.StoreID).
		Select("status, amount, failure_reason, highlights").
		Updates(map[string]interface{}{
			"status":         ps.Status,
			"amount":         ps.Amount,
			"failure_reason": ps.FailureReason,
			"highlights":     ps.Highlights,
		}).
		Error; err != nil {
		return err
	}
	return nil
}
