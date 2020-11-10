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
	"net/http"
	"sort"
	"time"
)

func RegisterStatsRoutes(publicEndpoints, platformEndpoints *echo.Group) {
	statsPublicPath := publicEndpoints.Group("/stats")
	statsPlatformPath := platformEndpoints.Group("/stats")

	func(g echo.Group) {
		g.Use(middlewares.HasStore())
		g.Use(middlewares.IsStoreActive())
		g.Use(middlewares.IsStoreManager())
		g.GET("/products/", productStatsAsStoreOwner)
		g.GET("/categories/", categoryStatsAsStoreOwner)
		g.GET("/orders/", orderStats)
	}(*statsPlatformPath)

	func(g echo.Group) {
		g.GET("/products/", productStats)
		g.GET("/categories/", categoryStats)
	}(*statsPublicPath)
}

func productStats(ctx echo.Context) error {
	resp := core.Response{}

	db := app.DB()
	pu := data.NewProductRepository()

	var res interface{}
	var err error

	res, err = pu.Stats(db, 0, 25)
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = res
	return resp.ServerJSON(ctx)
}

func productStatsAsStoreOwner(ctx echo.Context) error {
	resp := core.Response{}

	db := app.DB()
	pu := data.NewProductRepository()

	var res interface{}
	var err error

	res, err = pu.StatsAsStoreStaff(db, utils.GetStoreID(ctx), 0, 25)
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = res
	return resp.ServerJSON(ctx)
}

func categoryStats(ctx echo.Context) error {
	resp := core.Response{}

	db := app.DB()
	cu := data.NewCategoryRepository()

	var res interface{}
	var err error

	res, err = cu.Stats(db, 0, 25)
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = res
	return resp.ServerJSON(ctx)
}

func categoryStatsAsStoreOwner(ctx echo.Context) error {
	resp := core.Response{}

	db := app.DB()
	cu := data.NewCategoryRepository()

	var res interface{}
	var err error

	res, err = cu.StatsAsStoreStuff(db, utils.GetStoreID(ctx), 0, 25)
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = res
	return resp.ServerJSON(ctx)
}

func orderStats(ctx echo.Context) error {
	resp := core.Response{}

	db := app.DB()
	ou := data.NewOrderRepository()

	timeFrames := map[time.Time]time.Time{}

	timeline := ctx.QueryParam("timeline")
	switch timeline {
	case "w":
		now := time.Now()
		prev := now.Add(time.Hour * -24)

		for len(timeFrames) < 7 {
			timeFrames[prev] = now

			now = prev
			prev = now.Add(time.Hour * -24)
		}
	case "m":
		now := time.Now()
		prev := now.Add(time.Hour * 7 * -24)

		for len(timeFrames) < 5 {
			timeFrames[prev] = now

			now = prev
			prev = now.Add(time.Hour * 7 * -24)
		}
	case "y":
		now := time.Now()
		prev := now.Add(time.Hour * 30 * -24)

		for len(timeFrames) < 12 {
			timeFrames[prev] = now

			now = prev
			prev = now.Add(time.Hour * 30 * -24)
		}
	}

	summary, err := ou.StoreSummary(db, utils.GetStoreID(ctx))
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	summaryAny, err := ou.StoreSummaryAny(db, utils.GetStoreID(ctx))
	if err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	var timeWiseSummary []*models.Summary
	var ordersStats []map[string]interface{}
	var earningsStats []map[string]interface{}

	for k, v := range timeFrames {
		sum, err := ou.StoreSummaryByTime(db, utils.GetStoreID(ctx), k, v)
		if err != nil {
			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}
		sum.Time = v.Format(utils.DateFormat)
		timeWiseSummary = append(timeWiseSummary, sum)

		// Orders Calculation
		cStat, err := ou.CountByStatus(db, utils.GetStoreID(ctx), k, v)
		if err != nil {
			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		pending := 0
		confirmed := 0
		shipping := 0
		delivered := 0
		cancelled := 0

		for _, x := range cStat {
			switch x.Key {
			case string(models.OrderPending):
				pending = x.Value
			case string(models.OrderConfirmed):
				confirmed = x.Value
			case string(models.OrderShipping):
				shipping = x.Value
			case string(models.OrderDelivered):
				delivered = x.Value
			case string(models.OrderCancelled):
				cancelled = x.Value
			}
		}

		ordersStats = append(ordersStats, map[string]interface{}{
			"time":      sum.Time,
			"pending":   pending,
			"confirmed": confirmed,
			"shipping":  shipping,
			"delivered": delivered,
			"cancelled": cancelled,
		})

		// Earnings Calculation
		eStat, err := ou.EarningsByStatus(db, utils.GetStoreID(ctx), k, v)
		if err != nil {
			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		pending = 0
		completed := 0
		failed := 0
		reverted := 0

		for _, x := range eStat {
			switch x.Key {
			case string(models.PaymentPending):
				pending = x.Value
			case string(models.PaymentCompleted):
				completed = x.Value
			case string(models.PaymentFailed):
				failed = x.Value
			case string(models.PaymentReverted):
				reverted = x.Value
			}
		}

		earningsStats = append(earningsStats, map[string]interface{}{
			"time":      sum.Time,
			"pending":   pending,
			"completed": completed,
			"failed":    failed,
			"reverted":  reverted,
		})
	}

	sort.Slice(timeWiseSummary, func(i, j int) bool {
		x := timeWiseSummary[i].Time
		y := timeWiseSummary[j].Time

		xStart, _ := time.Parse(utils.DateFormat, x)
		yStart, _ := time.Parse(utils.DateFormat, y)
		return !xStart.After(yStart)
	})

	sort.Slice(ordersStats, func(i, j int) bool {
		x := ordersStats[i]
		y := ordersStats[j]

		xStart, _ := time.Parse(utils.DateFormat, x["time"].(string))
		yStart, _ := time.Parse(utils.DateFormat, y["time"].(string))
		return !xStart.After(yStart)
	})

	sort.Slice(earningsStats, func(i, j int) bool {
		x := earningsStats[i]
		y := earningsStats[j]

		xStart, _ := time.Parse(utils.DateFormat, x["time"].(string))
		yStart, _ := time.Parse(utils.DateFormat, y["time"].(string))
		return !xStart.After(yStart)
	})

	resp.Status = http.StatusOK
	resp.Data = map[string]interface{}{
		"report_completed": summary,
		"report":           summaryAny,
		"reports_by_time":  timeWiseSummary,
		"orders_by_time":   ordersStats,
		"earnings_by_time": earningsStats,
	}
	return resp.ServerJSON(ctx)
}
