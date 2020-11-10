package api

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/data"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/middlewares"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/services"
	"github.com/shopicano/shopicano-backend/utils"
	"github.com/shopicano/shopicano-backend/validators"
	"github.com/shopicano/shopicano-backend/values"
)

func RegisterProductRoutes(publicEndpoints, platformEndpoints *echo.Group) {
	productsPublicPath := publicEndpoints.Group("/products")
	productsPlatformPath := platformEndpoints.Group("/products")

	func(g echo.Group) {
		g.GET("/", listProducts)
		g.GET("/:product_id/", getProduct)
	}(*productsPublicPath)

	func(g echo.Group) {
		g.Use(middlewares.HasStore())
		g.Use(middlewares.IsStoreActive())
		g.Use(middlewares.IsStoreManager())
		g.POST("/", createProduct)
		g.PATCH("/:product_id/", updateProduct)
		g.DELETE("/:product_id/", deleteProduct)
		g.GET("/:product_id/", getProductAsStoreOwner)
		g.GET("/", listProductsAsStoreOwner)
		g.PUT("/:product_id/attributes/", addProductAttribute)
		g.DELETE("/:product_id/attributes/:attribute_id/", deleteProductAttribute)
		g.GET("/:product_id/download/", downloadProduct)
		g.POST("/:product_id/upload/", saveDownloadableProduct)
	}(*productsPlatformPath)
}

