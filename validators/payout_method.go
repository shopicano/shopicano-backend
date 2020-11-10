package validators

import (
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/errors"
)

type ReqCreatePayoutMethod struct {
	Name        string `json:"name" valid:"required,stringlength(1|100)"`
	Inputs      string `json:"inputs"`
	IsPublished bool   `json:"is_published"`
}

func ValidateCreatePayoutMethod(ctx echo.Context) (*ReqCreatePayoutMethod, error) {
	pld := ReqCreatePayoutMethod{}
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

type ReqUpdatePayoutMethod struct {
	Name        *string `json:"name"`
	Inputs      *string `json:"inputs"`
	IsPublished *bool   `json:"is_published"`
}

func ValidateReqUpdatePayoutMethod(ctx echo.Context) (*ReqUpdatePayoutMethod, error) {
	pld := ReqUpdatePayoutMethod{}
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
