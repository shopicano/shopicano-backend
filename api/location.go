package api

import (
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/data"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/middlewares"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/validators"
	"net/http"
	"strconv"
	"strings"
)

func RegisterLocationRoutes(publicEndpoints, platformEndpoints *echo.Group) {
	locationsPublicPath := publicEndpoints.Group("/locations")
	locationsPlatformPath := platformEndpoints.Group("/locations")

	func(g echo.Group) {
		g.Use(middlewares.JWTAuth())
		g.GET("/", listLocationsForPublic)
		g.GET("/:location_id/shipping-methods/", listShippingMethodsByLocationForUser)
		g.GET("/:location_id/payment-methods/", listPaymentMethodsByLocationForUser)
	}(*locationsPublicPath)

	func(g echo.Group) {
		g.Use(middlewares.IsPlatformManager)
		g.GET("/:location_id/shipping-methods/", listShippingMethodsByLocation)
		g.GET("/:location_id/payment-methods/", listPaymentMethodsByLocation)
		g.PATCH("/:location_id/", updateLocation)
		g.DELETE("/:location_id/", deleteLocationParams)
		g.PATCH("/", updateAllLocations)
		g.GET("/", listLocations)
	}(*locationsPlatformPath)
}

func listLocations(ctx echo.Context) error {
	resp := core.Response{}

	name := ctx.QueryParam("name")
	//locationTypeQ := ctx.QueryParam("type")
	var locationType models.LocationType
	//if locationTypeQ == "city" {
	//	locationType = models.LocationTypeCity
	//} else if locationTypeQ == "state" {
	//	locationType = models.LocationTypeState
	//} else {
	//	locationType = models.LocationTypeCountry
	//}

	locationType = models.LocationTypeCountry

	parentIDQ := ctx.QueryParam("parent_id")
	parentID, err := strconv.ParseInt(parentIDQ, 10, 64)
	if locationType != models.LocationTypeCountry && err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.InvalidRequest
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	var locations []models.Location

	db := app.DB()
	locations, err = listLocationsForAdmin(name, locationType, parentID, db)
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = locations
	return resp.ServerJSON(ctx)
}

func listLocationsForPublic(ctx echo.Context) error {
	resp := core.Response{}

	name := ctx.QueryParam("name")
	//locationTypeQ := ctx.QueryParam("type")
	var locationType models.LocationType
	//if locationTypeQ == "city" {
	//	locationType = models.LocationTypeCity
	//} else if locationTypeQ == "state" {
	//	locationType = models.LocationTypeState
	//} else {
	//	locationType = models.LocationTypeCountry
	//}

	locationType = models.LocationTypeCountry

	parentIDQ := ctx.QueryParam("parent_id")
	parentID, err := strconv.ParseInt(parentIDQ, 10, 64)
	if locationType != models.LocationTypeCountry && err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.InvalidRequest
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	var locations []models.Location

	db := app.DB()
	locations, err = listLocationsForUser(name, locationType, parentID, db)
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = locations
	return resp.ServerJSON(ctx)
}

func listLocationsForAdmin(name string, locationType models.LocationType, parentID int64, db *gorm.DB) ([]models.Location, error) {
	locDao := data.NewLocationRepository()

	where := "(location_type = ? AND parent_id = ?)"
	var args []interface{}
	args = append(args, locationType)
	args = append(args, parentID)

	if name != "" {
		where += " AND (LOWER(name) LIKE ? OR LOWER(iso_name) = ?)"
		args = append(args, "%"+strings.ToLower(name)+"%", strings.ToLower(name))
	}

	return locDao.List(db, where, args)
}

func listLocationsForUser(name string, locationType models.LocationType, parentID int64, db *gorm.DB) ([]models.Location, error) {
	locDao := data.NewLocationRepository()

	where := "(location_type = ? AND parent_id = ? AND is_published = ?)"
	var args []interface{}
	args = append(args, locationType)
	args = append(args, parentID)
	args = append(args, 1)

	if name != "" {
		where += " AND (LOWER(name) LIKE ? OR LOWER(iso_name) = ?)"
		args = append(args, "%"+strings.ToLower(name)+"%", strings.ToLower(name))
	}

	return locDao.List(db, where, args)
}

