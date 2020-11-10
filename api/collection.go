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

func RegisterCollectionRoutes(publicEndpoints, platformEndpoints *echo.Group) {
	collectionsPublicPath := publicEndpoints.Group("/collections")
	collectionsPlatformPath := platformEndpoints.Group("/collections")

	func(g echo.Group) {
		g.GET("/", listCollections)
		g.GET("/:collection_id/", getCollection)
		g.GET("/:collection_id/products/", listProductsByCollection)
	}(*collectionsPublicPath)

	func(g echo.Group) {
		g.Use(middlewares.HasStore())
		g.Use(middlewares.IsStoreActive())
		g.Use(middlewares.IsStoreManager())
		g.POST("/", createCollection)
		g.DELETE("/:collection_id/", deleteCollection)
		g.PATCH("/:collection_id/", updateCollection)
		g.GET("/:collection_id/", getCollectionAsStoreOwner)
		g.PATCH("/:collection_id/products/", addCollectionProducts)
		g.DELETE("/:collection_id/products/", removeCollectionProducts)
		g.GET("/:collection_id/products/", listProductsByCollectionAsStoreStuff)
		g.GET("/", listCollectionsAsStoreOwner)
	}(*collectionsPlatformPath)
}

func createCollection(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)

	c, err := validators.ValidateCreateCollection(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.CollectionCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	c.StoreID = storeID

	db := app.DB()

	cu := data.NewCollectionRepository()
	if err := cu.Create(db, c); err != nil {
		msg, ok := errors.IsDuplicateKeyError(err)
		if ok {
			resp.Title = msg
			resp.Status = http.StatusConflict
			resp.Code = errors.CollectionAlreadyExists
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

func deleteCollection(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)
	collectionID := ctx.Param("collection_id")

	resp := core.Response{}

	db := app.DB()

	cu := data.NewCollectionRepository()
	if err := cu.Delete(db, storeID, collectionID); err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Collection not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CollectionNotFound
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

func listCollections(ctx echo.Context) error {
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

	var collections []models.Collection

	if query == "" {
		collections, err = fetchCollections(ctx, page, limit, true)
	} else {
		collections, err = searchCollections(ctx, query, page, limit, true)
	}

	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = collections
	return resp.ServerJSON(ctx)
}

func listCollectionsAsStoreOwner(ctx echo.Context) error {
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

	var collections []models.Collection

	if query == "" {
		collections, err = fetchCollections(ctx, page, limit, false)
	} else {
		collections, err = searchCollections(ctx, query, page, limit, false)
	}

	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = collections
	return resp.ServerJSON(ctx)
}

func searchCollections(ctx echo.Context, query string, page, limit int64, isPublic bool) ([]models.Collection, error) {
	db := app.DB()

	from := (page - 1) * limit
	cu := data.NewCollectionRepository()
	if isPublic {
		return cu.Search(db, query, int(from), int(limit))
	}
	return cu.SearchAsStoreStuff(db, ctx.Get(utils.StoreID).(string), query, int(from), int(limit))
}

func fetchCollections(ctx echo.Context, page, limit int64, isPublic bool) ([]models.Collection, error) {
	db := app.DB()

	from := (page - 1) * limit
	cu := data.NewCollectionRepository()
	if isPublic {
		return cu.List(db, int(from), int(limit))
	}
	return cu.ListAsStoreStuff(db, ctx.Get(utils.StoreID).(string), int(from), int(limit))
}

func updateCollection(ctx echo.Context) error {
	collectionID := ctx.Param("collection_id")
	storeID := ctx.Get(utils.StoreID).(string)

	pld, err := validators.ValidateUpdateCollection(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.CollectionCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()
	cu := data.NewCollectionRepository()

	c, err := cu.GetAsStoreOwner(db, storeID, collectionID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Collection not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CollectionNotFound
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

func getCollection(ctx echo.Context) error {
	collectionID := ctx.Param("collection_id")

	resp := core.Response{}

	db := app.DB()
	cu := data.NewCollectionRepository()

	c, err := cu.Get(db, collectionID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Collection not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CollectionNotFound
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

func getCollectionAsStoreOwner(ctx echo.Context) error {
	collectionID := ctx.Param("collection_id")
	storeID := ctx.Get(utils.StoreID).(string)

	resp := core.Response{}

	db := app.DB()
	cu := data.NewCollectionRepository()

	c, err := cu.GetAsStoreOwner(db, storeID, collectionID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Collection not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CollectionNotFound
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

func addCollectionProducts(ctx echo.Context) error {
	collectionID := ctx.Param("collection_id")
	storeID := ctx.Get(utils.StoreID).(string)

	resp := core.Response{}

	pld := struct {
		ProductIDs []string `json:"product_ids"`
	}{}

	if err := ctx.Bind(&pld); err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.ProductCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB().Begin()
	cu := data.NewCollectionRepository()

	c, err := cu.GetAsStoreOwner(db, storeID, collectionID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Collection not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CollectionNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	for _, p := range pld.ProductIDs {
		if err := cu.AddProducts(db, &models.CollectionOfProduct{
			CollectionID: c.ID,
			ProductID:    p,
		}); err != nil {
			db.Rollback()

			msg, ok := errors.IsDuplicateKeyError(err)
			if ok {
				resp.Title = msg
				resp.Status = http.StatusConflict
				resp.Code = errors.ProductAlreadyExists
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

	resp.Status = http.StatusCreated
	return resp.ServerJSON(ctx)
}

func removeCollectionProducts(ctx echo.Context) error {
	collectionID := ctx.Param("collection_id")
	storeID := ctx.Get(utils.StoreID).(string)

	resp := core.Response{}

	pld := struct {
		ProductIDs []string `json:"product_ids"`
	}{}

	if err := ctx.Bind(&pld); err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.ProductCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB().Begin()
	cu := data.NewCollectionRepository()

	c, err := cu.GetAsStoreOwner(db, storeID, collectionID)
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	for _, p := range pld.ProductIDs {
		if err := cu.RemoveProducts(db, &models.CollectionOfProduct{
			CollectionID: c.ID,
			ProductID:    p,
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

func listProductsByCollection(ctx echo.Context) error {
	collectionID := ctx.Param("collection_id")

	resp := core.Response{}

	db := app.DB()
	cu := data.NewCollectionRepository()
	pu := data.NewProductRepository()

	c, err := cu.Get(db, collectionID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Collection not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CollectionNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	pageQ := ctx.Request().URL.Query().Get("page")
	limitQ := ctx.Request().URL.Query().Get("limit")

	page, err := strconv.ParseInt(pageQ, 10, 64)
	if err != nil {
		page = 1
	}
	limit, err := strconv.ParseInt(limitQ, 10, 64)
	if err != nil {
		limit = 10
	}

	from := (page * limit) - limit
	products, err := pu.ListByCollection(db, c.ID, int(from), int(limit))
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = products
	return resp.ServerJSON(ctx)
}

func listProductsByCollectionAsStoreStuff(ctx echo.Context) error {
	collectionID := ctx.Param("collection_id")

	resp := core.Response{}

	db := app.DB()
	cu := data.NewCollectionRepository()
	pu := data.NewProductRepository()

	c, err := cu.GetAsStoreOwner(db, utils.GetStoreID(ctx), collectionID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Collection not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.CollectionNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	pageQ := ctx.Request().URL.Query().Get("page")
	limitQ := ctx.Request().URL.Query().Get("limit")

	page, err := strconv.ParseInt(pageQ, 10, 64)
	if err != nil {
		page = 1
	}
	limit, err := strconv.ParseInt(limitQ, 10, 64)
	if err != nil {
		limit = 10
	}

	from := (page * limit) - limit
	products, err := pu.ListByCollectionAsStoreStuff(db, c.ID, int(from), int(limit))
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = products
	return resp.ServerJSON(ctx)
}
