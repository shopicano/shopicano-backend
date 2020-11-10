package validators

import (
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/errors"
)

type ReqCreateBusinessAccountType struct {
	Name        string `json:"name" valid:"required,stringlength(1|100)"`
	IsPublished bool   `json:"is_published"`
}

func ValidateCreateBusinessAccountType(ctx echo.Context) (*ReqCreateBusinessAccountType, error) {
	pld := ReqCreateBusinessAccountType{}
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

type ReqUpdateBusinessAccountType struct {
	Name        *string `json:"name"`
	IsPublished *bool   `json:"is_published"`
}

func ValidateUpdateBusinessAccountType(ctx echo.Context) (*ReqUpdateBusinessAccountType, error) {
	pld := ReqUpdateBusinessAccountType{}
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
