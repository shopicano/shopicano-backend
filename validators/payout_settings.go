package validators

import (
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/errors"
)

type ReqCreateOrUpdatePayoutSettings struct {
	CountryID              int64  `json:"country_id" valid:"required"`
	AccountTypeID          string `json:"account_type_id" valid:"required"`
	BusinessName           string `json:"business_name" valid:"required"`
	BusinessAddressID      string `json:"business_address_id" valid:"required"`
	VatNumber              string `json:"vat_number"`
	PayoutMethodID         string `json:"payout_method_id" valid:"required"`
	PayoutMethodDetails    string `json:"payout_method_details" valid:"required"`
	PayoutMinimumThreshold int64  `json:"payout_minimum_threshold"`
}

func ValidateCreateOrUpdatePayoutSettings(ctx echo.Context) (*ReqCreateOrUpdatePayoutSettings, error) {
	pld := ReqCreateOrUpdatePayoutSettings{}
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