func createProduct(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)

	req, err := validators.ValidateCreateProduct(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.ProductCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if req.CategoryID != nil && *req.CategoryID == "" {
		req.CategoryID = nil
	}

	p := models.Product{
		ID:               utils.NewUUID(),
		StoreID:          storeID,
		Price:            req.Price,
		ProductCost:      req.ProductCost,
		Stock:            req.Stock,
		Name:             req.Name,
		Slug:             slug.Make(req.Name),
		IsShippable:      req.IsShippable,
		CategoryID:       req.CategoryID,
		IsPublished:      req.IsPublished,
		IsDigital:        req.IsDigital,
		MaxQuantityCount: req.MaxQuantityCount,
		SKU:              req.SKU,
		Unit:             req.Unit,
		Image:            req.Image,
		Description:      req.Description,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	db := app.DB().Begin()

	pu := data.NewProductRepository()
	err = pu.Create(db, &p)
	if err != nil {
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

	for _, i := range req.AdditionalImages {
		if strings.TrimSpace(i) == "" {
			continue
		}

		if err := pu.AddImage(db, p.ID, strings.TrimSpace(i)); err != nil {
			db.Rollback()

			resp.Title = "Failed to add additional image"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}
	}

	if err := db.Commit().Error; err != nil {
		resp.Title = "Failed to commit data"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusCreated
	resp.Title = "Product created"
	resp.Data = p
	return resp.ServerJSON(ctx)
}

func updateProduct(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)
	productID := ctx.Param("product_id")

	req, err := validators.ValidateUpdateProduct(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.ProductCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if req.CategoryID != nil && *req.CategoryID == "" {
		req.CategoryID = nil
	}

	db := app.DB().Begin()

	pu := data.NewProductRepository()

	p, err := pu.GetAsStoreStuff(db, storeID, productID)
	if err != nil {
		db.Rollback()

		resp.Title = "Product not found"
		resp.Status = http.StatusNotFound
		resp.Code = errors.ProductNotFound
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if req.Name != nil {
		p.Name = *req.Name
		p.Slug = slug.Make(*req.Name)
	}
	if req.Price != nil {
		p.Price = *req.Price
	}
	if req.ProductCost != nil {
		p.ProductCost = *req.ProductCost
	}
	if req.IsShippable != nil {
		p.IsShippable = *req.IsShippable
	}
	if req.IsDigital != nil {
		p.IsDigital = *req.IsDigital
	}
	if req.IsPublished != nil {
		p.IsPublished = *req.IsPublished
	}
	if req.Stock != nil {
		p.Stock = *req.Stock
	}
	if req.CategoryID != nil {
		p.CategoryID = req.CategoryID
	}
	if req.Image != nil {
		p.Image = *req.Image
	}
	if len(req.AdditionalImages) > 0 {
		if err := pu.RemoveImage(db, p.ID); err != nil {
			db.Rollback()

			resp.Title = "Failed to remove additional image"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		for _, i := range req.AdditionalImages {
			if strings.TrimSpace(i) == "" {
				continue
			}

			if err := pu.AddImage(db, p.ID, strings.TrimSpace(i)); err != nil {
				db.Rollback()

				resp.Title = "Failed to add additional image"
				resp.Status = http.StatusInternalServerError
				resp.Code = errors.DatabaseQueryFailed
				resp.Errors = err
				return resp.ServerJSON(ctx)
			}
		}
	}

	if req.SKU != nil {
		p.SKU = *req.SKU
	}
	if req.Unit != nil {
		p.Unit = *req.Unit
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.DigitalDownloadLink != nil {
		p.DigitalDownloadLink = *req.DigitalDownloadLink
	}
	if req.MaxQuantityCount != nil {
		p.MaxQuantityCount = *req.MaxQuantityCount
	}

	p.UpdatedAt = time.Now().UTC()

	err = pu.Update(db, p)
	if err != nil {
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

	if err := db.Commit().Error; err != nil {
		resp.Title = "Failed to commit data"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Title = "Product updated"
	resp.Data = p
	return resp.ServerJSON(ctx)
}

func deleteProduct(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)
	productID := ctx.Param("product_id")

	resp := core.Response{}

	db := app.DB().Begin()

	pu := data.NewProductRepository()
	if err := pu.Delete(db, storeID, productID); err != nil {
		db.Rollback()

		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Product not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.ProductNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if err := pu.RemoveImage(db, productID); err != nil {
		db.Rollback()

		resp.Title = "Failed to remove additional image"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if err := db.Commit().Error; err != nil {
		resp.Title = "Failed to commit data"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusNoContent
	return resp.ServerJSON(ctx)
}

func getProduct(ctx echo.Context) error {
	productID := ctx.Param("product_id")

	resp := core.Response{}

	db := app.DB()

	pu := data.NewProductRepository()

	var p *models.ProductDetails
	var err error

	p, err = pu.GetDetails(db, productID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Product not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.ProductNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if err := pu.IncreaseViewCounter(db, p.ID, p.StoreID); err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = p
	return resp.ServerJSON(ctx)
}

func getProductAsStoreOwner(ctx echo.Context) error {
	productID := ctx.Param("product_id")

	resp := core.Response{}

	db := app.DB()

	pu := data.NewProductRepository()

	var p interface{}
	var err error

	p, err = pu.GetDetailsAsStoreStuff(db, ctx.Get(utils.StoreID).(string), productID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Product not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.ProductNotFound
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
	resp.Data = p
	return resp.ServerJSON(ctx)
}

func listProducts(ctx echo.Context) error {
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
		r, err = fetchProducts(ctx, page, limit, true)
	} else {
		r, err = searchProducts(ctx, query, page, limit, true)
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

func listProductsAsStoreOwner(ctx echo.Context) error {
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
		r, err = fetchProducts(ctx, page, limit, false)
	} else {
		r, err = searchProducts(ctx, query, page, limit, false)
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

func fetchProducts(ctx echo.Context, page int64, limit int64, isPublic bool) (interface{}, error) {
	from := (page - 1) * limit
	pu := data.NewProductRepository()

	db := app.DB()

	if isPublic {
		return pu.List(db, int(from), int(limit))
	}
	return pu.ListAsStoreStuff(db, ctx.Get(utils.StoreID).(string), int(from), int(limit))
}

func searchProducts(ctx echo.Context, query string, page int64, limit int64, isPublic bool) (interface{}, error) {
	from := (page - 1) * limit
	pu := data.NewProductRepository()

	db := app.DB()

	if isPublic {
		return pu.Search(db, query, int(from), int(limit))
	}
	return pu.SearchAsStoreStuff(db, ctx.Get(utils.StoreID).(string), query, int(from), int(limit))
}

func addProductAttribute(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)
	productID := ctx.Param("product_id")

	resp := core.Response{}

	req, err := validators.ValidateAddProductAttribute(ctx)
	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.ProductAttributeCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()
	pu := data.NewProductRepository()

	p, err := pu.GetAsStoreStuff(db, storeID, productID)
	if err != nil {
		resp.Title = "Product not found"
		resp.Status = http.StatusNotFound
		resp.Code = errors.ProductNotFound
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	v := models.ProductAttribute{
		ID:        utils.NewUUID(),
		ProductID: p.ID,
		Key:       req.Key,
		Value:     req.Value,
	}

	err = pu.AddAttribute(db, &v)
	if err != nil {
		msg, ok := errors.IsDuplicateKeyError(err)
		if ok {
			resp.Title = msg
			resp.Status = http.StatusConflict
			resp.Code = errors.ProductAttributeAlreadyExists
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
	resp.Title = "Product attribute added"
	return resp.ServerJSON(ctx)
}

func deleteProductAttribute(ctx echo.Context) error {
	storeID := ctx.Get(utils.StoreID).(string)
	productID := ctx.Param("product_id")
	attributeID := ctx.Param("attribute_id")

	resp := core.Response{}

	db := app.DB()
	pu := data.NewProductRepository()

	p, err := pu.GetAsStoreStuff(db, storeID, productID)
	if err != nil {
		resp.Title = "Product not found"
		resp.Status = http.StatusNotFound
		resp.Code = errors.ProductNotFound
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	err = pu.RemoveAttribute(db, p.ID, attributeID)
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusNoContent
	return resp.ServerJSON(ctx)
}

func saveDownloadableProduct(ctx echo.Context) error {
	storeID := utils.GetStoreID(ctx)
	productID := ctx.Param("product_id")

	resp := core.Response{}

	if err := ctx.Request().ParseMultipartForm(32 << 20); err != nil {
		resp.Title = "Couldn't parse multipart form"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.InvalidMultiPartBody
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	r := ctx.Request()
	r.Body = http.MaxBytesReader(ctx.Response(), r.Body, 32<<20) // 32 Mb

	f, h, e := r.FormFile("file")
	if e != nil {
		resp.Title = "No multipart file"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.InvalidMultiPartBody
		resp.Errors = e
		return resp.ServerJSON(ctx)
	}

	body := make([]byte, h.Size)
	_, errR := f.Read(body)
	if errR != nil {
		resp.Title = "Unable to read multipart data"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.UnableToReadMultiPartData
		resp.Errors = errR
		return resp.ServerJSON(ctx)
	}

	db := app.DB()
	pr := data.NewProductRepository()
	m, err := pr.Get(db, productID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Product not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.ProductNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	fileName := h.Filename
	extSeparatorIndex := strings.LastIndex(fileName, ".")
	fileName = base64.StdEncoding.EncodeToString([]byte(fileName[:extSeparatorIndex])) + "." + fileName[extSeparatorIndex+1:]

	newFileName := fmt.Sprintf("%s/%s-%s", storeID, utils.NewUUID(), fileName)
	newFileNameWithBucket := fmt.Sprintf("%s/%s", values.ReservedBucketName, newFileName)
	contentType := h.Header.Get("Content-Type")
	errU := services.UploadToMinio(newFileNameWithBucket, contentType, bytes.NewReader(body), h.Size)
	if errU != nil {
		resp.Title = "Minio service failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.MinioServiceFailed
		resp.Errors = errU
		return resp.ServerJSON(ctx)
	}

	m.DigitalDownloadLink = newFileName
	if err := pr.Update(db, m); err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Title = "Downloadable file has been added"
	return resp.ServerJSON(ctx)
}

func downloadProduct(ctx echo.Context) error {
	productID := ctx.Param("product_id")

	resp := core.Response{}

	db := app.DB()

	pu := data.NewProductRepository()
	m, err := pu.Get(db, productID)
	if err != nil {
		if errors.IsRecordNotFoundError(err) {
			resp.Title = "Product not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.ProductNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	f, err := services.ServeAsStreamFromMinio(fmt.Sprintf("%s/%s", values.ReservedBucketName, m.DigitalDownloadLink))

	if err != nil {
		resp.Title = "Minio service failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.MinioServiceFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	return resp.ServeStreamFromMinio(ctx, f)
}
