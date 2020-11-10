package server

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/log"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/utils"
	"time"
)

func EchoMonitoring() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			req := c.Request()
			res := c.Response()

			if err := next(c); err != nil {
				c.Error(err)
			}

			timeTaken := time.Since(start).Seconds()
			size := res.Size

			label := req.Method + " "
			if c.Path() != "" {
				label += c.Path()
			} else {
				label += req.URL.Path
			}

			first := false
			for k, v := range req.URL.Query() {
				if !first {
					label += "?"
				}
				if first {
					label += "&"
				}

				label += fmt.Sprintf("%s=%s", k, v)
				first = true
			}

			l := models.Log{
				ID:        utils.NewUUID(),
				Label:     label,
				Path:      req.URL.Path,
				Status:    res.Status,
				Size:      size,
				IP:        req.RemoteAddr,
				User:      "", // TODO :
				TimeTaken: timeTaken,
				CreatedAt: time.Now().UTC(),
			}

			db := app.DB()
			if err := db.Table(l.TableName()).Create(&l).Error; err != nil {
				log.Log().Errorln(err)
			}
			return nil
		}
	}
}
