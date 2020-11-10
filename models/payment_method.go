package models

import (
	"github.com/shopicano/shopicano-backend/log"
	"time"
)

type PaymentMethod struct {
	ID               string    `json:"id" sql:"id" gorm:"primary_key"`
	Name             string    `json:"name" sql:"name" gorm:"unique;not null"`
	ProcessingFee    int64     `json:"processing_fee" gorm:"processing_fee"`
	MinProcessingFee int64     `json:"min_processing_fee" gorm:"min_processing_fee"`
	MaxProcessingFee int64     `json:"max_processing_fee" sql:"max_processing_fee"`
	IsPublished      bool      `json:"is_published" sql:"is_published" gorm:"index"`
	IsOfflinePayment bool      `json:"is_offline_payment" sql:"is_offline_payment" gorm:"is_offline_payment"`
	IsFlat           bool      `json:"is_flat" gorm:"column:is_flat"`
	CreatedAt        time.Time `json:"created_at" sql:"created_at" gorm:"not null;index"`
	UpdatedAt        time.Time `json:"updated_at" sql:"updated_at" gorm:"not null"`
}

func (pm *PaymentMethod) TableName() string {
	return "payment_methods"
}

func (pm *PaymentMethod) CalculateProcessingFee(bill int64) int64 {
	if pm.IsOfflinePayment || bill == 0 {
		return 0
	}

	log.Log().Info("Bill : ", bill)
	log.Log().Info("Pf : ", pm.ProcessingFee)

	if pm.IsFlat {
		fee := pm.ProcessingFee
		log.Log().Info("Fee Flat : ", fee)

		if fee > pm.MaxProcessingFee && pm.MaxProcessingFee != 0 {
			return pm.MaxProcessingFee
		} else if fee < pm.MinProcessingFee && pm.MinProcessingFee != 0 {
			return pm.MinProcessingFee
		} else {
			return fee
		}
	}

	fee := (bill * pm.ProcessingFee) / 100

	log.Log().Info("Fee : ", fee)

	if fee > pm.MaxProcessingFee && pm.MaxProcessingFee != 0 {
		return pm.MaxProcessingFee
	} else if fee < pm.MinProcessingFee && pm.MinProcessingFee != 0 {
		return pm.MinProcessingFee
	}
	return fee
}
