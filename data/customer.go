package data

import (
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/models"
)

type CustomerRepository interface {
	List(db *gorm.DB, storeID string, offset, limit int) ([]models.Customer, error)
	Search(db *gorm.DB, query, storeID string, offset, limit int) ([]models.Customer, error)
}
