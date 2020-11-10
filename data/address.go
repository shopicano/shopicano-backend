package data

import (
	"git.cloudbro.net/michaelfigg/yallawebsites/models"
	"github.com/jinzhu/gorm"
)

type AddressRepository interface {
	CreateAddress(db *gorm.DB, a *models.Address) error
	UpdateAddress(db *gorm.DB, a *models.Address) error
	GetAddress(db *gorm.DB, userID, addressID string) (*models.AddressView, error)
	GetAddressByID(db *gorm.DB, addressID string) (*models.AddressView, error)
	GetRawAddressByID(db *gorm.DB, addressID string) (*models.Address, error)
	ListAddresses(db *gorm.DB, userID string, from, limit int) ([]models.AddressView, error)
	DeleteAddress(db *gorm.DB, userID, addressID string) error
}
