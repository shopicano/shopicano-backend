package validators

import (
	"git.cloudbro.net/michaelfigg/yallawebsites/errors"
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
)

type ReqBrainTreeCreatePayment struct {
	Nonce *string `json:"nonce" valid:"required,stringlength(1|1000)"`
}

func ValidateCreateReqBrainTreePayment(ctx echo.Context) (*ReqBrainTreeCreatePayment, error) {
	pld := ReqBrainTreeCreatePayment{}
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
