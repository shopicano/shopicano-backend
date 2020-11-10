package validators

import (
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/utils"
)

type ReqCreateCoupon struct {
	Code            string            `json:"code" valid:"required,stringlength(1|100)"`
	IsActive        bool              `json:"is_active"`
	DiscountAmount  int64             `json:"discount_amount" valid:"required,range(1|100000000)"`
	IsFlatDiscount  bool              `json:"is_flat_discount"`
	IsUserSpecific  bool              `json:"is_user_specific"`
	MaxDiscount     int64             `json:"max_discount" valid:"range(0|100000000)"`
	MaxUsage        int               `json:"max_usage" valid:"required,range(1|100000000)"`
	MaxUsagePerUser int               `json:"max_usage_per_user" valid:"range(0|100000000)"`
	MinOrderValue   int64             `json:"min_order_value" valid:"range(0|100000000)"`
	DiscountType    models.CouponType `json:"discount_type" valid:"required"`
	StartAt         string            `json:"start_at" valid:"required"`
	EndAt           string            `json:"end_at" valid:"required"`
}

func ValidateCreateCoupon(ctx echo.Context) (*ReqCreateCoupon, error) {
	pld := ReqCreateCoupon{}

	if err := ctx.Bind(&pld); err != nil {
		return nil, err
	}

	ve := errors.ValidationError{}

	ok, err := govalidator.ValidateStruct(&pld)

	for k, v := range govalidator.ErrorsByField(err) {
		ve.Add(k, v)
	}

	st, err := utils.ParseDateTimeForInput(pld.StartAt)
	if err != nil {
		ve.Add("start_at", "is invalid")
	}

	et, err := utils.ParseDateTimeForInput(pld.EndAt)
	if err != nil {
		ve.Add("end_at", "is invalid")
	}

	if st.After(et) {
		ve.Add("start_at", "must be before end_at")
	}

	if ok && len(ve) == 0 {
		return &pld, nil
	}

	return nil, &ve
}

type ReqUpdateCoupon struct {
	Code            *string            `json:"code"`
	IsActive        *bool              `json:"is_active"`
	DiscountAmount  *int64             `json:"discount_amount"`
	IsFlatDiscount  *bool              `json:"is_flat_discount"`
	IsUserSpecific  *bool              `json:"is_user_specific"`
	MaxDiscount     *int64             `json:"max_discount"`
	MaxUsage        *int               `json:"max_usage"`
	MaxUsagePerUser *int               `json:"max_usage_per_user"`
	MinOrderValue   *int64             `json:"min_order_value"`
	DiscountType    *models.CouponType `json:"discount_type"`
	StartAt         *string            `json:"start_at"`
	EndAt           *string            `json:"end_at"`
}

func ValidateUpdateCoupon(ctx echo.Context) (*ReqUpdateCoupon, error) {
	pld := ReqUpdateCoupon{}

	if err := ctx.Bind(&pld); err != nil {
		return nil, err
	}

	ve := errors.ValidationError{}

	if pld.StartAt != nil {
		_, err := utils.ParseDateTimeForInput(*pld.StartAt)
		if err != nil {
			ve.Add("start_at", "is invalid")
		}
	}

	if pld.EndAt != nil {
		_, err := utils.ParseDateTimeForInput(*pld.EndAt)
		if err != nil {
			ve.Add("end_at", "is invalid")
		}
	}

	if len(ve) == 0 {
		return &pld, nil
	}
	return nil, &ve
}
