package validators

import (
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/errors"
)

type ReqReviewCreate struct {
	Rating      int    `json:"rating" valid:"required,range(1|5)"`
	Description string `json:"description" valid:"required,stringlength(2|100000)"`
}

func ValidateCreateReview(ctx echo.Context) (*ReqReviewCreate, error) {
	pld := ReqReviewCreate{}

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
