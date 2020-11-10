package api

import (
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/data"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/log"
	"github.com/shopicano/shopicano-backend/middlewares"
	"github.com/shopicano/shopicano-backend/utils"
	"github.com/shopicano/shopicano-backend/validators"
	"net/http"
	"strconv"
	"time"
)

func RegisterUserRoutes(publicEndpoints, platformEndpoints *echo.Group) {
	usersPublicPath := publicEndpoints.Group("/users")
	usersPlatformPath := platformEndpoints.Group("/users")

	func(g echo.Group) {
		g.Use(middlewares.JWTAuth())
		g.PUT("/", update)
		g.GET("/", get)
	}(*usersPublicPath)

	func(g echo.Group) {
		g.Use(middlewares.IsPlatformManager)
		g.PATCH("/:user_id/status/", updateStatus)
	}(*usersPlatformPath)

	func(g echo.Group) {
		g.Use(middlewares.IsPlatformAdmin)
		g.PATCH("/:user_id/permission/", updatePermission)
	}(*usersPlatformPath)
}

func update(ctx echo.Context) error {
	req, err := validators.ValidateUserUpdate(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.StoreCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	userID := ctx.Get(utils.UserID).(string)

	db := app.DB().Begin()

	uc := data.NewUserRepository()
	u, err := uc.Get(db, userID)
	if err != nil {
		db.Rollback()
		log.Log().Errorln(err)

		resp.Title = "Failed to get user profile"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if req.Name != nil {
		u.Name = *req.Name
	}
	if req.ProfilePicture != nil {
		u.ProfilePicture = req.ProfilePicture
	}
	if req.Phone != nil {
		u.Phone = req.Phone
	}

	if req.NewPassword != nil {
		if req.CurrentPassword == nil {
			db.Rollback()

			resp.Title = "Current password is required"
			resp.Status = http.StatusBadRequest
			resp.Code = errors.InvalidRequest
			return resp.ServerJSON(ctx)
		}

		if len(*req.NewPassword) < 8 {
			db.Rollback()

			resp.Title = "New password must be at least 8 characters long"
			resp.Status = http.StatusBadRequest
			resp.Code = errors.InvalidRequest
			return resp.ServerJSON(ctx)
		}

		if req.NewPasswordAgain == nil || *req.NewPassword != *req.NewPasswordAgain {
			db.Rollback()

			resp.Title = "New password and new password again mismatched"
			resp.Status = http.StatusBadRequest
			resp.Code = errors.InvalidRequest
			return resp.ServerJSON(ctx)
		}

		if err := utils.CheckPassword(u.Password, *req.CurrentPassword); err != nil {
			db.Rollback()
			log.Log().Errorln(err)

			resp.Title = "Current password mismatched"
			resp.Status = http.StatusForbidden
			resp.Code = errors.UnauthorizedRequest
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		pass, err := utils.GeneratePassword(*req.NewPassword)
		if err != nil {
			db.Rollback()
			log.Log().Errorln(err)

			resp.Title = "Failed to generate password"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.PasswordEncryptionFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		u.Password = pass
	}

	u.UpdatedAt = time.Now().UTC()

	err = uc.Update(db, u)
	if err != nil {
		db.Rollback()
		log.Log().Errorln(err)

		resp.Title = "Failed to update profile"
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

	resp.Data = map[string]interface{}{
		"id":              u.ID,
		"name":            u.Name,
		"email":           u.Email,
		"status":          u.Status,
		"phone":           u.Phone,
		"profile_picture": u.ProfilePicture,
		"created_at":      u.CreatedAt,
		"updated_at":      u.UpdatedAt,
		"permission":      ctx.Get(utils.UserPermission),
	}

	resp.Status = http.StatusOK
	return resp.ServerJSON(ctx)
}

func updateStatus(ctx echo.Context) error {
	req, err := validators.ValidateUserUpdateStatus(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.StoreCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	userID := ctx.Param("user_id")

	db := app.DB().Begin()

	uc := data.NewUserRepository()
	u, err := uc.Get(db, userID)
	if err != nil {
		db.Rollback()
		log.Log().Errorln(err)

		if errors.IsRecordNotFoundError(err) {
			resp.Title = "User not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.UserNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Failed to get user profile"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	u.Status = req.NewStatus
	u.UpdatedAt = time.Now().UTC()

	err = uc.Update(db, u)
	if err != nil {
		db.Rollback()
		log.Log().Errorln(err)

		resp.Title = "Failed to update profile"
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

	resp.Data = map[string]interface{}{
		"id":              u.ID,
		"name":            u.Name,
		"email":           u.Email,
		"status":          u.Status,
		"phone":           u.Phone,
		"profile_picture": u.ProfilePicture,
		"created_at":      u.CreatedAt,
		"updated_at":      u.UpdatedAt,
		"permission":      ctx.Get(utils.UserPermission),
	}

	resp.Status = http.StatusOK
	return resp.ServerJSON(ctx)
}

func updatePermission(ctx echo.Context) error {
	req, err := validators.ValidateUserUpdatePermission(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.StoreCreationDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	userID := ctx.Param("user_id")

	db := app.DB().Begin()

	uc := data.NewUserRepository()
	u, err := uc.Get(db, userID)
	if err != nil {
		db.Rollback()
		log.Log().Errorln(err)

		if errors.IsRecordNotFoundError(err) {
			resp.Title = "User not found"
			resp.Status = http.StatusNotFound
			resp.Code = errors.UserNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Failed to get user profile"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	u.PermissionID = req.NewPermissionID
	u.UpdatedAt = time.Now().UTC()

	err = uc.Update(db, u)
	if err != nil {
		db.Rollback()
		log.Log().Errorln(err)

		resp.Title = "Failed to update profile"
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

	resp.Data = map[string]interface{}{
		"id":              u.ID,
		"name":            u.Name,
		"email":           u.Email,
		"status":          u.Status,
		"phone":           u.Phone,
		"profile_picture": u.ProfilePicture,
		"created_at":      u.CreatedAt,
		"updated_at":      u.UpdatedAt,
		"permission":      ctx.Get(utils.UserPermission),
	}

	resp.Status = http.StatusOK
	return resp.ServerJSON(ctx)
}

func get(ctx echo.Context) error {
	userID := ctx.Get(utils.UserID).(string)

	resp := core.Response{}

	db := app.DB()

	uc := data.NewUserRepository()
	u, err := uc.Get(db, userID)

	if err != nil {
		log.Log().Errorln(err)

		resp.Title = "Failed to get user profile"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Data = map[string]interface{}{
		"id":              u.ID,
		"name":            u.Name,
		"email":           u.Email,
		"status":          u.Status,
		"phone":           u.Phone,
		"profile_picture": u.ProfilePicture,
		"created_at":      u.CreatedAt,
		"updated_at":      u.UpdatedAt,
		"permission":      ctx.Get(utils.UserPermission),
	}

	resp.Status = http.StatusOK
	return resp.ServerJSON(ctx)
}

func listUsers(ctx echo.Context) error {
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

	from := (page * limit) - limit

	db := app.DB()

	uc := data.NewUserRepository()

	resp := core.Response{}

	var r interface{}

	if query == "" {
		r, err = uc.List(db, int(from), int(limit))
	} else {
		r, err = uc.Search(db, query, int(from), int(limit))
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
