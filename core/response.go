package core

import (
	"fmt"
	"image"
	"io"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v6"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/log"
	"github.com/shopicano/shopicano-backend/utils"
)

type Response struct {
	Code   errors.ErrorCode `json:"code,omitempty"`
	Status int              `json:"-"`
	Title  string           `json:"title,omitempty"`
	Data   interface{}      `json:"data,omitempty"`
	Errors error            `json:"errors,omitempty"`
}

func (r *Response) ServerJSON(ctx echo.Context) error {
	ctx.Response().Header().Set("X-Platform", "Shopicano")
	ctx.Response().Header().Set("X-Platform-Developer", "www.codersgarage.com")
	ctx.Response().Header().Set("X-Platform-Connect", "www.shopicano.com")
	ctx.Response().Header().Set("Content-Type", "application/json")

	if err := ctx.JSON(r.Status, r); err != nil {
		return err
	}
	return nil
}

func (r *Response) ServeStreamFromMinioAsDownload(ctx echo.Context, object *minio.Object) error {
	s, _ := object.Stat()

	fileName := fmt.Sprintf("%s.%s", s.ETag, s.Key[strings.LastIndex(s.Key, ".")+1:])
	ctx.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	ctx.Response().Header().Set("Content-Type", s.ContentType)
	ctx.Response().Header().Set("X-Platform", "Shopicano")
	ctx.Response().Header().Set("X-Platform-Developer", "www.codersgarage.com")
	ctx.Response().Header().Set("X-Platform-Connect", "www.shopicano.com")

	if _, err := io.Copy(ctx.Response().Writer, object); err != nil {
		log.Log().Errorln(err)
		return err
	}
	return nil
}

func (r *Response) ServeStreamFromMinio(ctx echo.Context, object *minio.Object) error {
	s, _ := object.Stat()

	fileName := fmt.Sprintf("%s.%s", s.ETag, s.Key[strings.LastIndex(s.Key, ".")+1:])
	ctx.Response().Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", fileName))
	ctx.Response().Header().Set("Content-Type", s.ContentType)
	ctx.Response().Header().Set("X-Platform", "Shopicano")
	ctx.Response().Header().Set("X-Platform-Developer", "www.codersgarage.com")
	ctx.Response().Header().Set("X-Platform-Connect", "www.shopicano.com")

	if utils.IsImage(fileName) {
		return r.ServeStreamFromMinioAsImage(ctx, object)
	}

	if _, err := io.Copy(ctx.Response().Writer, object); err != nil {
		log.Log().Errorln(err)
		return err
	}
	return nil
}

func (r *Response) ServeStreamFromMinioAsImage(ctx echo.Context, object *minio.Object) error {
	s, _ := object.Stat()

	fileName := fmt.Sprintf("%s.%s", s.ETag, s.Key[strings.LastIndex(s.Key, ".")+1:])
	ctx.Response().Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", fileName))
	ctx.Response().Header().Set("Content-Type", s.ContentType)
	ctx.Response().Header().Set("X-Platform", "Shopicano")
	ctx.Response().Header().Set("X-Platform-Developer", "Coders Garage")
	ctx.Response().Header().Set("X-Platform-Connect", "www.shopicano.com")

	heightQ := ctx.QueryParam("height")
	widthQ := ctx.QueryParam("width")
	qualityQ := ctx.QueryParam("quality")

	height, _ := strconv.ParseInt(heightQ, 10, 64)
	width, _ := strconv.ParseInt(widthQ, 10, 64)
	quality, _ := strconv.ParseInt(qualityQ, 10, 64)

	var img image.Image
	var err error

	if height != 0 && width != 0 {
		img, err = imaging.Decode(object)
		if err != nil {
			return nil
		}
		img = imaging.Resize(img, int(width), int(height), imaging.Lanczos)
	} else if height != 0 && width == 0 {
		img, err = imaging.Decode(object)
		if err != nil {
			return nil
		}
		img = imaging.Resize(img, 0, int(height), imaging.Lanczos)
	} else if height == 0 && width != 0 {
		img, err = imaging.Decode(object)
		if err != nil {
			return nil
		}
		img = imaging.Resize(img, int(width), 0, imaging.Lanczos)
	} else {
		if _, err := io.Copy(ctx.Response().Writer, object); err != nil {
			return err
		}
		return nil
	}

	if quality <= 0 || quality > 100 {
		return imaging.Encode(ctx.Response().Writer, img, imaging.JPEG, imaging.JPEGQuality(100))
	}
	return imaging.Encode(ctx.Response().Writer, img, imaging.JPEG, imaging.JPEGQuality(int(quality)))
}
