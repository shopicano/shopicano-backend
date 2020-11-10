package api

import (
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/data"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/middlewares"
	"github.com/shopicano/shopicano-backend/utils"
	"github.com/shopicano/shopicano-backend/validators"
	"net/http"
	"strconv"
	"time"
)

func RegisterCategoryRoutes(publicEndpoints, platformEndpoints *echo.Group) {
	categoryPublicPath := publicEndpoints.Group("/categories")
	categoryPlatformPath := platformEndpoints.Group("/categories")

	func(g echo.Group) {
		g.GET("/", listCategories)
		g.GET("/:category_id/", getCategory)
	}(*categoryPublicPath)

	func(g echo.Group) {
		g.Use(middlewares.HasStore())
		g.Use(middlewares.IsStoreActive())
		g.Use(middlewares.IsStoreManager())
		g.POST("/", createCategory)
		g.DELETE("/:category_id/", deleteCategory)
		g.PATCH("/:category_id/", updateCategory)
		g.GET("/:category_id/", getCategoryAsOwner)
		g.GET("/", listCategoriesAsStoreOwner)
	}(*categoryPlatformPath)
}

func createCategory(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)

	c, err := validators.ValidateCreateCategory(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.CategoryCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()

	c.StoreID = storeID

	cu := data.NewCategoryRepository()
	if err := cu.Create(db, c); err != nil {
		msg, ok := errors.IsDuplicateKeyError(err)
		if ok {
			resp.Title = msg
			resp.Status = http.StatusConflict
			resp.Code = errors.CategoryAlreadyExists
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
	resp.Data = c
	return resp.ServerJSON(ctx)
}

func listCategories(ctx echo.Context) error {
	pageQ := ctx.Request().URL.Query().Get("page")
	limitQ := ctx.Request().URL.Query().Get("limit")
	query := ctx.Request().URL.Query().Get("query")

	var err error

	page, err := strconv.ParseInt(pageQ, 10, 64)
	if err != nil {
		page = 1
	}
	limit, err := strconv.ParseInt(limitQ, 10, 64)
	if err != nil {
		limit = 10
	}

	resp := core.Response{}

	var categories interface{}

	if query == "" {
		categories, err = fetchCategories(ctx, page, limit, true)
	} else {
		categories, err = searchCategories(ctx, query, page, limit, true)
	}

	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = categories
	return resp.ServerJSON(ctx)
}

func listCategoriesAsStoreOwner(ctx echo.Context) error {
	pageQ := ctx.Request().URL.Query().Get("page")
	limitQ := ctx.Request().URL.Query().Get("limit")
	query := ctx.Request().URL.Query().Get("query")

	var err error

	page, err := strconv.ParseInt(pageQ, 10, 64)
	if err != nil {
		page = 1
	}
	limit, err := strconv.ParseInt(limitQ, 10, 64)
	if err != nil {
		limit = 10
	}

	resp := core.Response{}

	var categories interface{}

	if query == "" {
		categories, err = fetchCategories(ctx, page, limit, false)
	} else {
		categories, err = searchCategories(ctx, query, page, limit, false)
	}

	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = categories
	return resp.ServerJSON(ctx)
}

func searchCategories(ctx echo.Context, query string, page, limit int64, isPublic bool) (interface{}, error) {
	db := app.DB()

	from := (page - 1) * limit
	cu := data.NewCategoryRepository()
	if isPublic {
		return cu.Search(db, query, int(from), int(limit))
	}
	return cu.SearchAsStoreStuff(db, ctx.Get(utils.StoreID).(string), query, int(from), int(limit))
}

func fetchCategories(ctx echo.Context, page, limit int64, isPublic bool) (interface{}, error) {
	db := app.DB()

	from := (page - 1) * limit
	cu := data.NewCategoryRepository()
	if isPublic {
		return cu.List(db, int(from), int(limit))
	}
	return cu.ListAsStoreStuff(db, ctx.Get(utils.StoreID).(string), int(from), int(limit))
}

func deleteCategory(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)
	categoryID := ctx.Param("id")

	resp := core.Response{}

	db := app.DB()

	cu := data.NewCategoryRepository()
	if err := cu.Delete(db, storeID, categoryID); err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Category not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CategoryNotFound
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

func updateCategory(ctx echo.Context) error {
	categoryID := ctx.Param("category_id")
	storeID := ctx.Get(utils.StoreID).(string)

	pld, err := validators.ValidateUpdateCategory(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.CategoryCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()
	cu := data.NewCategoryRepository()

	c, err := cu.GetAsStoreOwner(db, storeID, categoryID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Category not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CategoryNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if pld.Name != nil {
		c.Name = *pld.Name
	}
	if pld.Description != nil {
		c.Description = *pld.Description
	}
	if pld.Image != nil {
		c.Image = *pld.Image
	}
	if pld.IsPublished != nil {
		c.IsPublished = *pld.IsPublished
	}

	c.UpdatedAt = time.Now().UTC()

	if err := cu.Update(db, c); err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = c
	return resp.ServerJSON(ctx)
}

func getCategory(ctx echo.Context) error {
	categoryID := ctx.Param("category_id")

	resp := core.Response{}

	db := app.DB()
	cu := data.NewCategoryRepository()

	c, err := cu.Get(db, categoryID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Category not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CategoryNotFound
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
	resp.Data = c
	return resp.ServerJSON(ctx)
}

func getCategoryAsOwner(ctx echo.Context) error {
	categoryID := ctx.Param("category_id")
	storeID := ctx.Get(utils.StoreID).(string)

	resp := core.Response{}

	db := app.DB()
	cu := data.NewCategoryRepository()

	c, err := cu.GetAsStoreOwner(db, storeID, categoryID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Category not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CategoryNotFound
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
	resp.Data = c
	return resp.ServerJSON(ctx)
}
