package server

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/shopicano/shopicano-backend/api"
	"github.com/shopicano/shopicano-backend/middlewares"
	"net/http"
	"strings"
)

var router = echo.New()

// GetRouter returns the api router
func GetRouter() http.Handler {
	router.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 6,
		Skipper: func(ctx echo.Context) bool {
			return strings.Contains(ctx.Path(), "/fs/") || strings.Contains(ctx.Path(), "/download/")
		},
	}))

	router.Pre(middleware.AddTrailingSlash())
	router.Use(middleware.Logger())
	router.Use(middleware.Recover())
	router.Use(EchoMonitoring())

	router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"*"},
		AllowMethods: []string{"*"},
	}))

	router.GET("/", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "OK")
	})

	registerV1Routes()
	return router
}

func registerV1Routes() {
	v1 := router.Group("/v1")
	publicEndpoints := v1
	platformEndpoints := v1.Group("/marketplace")

	platformEndpoints.Use(middlewares.JWTAuth())

	api.RegisterLegacyRoutes(publicEndpoints, platformEndpoints)
	api.RegisterPlatformRoutes(publicEndpoints, platformEndpoints)
	api.RegisterUserRoutes(publicEndpoints, platformEndpoints)
	api.RegisterStoreRoutes(publicEndpoints, platformEndpoints)
	api.RegisterProductRoutes(publicEndpoints, platformEndpoints)
	api.RegisterCategoryRoutes(publicEndpoints, platformEndpoints)
	api.RegisterCollectionRoutes(publicEndpoints, platformEndpoints)
	api.RegisterFSRoutes(publicEndpoints, platformEndpoints)
	api.RegisterAddressRoutes(publicEndpoints, platformEndpoints)
	api.RegisterOrderRoutes(publicEndpoints, platformEndpoints)
	api.RegisterPaymentRoutes(publicEndpoints, platformEndpoints)
	api.RegisterCustomerRoutes(publicEndpoints, platformEndpoints)
	api.RegisterStatsRoutes(publicEndpoints, platformEndpoints)
	api.RegisterCouponRoutes(publicEndpoints, platformEndpoints)
	api.RegisterLocationRoutes(publicEndpoints, platformEndpoints)
}
