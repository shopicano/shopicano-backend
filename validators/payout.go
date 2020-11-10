package validators

import (
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/models"
)

type ReqCreatePayoutEntry struct {
	Amount int64  `json:"amount" valid:"range(1|1000000)"`
	Note   string `json:"note"`
}

func ValidateCreatePayoutEntry(ctx echo.Context) (*ReqCreatePayoutEntry, error) {
	pld := ReqCreatePayoutEntry{}
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

type ReqUpdatePayoutEntry struct {
	Amount        *int64                   `json:"amount"`
	Status        *models.PayoutSendStatus `json:"status"`
	Highlights    *string                  `json:"highlights"`
	FailureReason *string                  `json:"failure_reason"`
}

func ValidateUpdatePayoutEntry(ctx echo.Context) (*ReqUpdatePayoutEntry, error) {
	pld := ReqUpdatePayoutEntry{}
	if err := ctx.Bind(&pld); err != nil {
		return nil, err
	}

	ve := errors.ValidationError{}

	if pld.Status != nil && *pld.Status == models.PayoutSendStatusFailed && pld.FailureReason == nil {
		ve.Add("failure_reason", "is required")
	}
	if pld.Status != nil && !pld.Status.IsValid() {
		ve.Add("status", "is invalid")
	}

	if len(ve) > 0 {
		return nil, &ve
	}

	return &pld, nil
}
