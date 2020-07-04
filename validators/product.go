package validators

import (
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/errors"
)

type ReqProductCreate struct {
	Name             string   `json:"name" valid:"required,stringlength(3|100)"`
	Description      string   `json:"description" valid:"required,stringlength(3|100000)"`
	IsPublished      bool     `json:"is_published"`
	CategoryID       *string  `json:"category_id"`
	Image            string   `json:"image"`
	IsShippable      bool     `json:"is_shippable"`
	IsDigital        bool     `json:"is_digital"`
	SKU              string   `json:"sku" valid:"required,stringlength(1|100)"`
	Stock            int      `json:"stock" valid:"range(0|100000)"`
	Unit             string   `json:"unit" valid:"required,stringlength(1|20)"`
	Price            int64    `json:"price" valid:"range(0|10000000)"`
	MaxQuantityCount int      `json:"max_quantity_count"`
	ProductCost      int64    `json:"product_cost" valid:"range(0|10000000)"`
	AdditionalImages []string `json:"additional_images"`
}

func ValidateCreateProduct(ctx echo.Context) (*ReqProductCreate, error) {
	pld := ReqProductCreate{}

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

type ReqProductUpdate struct {
	Name                *string  `json:"name" valid:"required,stringlength(3|100)"`
	Description         *string  `json:"description" valid:"required,stringlength(3|100000)"`
	IsPublished         *bool    `json:"is_published"`
	CategoryID          *string  `json:"category_id"`
	Image               *string  `json:"image"`
	IsShippable         *bool    `json:"is_shippable"`
	IsDigital           *bool    `json:"is_digital"`
	SKU                 *string  `json:"sku" valid:"required,stringlength(1|100)"`
	Stock               *int     `json:"stock" valid:"range(0|100000)"`
	Unit                *string  `json:"unit" valid:"required,stringlength(1|20)"`
	Price               *int64   `json:"price" valid:"range(0|10000000)"`
	ProductCost         *int64   `json:"product_cost" valid:"range(0|10000000)"`
	MaxQuantityCount    *int     `json:"max_quantity_count" valid:"range(0,10000)"`
	DigitalDownloadLink *string  `json:"digital_download_link" valid:"stringlength(1|1000000)"`
	AdditionalImages    []string `json:"additional_images"`
}

func ValidateUpdateProduct(ctx echo.Context) (*ReqProductUpdate, error) {
	pld := ReqProductUpdate{}

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

type ReqAddProductAttribute struct {
	Key   string `json:"key" valid:"required"`
	Value string `json:"value" valid:"required"`
}

func ValidateAddProductAttribute(ctx echo.Context) (*ReqAddProductAttribute, error) {
	pld := ReqAddProductAttribute{}

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
