package models

import (
	"fmt"
	"github.com/shopicano/shopicano-backend/log"
	"time"
)

type CouponType string

const (
	ProductDiscount  CouponType = "product_discount"
	ShippingDiscount CouponType = "shipping_discount"
	TotalDiscount    CouponType = "total_discount"
)

type Coupon struct {
	ID              string     `json:"id" gorm:"column:id;primary_key"`
	StoreID         string     `json:"store_id" gorm:"column:store_id;index;not null"`
	Code            string     `json:"code" gorm:"column:code;unique_index"`
	IsActive        bool       `json:"is_active" gorm:"column:is_active;index"`
	DiscountAmount  int64      `json:"discount_amount" gorm:"column:discount_amount"`
	IsFlatDiscount  bool       `json:"is_flat_discount" gorm:"column:is_flat_discount"`
	IsUserSpecific  bool       `json:"is_user_specific" gorm:"column:is_user_specific"`
	MaxDiscount     int64      `json:"max_discount" gorm:"column:max_discount"`
	MaxUsage        int        `json:"max_usage" gorm:"column:max_usage"`
	MaxUsagePerUser int        `json:"max_usage_per_user" gorm:"column:max_usage_per_user;default:0"`
	MinOrderValue   int64      `json:"min_order_value" gorm:"column:min_order_value"`
	DiscountType    CouponType `json:"discount_type" gorm:"column:discount_type;index"`
	StartAt         time.Time  `json:"start_at" gorm:"column:start_at;index"`
	EndAt           time.Time  `json:"end_at" gorm:"column:end_at;index"`
	CreatedAt       time.Time  `json:"created_at" gorm:"column:created_at;index"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"column:updated_at;index"`
}

func (c *Coupon) TableName() string {
	return "coupons"
}

func (c *Coupon) ForeignKeys() []string {
	s := Store{}

	return []string{
		fmt.Sprintf("store_id;%s(id);RESTRICT;RESTRICT", s.TableName()),
	}
}

func (c *Coupon) IsValid() bool {
	now := time.Now().UTC()
	log.Log().Infoln(now)
	log.Log().Infoln(c.StartAt)
	log.Log().Infoln(c.EndAt)
	return c.IsActive && (now.After(c.StartAt) && now.Before(c.EndAt))
}

func (c *Coupon) CalculateDiscount(value int64) int64 {
	if value == 0 {
		return 0
	}
	if c.MinOrderValue != 0 && value < c.MinOrderValue {
		return 0
	}

	if c.IsFlatDiscount {
		if c.MaxDiscount != 0 && c.DiscountAmount > c.MaxDiscount {
			return c.MaxDiscount
		}
		return c.DiscountAmount
	}

	discount := (value * c.DiscountAmount) / 100
	if c.MaxDiscount != 0 && discount > c.MaxDiscount {
		return c.MaxDiscount
	}
	return discount
}
