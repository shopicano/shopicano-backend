package data

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/models"
)

type CouponRepositoryImpl struct {
}

var couponRepository CouponRepository

func NewCouponRepository() CouponRepository {
	if couponRepository == nil {
		couponRepository = &CouponRepositoryImpl{}
	}
	return couponRepository
}

func (cr *CouponRepositoryImpl) Create(db *gorm.DB, c *models.Coupon) error {
	return db.Table(c.TableName()).Create(c).Error
}

func (cr *CouponRepositoryImpl) Update(db *gorm.DB, c *models.Coupon) error {
	return db.Table(c.TableName()).
		Where("store_id = ? AND id = ?", c.StoreID, c.ID).
		Update(map[string]interface{}{
			"code":             c.Code,
			"is_active":        c.IsActive,
			"discount_amount":  c.DiscountAmount,
			"is_flat_discount": c.IsFlatDiscount,
			"is_user_specific": c.IsUserSpecific,
			"max_discount":     c.MaxDiscount,
			"max_usage":        c.MaxUsage,
			"min_order_value":  c.MinOrderValue,
			"discount_type":    c.DiscountType,
			"start_at":         c.StartAt,
			"end_at":           c.EndAt,
			"updated_at":       c.UpdatedAt,
		}).Error
}

func (cr *CouponRepositoryImpl) List(db *gorm.DB, storeID string, from, limit int) ([]models.Coupon, error) {
	var coupons []models.Coupon
	c := models.Coupon{}
	if err := db.Table(c.TableName()).
		Where("store_id = ?", storeID).
		Order("updated_at DESC").
		Offset(from).
		Limit(limit).
		Find(&coupons).Error; err != nil {
		return nil, err
	}
	return coupons, nil
}

func (cr *CouponRepositoryImpl) Search(db *gorm.DB, storeID, query string, from, limit int) ([]models.Coupon, error) {
	var coupons []models.Coupon
	c := models.Coupon{}
	if err := db.Table(c.TableName()).
		Where("store_id = ? AND code LIKE ?", storeID, "%"+query+"%").
		Order("updated_at DESC").
		Offset(from).
		Limit(limit).
		Find(&coupons).Error; err != nil {
		return nil, err
	}
	return coupons, nil
}

func (cr *CouponRepositoryImpl) Delete(db *gorm.DB, storeID, couponID string) error {
	c := models.Coupon{}
	return db.Table(c.TableName()).
		Where("store_id = ? AND id = ?", storeID, couponID).
		Delete(&c).Error
}

func (cr *CouponRepositoryImpl) Get(db *gorm.DB, storeID, couponID string) (*models.Coupon, error) {
	c := models.Coupon{}
	if err := db.Table(c.TableName()).
		Where("store_id = ? AND id = ?", storeID, couponID).
		Find(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (cr *CouponRepositoryImpl) GetByCode(db *gorm.DB, storeID, couponCode string) (*models.Coupon, error) {
	c := models.Coupon{}
	if err := db.Table(c.TableName()).
		Where("store_id = ? AND code = ?", storeID, couponCode).
		Find(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (cr *CouponRepositoryImpl) AddUser(db *gorm.DB, cf *models.CouponFor) error {
	if err := db.Table(cf.TableName()).
		Create(cf).Error; err != nil {
		return err
	}
	return nil
}

func (cr *CouponRepositoryImpl) RemoveUser(db *gorm.DB, cf *models.CouponFor) error {
	if err := db.Table(cf.TableName()).
		Delete(cf, "coupon_id = ? AND user_id = ?", cf.CouponID, cf.UserID).Error; err != nil {
		return err
	}
	return nil
}

func (cr *CouponRepositoryImpl) ListUsers(db *gorm.DB, storeID, couponID string) ([]string, error) {
	var users []string

	c := models.Coupon{}
	cf := models.CouponFor{}
	rows, err := db.Table(fmt.Sprintf("%s AS c", c.TableName())).
		Select("cf.user_id AS users").
		Where("c.store_id = ? AND c.id = ?", storeID, couponID).
		Joins(fmt.Sprintf("JOIN %s AS cf ON c.id = cf.coupon_id", cf.TableName())).
		Group("cf.user_id").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}

		users = append(users, value)
	}

	return users, nil
}

func (cr *CouponRepositoryImpl) HasUser(db *gorm.DB, storeID, couponID, userID string) (bool, error) {
	var users []string

	c := models.Coupon{}
	cf := models.CouponFor{}
	if err := db.Table(fmt.Sprintf("%s AS c", c.TableName())).
		Select("cf.user_id").
		Joins(fmt.Sprintf("JOIN %s AS cf ON c.id = cf.coupon_id AND c.store_id = '%s' AND c.id = '%s' AND cf.user_id = '%s'",
			cf.TableName(), storeID, couponID, userID)).
		Group("cf.user_id").
		Scan(&users).Error; err != nil {
		if errors.IsRecordNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return len(users) > 0, nil
}

func (cr *CouponRepositoryImpl) AddUsage(db *gorm.DB, cu *models.CouponUsage) error {
	return db.Table(cu.TableName()).Create(cu).Error
}

func (cr *CouponRepositoryImpl) GetUsage(db *gorm.DB, couponID, userID string) (int, error) {
	cu := models.CouponUsage{}
	var count int
	if err := db.Table(cu.TableName()).
		Where("coupon_id = ? AND user_id = ?", couponID, userID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (cr *CouponRepositoryImpl) GetTotalUsage(db *gorm.DB, couponID string) (int, error) {
	cu := models.CouponUsage{}
	var count int
	if err := db.Table(cu.TableName()).
		Where("coupon_id = ?", couponID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
