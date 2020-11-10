package api

import (
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/data"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/middlewares"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/utils"
	"github.com/shopicano/shopicano-backend/validators"
	"net/http"
	"strconv"
	"time"
)

func RegisterCouponRoutes(publicEndpoints, platformEndpoints *echo.Group) {
	couponsPublicPath := publicEndpoints.Group("/coupons")
	couponsPlatformPath := platformEndpoints.Group("/coupons")

	func(g echo.Group) {
		g.Use(middlewares.HasStore())
		g.Use(middlewares.IsStoreActive())
		g.Use(middlewares.IsStoreManager())
		g.POST("/", createCoupon)
		g.PATCH("/:coupon_id/", updateCoupon)
		g.DELETE("/:coupon_id/", deleteCoupon)
		g.GET("/:coupon_id/", getCoupon)
		g.GET("/", listCoupons)
		g.PATCH("/:coupon_id/users/", addCouponUsers)
		g.DELETE("/:coupon_id/users/", removeCouponUsers)
	}(*couponsPlatformPath)

	func(g echo.Group) {
		g.Use(middlewares.JWTAuth())
		g.GET("/:coupon_code/check/", checkCouponAvailability)
	}(*couponsPublicPath)
}

func createCoupon(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)

	req, err := validators.ValidateCreateCoupon(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.ProductCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()
	cr := data.NewCouponRepository()

	st, _ := utils.ParseDateTimeForInput(req.StartAt)
	et, _ := utils.ParseDateTimeForInput(req.EndAt)

	c := models.Coupon{
		ID:              utils.NewUUID(),
		StoreID:         storeID,
		IsUserSpecific:  req.IsUserSpecific,
		DiscountType:    req.DiscountType,
		Code:            req.Code,
		DiscountAmount:  req.DiscountAmount,
		StartAt:         st.UTC(),
		EndAt:           et.UTC(),
		IsActive:        req.IsActive,
		IsFlatDiscount:  req.IsFlatDiscount,
		MaxDiscount:     req.MaxDiscount,
		MaxUsage:        req.MaxUsage,
		MaxUsagePerUser: req.MaxUsagePerUser,
		MinOrderValue:   req.MinOrderValue,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	err = cr.Create(db, &c)
	if err != nil {
		msg, ok := errors.IsDuplicateKeyError(err)
		if ok {
			resp.Title = msg
			resp.Status = http.StatusConflict
			resp.Code = errors.CouponAlreadyExists
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
	resp.Title = "Coupon created"
	resp.Data = c
	return resp.ServerJSON(ctx)
}

func updateCoupon(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)
	couponID := ctx.Param("coupon_id")

	req, err := validators.ValidateUpdateCoupon(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.ProductCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()
	cr := data.NewCouponRepository()

	c, err := cr.Get(db, storeID, couponID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Coupon not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CouponNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if req.Code != nil {
		c.Code = *req.Code
	}
	if req.IsUserSpecific != nil {
		c.IsUserSpecific = *req.IsUserSpecific
	}
	if req.DiscountType != nil {
		c.DiscountType = *req.DiscountType
	}
	if req.DiscountAmount != nil {
		c.DiscountAmount = *req.DiscountAmount
	}
	if req.IsActive != nil {
		c.IsActive = *req.IsActive
	}
	if req.IsFlatDiscount != nil {
		c.IsFlatDiscount = *req.IsFlatDiscount
	}
	if req.MaxDiscount != nil {
		c.MaxDiscount = *req.MaxDiscount
	}
	if req.MaxUsage != nil {
		c.MaxUsage = *req.MaxUsage
	}
	if req.MaxUsagePerUser != nil {
		c.MaxUsagePerUser = *req.MaxUsagePerUser
	}
	if req.MinOrderValue != nil {
		c.MinOrderValue = *req.MinOrderValue
	}
	if req.StartAt != nil {
		st, _ := utils.ParseDateTimeForInput(*req.StartAt)
		c.StartAt = st.UTC()
	}
	if req.EndAt != nil {
		et, _ := utils.ParseDateTimeForInput(*req.EndAt)
		c.EndAt = et.UTC()
	}

	c.UpdatedAt = time.Now().UTC()

	err = cr.Update(db, c)
	if err != nil {
		msg, ok := errors.IsDuplicateKeyError(err)
		if ok {
			resp.Title = msg
			resp.Status = http.StatusConflict
			resp.Code = errors.CouponAlreadyExists
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Title = "Coupon updated"
	resp.Data = c
	return resp.ServerJSON(ctx)
}

func deleteCoupon(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)
	couponID := ctx.Param("coupon_id")

	resp := core.Response{}

	db := app.DB()

	cr := data.NewCouponRepository()
	if err := cr.Delete(db, storeID, couponID); err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Coupon not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CouponNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusNoContent
	return resp.ServerJSON(ctx)
}

func getCoupon(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)
	couponID := ctx.Param("coupon_id")

	resp := core.Response{}

	db := app.DB()

	cr := data.NewCouponRepository()
	v, err := cr.Get(db, storeID, couponID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Coupon not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CouponNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	users, err := cr.ListUsers(db, storeID, v.ID)
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = map[string]interface{}{
		"coupon": v,
		"users":  users,
	}
	return resp.ServerJSON(ctx)
}

func listCoupons(ctx echo.Context) error {
	pageQ := ctx.Request().URL.Query().Get("page")
	limitQ := ctx.Request().URL.Query().Get("limit")
	query := ctx.Request().URL.Query().Get("query")

	page, err := strconv.ParseInt(pageQ, 10, 64)
	if err != nil {
		page = 1
	}
	limit, err := strconv.ParseInt(limitQ, 10, 64)
	if err != nil {
		limit = 10
	}

	resp := core.Response{}

	var r interface{}

	if query == "" {
		r, err = fetchCoupons(ctx, page, limit)
	} else {
		r, err = searchCoupons(ctx, query, page, limit)
	}

	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = r
	return resp.ServerJSON(ctx)
}

func fetchCoupons(ctx echo.Context, page int64, limit int64) (interface{}, error) {
	from := (page - 1) * limit
	cr := data.NewCouponRepository()
	db := app.DB()
	return cr.List(db, ctx.Get(utils.StoreID).(string), int(from), int(limit))
}

func searchCoupons(ctx echo.Context, query string, page int64, limit int64) (interface{}, error) {
	from := (page - 1) * limit
	cr := data.NewCouponRepository()
	db := app.DB()
	return cr.Search(db, ctx.Get(utils.StoreID).(string), query, int(from), int(limit))
}

func checkCouponAvailability(ctx echo.Context) error {
	couponCode := ctx.Param("coupon_code")
	storeID := ctx.QueryParam("store_id")
	orderAmountQ := ctx.QueryParam("order_amount")
	shippingCostQ := ctx.QueryParam("shipping_cost")

	orderAmount, err := strconv.ParseInt(orderAmountQ, 10, 64)
	if err != nil {
		orderAmount = 0
	}

	shippingCost, err := strconv.ParseInt(shippingCostQ, 10, 64)
	if err != nil {
		shippingCost = 0
	}

	resp := core.Response{}

	db := app.DB()

	cu := data.NewCouponRepository()

	coupon, err := cu.GetByCode(db, storeID, couponCode)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Coupon not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CouponNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if !coupon.IsValid() {
		resp.Title = "Coupon is invalid"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.InvalidCoupon
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if coupon.IsUserSpecific {
		ok, err := cu.HasUser(db, storeID, coupon.ID, utils.GetUserID(ctx))
		if err != nil {
			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		if !ok {
			resp.Title = "Coupon not applicable for the user"
			resp.Status = http.StatusForbidden
			resp.Code = errors.CouponNotApplicable
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}
	}

	usagePerUser, err := cu.GetUsage(db, coupon.ID, utils.GetUserID(ctx))
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if coupon.MaxUsagePerUser != 0 && usagePerUser > coupon.MaxUsagePerUser {
		resp.Title = "Coupon usage limit per user exceed"
		resp.Status = http.StatusForbidden
		resp.Code = errors.CouponNotApplicable
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	usage, err := cu.GetTotalUsage(db, coupon.ID)
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if coupon.MaxUsage != 0 && usage > coupon.MaxUsage {
		resp.Title = "Coupon usage limit exceed"
		resp.Status = http.StatusForbidden
		resp.Code = errors.CouponNotApplicable
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	result := map[string]interface{}{}
	result["discount_type"] = coupon.DiscountType

	switch coupon.DiscountType {
	case models.TotalDiscount:
		result["discount_amount"] = coupon.CalculateDiscount(orderAmount + shippingCost)
	case models.ShippingDiscount:
		result["discount_amount"] = coupon.CalculateDiscount(shippingCost)
	case models.ProductDiscount:
		result["discount_amount"] = coupon.CalculateDiscount(orderAmount)
	default:
		result["discount_amount"] = 0
	}

	resp.Status = http.StatusOK
	resp.Data = result
	return resp.ServerJSON(ctx)
}

func addCouponUsers(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)
	couponID := ctx.Param("coupon_id")

	resp := core.Response{}

	body := struct {
		Users []string
	}{}

	if err := ctx.Bind(&body); err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.ProductCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB().Begin()

	cr := data.NewCouponRepository()
	v, err := cr.Get(db, storeID, couponID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Coupon not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CouponNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	for _, u := range body.Users {
		cf := models.CouponFor{
			UserID:   u,
			CouponID: v.ID,
		}
		if err := cr.AddUser(db, &cf); err != nil {
			db.Rollback()

			msg, ok := errors.IsDuplicateKeyError(err)
			if ok {
				resp.Title = msg
				resp.Status = http.StatusConflict
				resp.Code = errors.CouponNotFound
				resp.Errors = err
				return resp.ServerJSON(ctx)
			}

			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}
	}

	if err := db.Commit().Error; err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	return resp.ServerJSON(ctx)
}

func removeCouponUsers(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)
	couponID := ctx.Param("coupon_id")

	resp := core.Response{}

	body := struct {
		Users []string
	}{}

	if err := ctx.Bind(&body); err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.ProductCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB().Begin()

	cr := data.NewCouponRepository()
	v, err := cr.Get(db, storeID, couponID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Coupon not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CouponNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	for _, u := range body.Users {
		cf := models.CouponFor{
			UserID:   u,
			CouponID: v.ID,
		}
		if err := cr.RemoveUser(db, &cf); err != nil {
			db.Rollback()

			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}
	}

	if err := db.Commit().Error; err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusNotFound
	return resp.ServerJSON(ctx)
}
