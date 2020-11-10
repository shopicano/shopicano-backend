package validators

import (
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/errors"
	"strconv"
)

type reqUpdateLocation struct {
	IsPublished     *bool    `json:"is_published"`
	ShippingMethods []string `json:"shipping_methods"`
	PaymentMethods  []string `json:"payment_methods"`
}

func ValidateUpdateLocation(ctx echo.Context, single bool) (*reqUpdateLocation, error) {
	body := reqUpdateLocation{}

	if err := ctx.Bind(&body); err != nil {
		return nil, err
	}

	ve := errors.ValidationError{}

	if single {
		locationIDQ := ctx.Param("location_id")
		_, err := strconv.ParseInt(locationIDQ, 10, 64)
		if err != nil {
			ve.Add("location_id", "is invalid")
		}
	}

	if len(ve) == 0 {
		return &body, nil
	}
	return nil, &ve
}