func updateLocation(ctx echo.Context) error {
	resp := core.Response{}

	req, err := validators.ValidateUpdateLocation(ctx, true)
	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.InvalidRequest
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	locationIDQ := ctx.Param("location_id")
	locationID, _ := strconv.ParseInt(locationIDQ, 10, 64)

	db := app.DB().Begin()
	locDao := data.NewLocationRepository()

	loc, err := locDao.FindByID(db, int(locationID))
	if err != nil {
		db.Rollback()

		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Location not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.LocationNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if req.IsPublished != nil {
		if *req.IsPublished {
			loc.IsPublished = 1
		} else {
			loc.IsPublished = 0
		}
	}

	for _, v := range req.ShippingMethods {
		if err := locDao.AddShippingMethod(db, &models.ShippingForLocation{
			LocationID:       locationID,
			ShippingMethodID: v,
		}); err != nil {
			db.Rollback()

			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}
	}

	for _, v := range req.PaymentMethods {
		if err := locDao.AddPaymentMethod(db, &models.PaymentForLocation{
			LocationID:      locationID,
			PaymentMethodID: v,
		}); err != nil {
			db.Rollback()

			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}
	}

	err = locDao.UpdateByID(db, loc)
	if err != nil {
		db.Rollback()

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
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

func deleteLocationParams(ctx echo.Context) error {
	resp := core.Response{}

	req, err := validators.ValidateUpdateLocation(ctx, true)
	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.InvalidRequest
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	locationIDQ := ctx.Param("location_id")
	locationID, _ := strconv.ParseInt(locationIDQ, 10, 64)

	db := app.DB().Begin()
	locDao := data.NewLocationRepository()

	for _, v := range req.ShippingMethods {
		if err := locDao.RemoveShippingMethod(db, &models.ShippingForLocation{
			LocationID:       locationID,
			ShippingMethodID: v,
		}); err != nil {
			db.Rollback()

			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}
	}

	for _, v := range req.PaymentMethods {
		if err := locDao.RemovePaymentMethod(db, &models.PaymentForLocation{
			LocationID:      locationID,
			PaymentMethodID: v,
		}); err != nil {
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

	resp.Status = http.StatusNoContent
	return resp.ServerJSON(ctx)
}

func updateAllLocations(ctx echo.Context) error {
	resp := core.Response{}

	req, err := validators.ValidateUpdateLocation(ctx, false)
	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.InvalidRequest
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB().Begin()
	locDao := data.NewLocationRepository()

	locs, err := locDao.Find(db)
	if err != nil {
		db.Rollback()

		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Location not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.LocationNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	for _, loc := range locs {
		if req.IsPublished != nil {
			if *req.IsPublished {
				loc.IsPublished = 1
			} else {
				loc.IsPublished = 0
			}
		}

		for _, v := range req.ShippingMethods {
			if err := locDao.AddShippingMethod(db, &models.ShippingForLocation{
				LocationID:       loc.ID,
				ShippingMethodID: v,
			}); err != nil {
				db.Rollback()

				resp.Title = "Database query failed"
				resp.Status = http.StatusInternalServerError
				resp.Code = errors.DatabaseQueryFailed
				resp.Errors = err
				return resp.ServerJSON(ctx)
			}
		}

		for _, v := range req.PaymentMethods {
			if err := locDao.AddPaymentMethod(db, &models.PaymentForLocation{
				LocationID:      loc.ID,
				PaymentMethodID: v,
			}); err != nil {
				db.Rollback()

				resp.Title = "Database query failed"
				resp.Status = http.StatusInternalServerError
				resp.Code = errors.DatabaseQueryFailed
				resp.Errors = err
				return resp.ServerJSON(ctx)
			}
		}

		err = locDao.UpdateByID(db, &loc)
		if err != nil {
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

	resp.Status = http.StatusOK
	return resp.ServerJSON(ctx)
}

func listShippingMethodsByLocation(ctx echo.Context) error {
	resp := core.Response{}

	locationIDQ := ctx.Param("location_id")
	locationID, _ := strconv.ParseInt(locationIDQ, 10, 64)

	db := app.DB()
	marketDao := data.NewMarketplaceRepository()
	m, err := marketDao.ListShippingMethodsByLocation(db, locationID)
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = m
	return resp.ServerJSON(ctx)
}

func listPaymentMethodsByLocation(ctx echo.Context) error {
	resp := core.Response{}

	locationIDQ := ctx.Param("location_id")
	locationID, _ := strconv.ParseInt(locationIDQ, 10, 64)

	db := app.DB()
	marketDao := data.NewMarketplaceRepository()
	m, err := marketDao.ListPaymentMethodsByLocation(db, locationID)
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = m
	return resp.ServerJSON(ctx)
}

func listShippingMethodsByLocationForUser(ctx echo.Context) error {
	resp := core.Response{}

	locationIDQ := ctx.Param("location_id")
	locationID, _ := strconv.ParseInt(locationIDQ, 10, 64)

	db := app.DB()
	marketDao := data.NewMarketplaceRepository()
	m, err := marketDao.ListShippingMethodsByLocationForUser(db, locationID)
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = m
	return resp.ServerJSON(ctx)
}

func listPaymentMethodsByLocationForUser(ctx echo.Context) error {
	resp := core.Response{}

	locationIDQ := ctx.Param("location_id")
	locationID, _ := strconv.ParseInt(locationIDQ, 10, 64)

	db := app.DB()
	marketDao := data.NewMarketplaceRepository()
	m, err := marketDao.ListPaymentMethodsByLocationForUser(db, locationID)
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = m
	return resp.ServerJSON(ctx)
}
