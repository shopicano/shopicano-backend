package validators

import (
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/models"
)

type ReqPaymentMethodCreate struct {
	Name             string `json:"name" valid:"required"`
	IsPublished      bool   `json:"is_published"`
	IsFlat           bool   `json:"is_flat"`
	ProcessingFee    int64  `json:"processing_fee"`
	MinProcessingFee int64  `json:"min_processing_fee"`
	MaxProcessingFee int64  `json:"max_processing_fee"`
	IsOfflinePayment bool   `json:"is_offline_payment"`
}

func ValidateCreatePaymentMethod(ctx echo.Context) (*ReqPaymentMethodCreate, error) {
	pld := ReqPaymentMethodCreate{}
	if err := ctx.Bind(&pld); err != nil {
		return nil, err
	}

	ok, err := govalidator.ValidateStruct(&pld)
	if ok {
		return &pld, nil
	}

	ve := errors.ValidationError{}

	for k, v := range govalidator.ErrorsByField(err) {
		ve.Add(k, v)
	}

	return nil, &ve
}

type ReqShippingMethodCreate struct {
	Name                    string            `json:"name" valid:"required"`
	ApproximateDeliveryTime int               `json:"approximate_delivery_time" valid:"required"`
	DeliveryCharge          int64             `json:"delivery_charge" valid:"required"`
	IsPublished             bool              `json:"is_published"`
	IsFlat                  bool              `json:"is_flat"`
	WeightUnit              models.WeightUnit `json:"weight_unit" valid:"required"`
}

func ValidateCreateShippingMethod(ctx echo.Context) (*ReqShippingMethodCreate, error) {
	pld := ReqShippingMethodCreate{}
	if err := ctx.Bind(&pld); err != nil {
		return nil, err
	}

	ve := errors.ValidationError{}

	ok, err := govalidator.ValidateStruct(&pld)
	if !ok {
		for k, v := range govalidator.ErrorsByField(err) {
			ve.Add(k, v)
		}
	}

	if pld.WeightUnit.IsValid() {
		ve.Add("weight_unit", "is invalid")
	}

	if len(ve) == 0 {
		return &pld, nil
	}

	return nil, &ve
}

type ReqSettingsUpdate struct {
	Name                         *string                `json:"name"`
	Website                      *string                `json:"website"`
	Status                       *models.PlatformStatus `json:"status"`
	EnabledAutoStoreConfirmation *bool                  `json:"enabled_auto_store_confirmation"`
	CompanyAddressID             *string                `json:"company_address_id"`
	IsSignUpEnabled              *bool                  `json:"is_sign_up_enabled"`
	IsStoreCreationEnabled       *bool                  `json:"is_store_creation_enabled"`
	DefaultCommissionRate        *int64                 `json:"default_commission_rate"`
	TagLine                      *string                `json:"tag_line"`
}

func ValidateUpdateSettings(ctx echo.Context) (*ReqSettingsUpdate, error) {
	pld := ReqSettingsUpdate{}
	if err := ctx.Bind(&pld); err != nil {
		return nil, err
	}

	ok, err := govalidator.ValidateStruct(&pld)
	if ok {
		return &pld, nil
	}

	ve := errors.ValidationError{}

	for k, v := range govalidator.ErrorsByField(err) {
		ve.Add(k, v)
	}

	return nil, &ve
}
