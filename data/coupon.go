package data

import (
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/models"
)

type CouponRepository interface {
	Create(db *gorm.DB, c *models.Coupon) error
	Update(db *gorm.DB, c *models.Coupon) error
	List(db *gorm.DB, storeID string, from, limit int) ([]models.Coupon, error)
	Search(db *gorm.DB, storeID, query string, from, limit int) ([]models.Coupon, error)
	Delete(db *gorm.DB, storeID, couponID string) error
	Get(db *gorm.DB, storeID, couponID string) (*models.Coupon, error)
	GetByCode(db *gorm.DB, storeID, couponCode string) (*models.Coupon, error)
	AddUser(db *gorm.DB, cf *models.CouponFor) error
	RemoveUser(db *gorm.DB, cf *models.CouponFor) error
	ListUsers(db *gorm.DB, storeID, couponID string) ([]string, error)
	HasUser(db *gorm.DB, storeID, couponID, userID string) (bool, error)
	AddUsage(db *gorm.DB, cu *models.CouponUsage) error
	GetUsage(db *gorm.DB, couponID, userID string) (int, error)
	GetTotalUsage(db *gorm.DB, couponID string) (int, error)
}
