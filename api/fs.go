package api

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/middlewares"
	"github.com/shopicano/shopicano-backend/services"
	"github.com/shopicano/shopicano-backend/utils"
	"github.com/shopicano/shopicano-backend/values"
	"net/http"
	"strings"
)

func RegisterFSRoutes(publicEndpoints, platformEndpoints *echo.Group) {
	fsPublicPath := publicEndpoints.Group("/fs")

	fsPublicPath.GET("/:bucket_name/:file_name/", serveAsStream)

	func(g echo.Group) {
		g.Use(middlewares.JWTAuth())
		g.POST("/:bucket_name/", upload)
	}(*fsPublicPath)
}

func serveAsStream(ctx echo.Context) error {
	bucketName := ctx.Param("bucket_name")
	fileName := ctx.Param("file_name")

	resp := core.Response{}

	f, err := services.ServeAsStreamFromMinio(fmt.Sprintf("%s/%s", bucketName, fileName))

	if err != nil {
		resp.Title = "Minio service failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.MinioServiceFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	return resp.ServeStreamFromMinio(ctx, f)
}

func upload(ctx echo.Context) error {
	bucketName := ctx.Param("bucket_name")

	resp := core.Response{}

	if bucketName == values.ReservedBucketName {
		resp.Title = "Unauthorized request"
		resp.Status = http.StatusForbidden
		resp.Code = errors.RestrictedBucket
		return resp.ServerJSON(ctx)
	}

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

	fileName := h.Filename
	extSeparatorIndex := strings.LastIndex(fileName, ".")
	fileName = base64.StdEncoding.EncodeToString([]byte(fileName[:extSeparatorIndex])) + "." + fileName[extSeparatorIndex+1:]

	newFileNameWithBucket := fmt.Sprintf("%s/%s-%s", bucketName, utils.NewUUID(), fileName)
	contentType := h.Header.Get("Content-Type")
	errU := services.UploadToMinio(newFileNameWithBucket, contentType, bytes.NewReader(body), h.Size)
	if errU != nil {
		resp.Title = "Minio service failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.MinioServiceFailed
		resp.Errors = errU
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusCreated
	resp.Data = map[string]interface{}{
		"path": newFileNameWithBucket,
	}
	return resp.ServerJSON(ctx)
}
