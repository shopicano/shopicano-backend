package api

import (
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/data"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/utils"
	"github.com/shopicano/shopicano-backend/validators"
	"net/http"
	"time"
)

func createReview(ctx echo.Context) error {
	orderID := ctx.Param("order_id")

	resp := core.Response{}

	pld, err := validators.ValidateCreateReview(ctx)
	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.ReviewDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()

	ou := data.NewOrderRepository()
	r, err := ou.GetDetailsAsUser(db, utils.GetUserID(ctx), orderID)
	if err != nil {
		resp.Title = "Order not found"
		resp.Status = http.StatusNotFound
		resp.Code = errors.OrderNotFound
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	rv := &models.Review{
		ID:          utils.NewUUID(),
		OrderID:     r.ID,
		Rating:      pld.Rating,
		Description: pld.Description,
		CreatedAt:   time.Now().UTC(),
	}

	err = ou.CreateReview(db, rv)
	if err != nil {
		if _, ok := errors.IsDuplicateKeyError(err); ok {
			resp.Title = "Feedback already given"
			resp.Status = http.StatusConflict
			resp.Code = errors.ReviewAlreadyExists
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusCreated
	resp.Data = rv
	return resp.ServerJSON(ctx)
}
