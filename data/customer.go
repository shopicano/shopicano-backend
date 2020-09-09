package data

import (
	"git.cloudbro.net/michaelfigg/yallawebsites/models"
	"github.com/jinzhu/gorm"
)

type CustomerRepository interface {
	List(db *gorm.DB, storeID string, offset, limit int) ([]models.Customer, error)
	Search(db *gorm.DB, query, storeID string, offset, limit int) ([]models.Customer, error)
}
