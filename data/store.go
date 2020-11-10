package data

import (
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/models"
)

type StoreRepository interface {
	GetStoreUserProfile(db *gorm.DB, userID string) (*models.StaffProfile, error)
	CreateStore(db *gorm.DB, s *models.Store) error
	FindStoreByID(db *gorm.DB, ID string) (*models.Store, error)
	FindByID(db *gorm.DB, ID string) (*models.StoreView, error)
	AddStoreStuff(db *gorm.DB, staff *models.Staff) error
	ListStaffs(db *gorm.DB, storeID string, from, limit int) ([]models.StaffProfile, error)
	SearchStaffs(db *gorm.DB, storeID, query string, from, limit int) ([]models.StaffProfile, error)
	UpdateStoreStuffPermission(db *gorm.DB, staff *models.Staff) error
	DeleteStoreStuffPermission(db *gorm.DB, storeID, userID string) error
	IsAlreadyStaff(db *gorm.DB, userID string) (bool, error)
	List(db *gorm.DB, from, limit int) ([]models.Store, error)
	Search(db *gorm.DB, query string, from, limit int) ([]models.Store, error)
	UpdateStoreStatus(db *gorm.DB, s *models.Store) error
	UpdateStore(db *gorm.DB, s *models.Store) error

	GetStoreFinanceSummary(db *gorm.DB, storeID string) (*models.StoreFinanceSummaryView, error)
	GetStorePayoutSummary(db *gorm.DB, storeID string) (*models.StorePayoutSummaryView, error)
}
