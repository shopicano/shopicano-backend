package api

import (
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/data"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/log"
	"github.com/shopicano/shopicano-backend/middlewares"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/queue"
	"github.com/shopicano/shopicano-backend/utils"
	"github.com/shopicano/shopicano-backend/validators"
	"github.com/shopicano/shopicano-backend/values"
	"net/http"
	"time"
)

func RegisterLegacyRoutes(publicEndpoints, platformEndpoints *echo.Group) {
	publicEndpoints.POST("/login/", login)
	publicEndpoints.GET("/logout/", logout)
	publicEndpoints.GET("/refresh-token/", refreshToken)
	publicEndpoints.GET("/email-verification/", emailVerification)
	publicEndpoints.GET("/reset-password/", resetPasswordRequest)
	publicEndpoints.POST("/reset-password/", resetPasswordUpdate)

	func(g echo.Group) {
		g.Use(middlewares.IsSignUpEnabled)
		g.POST("/register/", register)
	}(*publicEndpoints)
}

func login(ctx echo.Context) error {
	pld, err := validators.ValidateLogin(ctx)

	resp := core.Response{}

	if err != nil {
		log.Log().Errorln(err)
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.UserLoginDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB().Begin()

	uc := data.NewUserRepository()
	u, err := uc.Login(db, pld.Email)

	if err != nil {
		db.Rollback()
		log.Log().Errorln(err)

		if errors.IsRecordNotFoundError(err) {
			resp.Title = "User not registered"
			resp.Status = http.StatusNotFound
			resp.Code = errors.UserNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if err := utils.CheckPassword(u.Password, pld.Password); err != nil {
		db.Rollback()

		resp.Title = "Invalid login credentials"
		resp.Status = http.StatusUnauthorized
		resp.Code = errors.LoginCredentialsInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if u.Status != models.UserActive {
		db.Rollback()

		resp.Title = "Please activate your account by verifying email"
		resp.Status = http.StatusForbidden
		resp.Code = errors.UserNotActive
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	_, permission, err := uc.GetPermissionByUserID(db, u.ID)
	if err != nil {
		db.Rollback()

		log.Log().Errorln(err)

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	scope := utils.FrontStore
	switch pld.Scope {
	case utils.Platform:
		scope = utils.Platform
	case utils.BackStore:
		scope = utils.BackStore
	}

	signedToken, err := utils.BuildJWTToken(u.ID, scope)
	if err != nil {
		db.Rollback()
		log.Log().Errorln(err)

		resp.Title = "Failed to sign auth token"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.UserLoginFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	s := models.Session{
		ID:           utils.NewUUID(),
		UserID:       u.ID,
		AccessToken:  signedToken,
		RefreshToken: utils.NewToken(),
		CreatedAt:    time.Now().UTC(),
		ExpireOn:     time.Now().Add(time.Hour * 48).Unix(),
	}

	if err := uc.CreateSession(db, &s); err != nil {
		db.Rollback()

		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	result := map[string]interface{}{
		"access_token":  s.AccessToken,
		"refresh_token": s.RefreshToken,
		"expire_on":     s.ExpireOn,
		"permission":    permission,
	}

	sc := data.NewStoreRepository()
	profile, err := sc.GetStoreUserProfile(db, u.ID)
	if err != nil {
		if !errors.IsRecordNotFoundError(err) {
			db.Rollback()

			resp.Title = "Database query failed"
			resp.Status = http.StatusInternalServerError
			resp.Code = errors.DatabaseQueryFailed
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		result["has_store"] = false
	} else {
		result["has_store"] = true
		result["staff"] = profile
	}

	if err := db.Commit().Error; err != nil {
		resp.Title = "Database query failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Status = http.StatusOK
	resp.Data = result
	return resp.ServerJSON(ctx)
}

func logout(ctx echo.Context) error {
	token, err := utils.ParseBearerToken(ctx)

	resp := core.Response{}

	if err != nil {
		log.Log().Errorln(err)

		resp.Title = "Invalid data"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.BearerTokenNotFound
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()

	uc := data.NewUserRepository()
	if err := uc.Logout(db, token); err != nil {
		log.Log().Errorln(err)

		resp.Title = "Failed to logout"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Title = "Logout successful"
	resp.Status = http.StatusNoContent
	return resp.ServerJSON(ctx)
}

func refreshToken(ctx echo.Context) error {
	token, err := utils.ParseBearerToken(ctx)

	resp := core.Response{}

	if err != nil {
		log.Log().Errorln(err)

		resp.Title = "Invalid data"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.BearerTokenNotFound
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB()

	uc := data.NewUserRepository()
	s, err := uc.RefreshToken(db, token)
	if err != nil {
		log.Log().Errorln(err)

		ok := errors.IsRecordNotFoundError(err)
		if ok {
			resp.Title = "User already register"
			resp.Status = http.StatusBadRequest
			resp.Code = errors.BearerTokenNotFound
			resp.Errors = errors.NewError("Invalid refresh token")
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Failed to generate refresh token"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Title = "Refresh token generation successful"
	resp.Status = http.StatusOK
	resp.Data = s
	return resp.ServerJSON(ctx)
}

func register(ctx echo.Context) error {
	u, err := validators.ValidateRegister(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.UserSignUpDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	db := app.DB().Begin()

	uc := data.NewUserRepository()

	u.Password, _ = utils.GeneratePassword(u.Password)
	u.PermissionID = values.UserGroupID

	if err := uc.Register(db, u); err != nil {
		db.Rollback()

		msg, ok := errors.IsDuplicateKeyError(err)
		if ok {
			resp.Title = "User already register"
			resp.Status = http.StatusConflict
			resp.Code = errors.UserAlreadyExists
			resp.Errors = errors.NewError(msg)
			return resp.ServerJSON(ctx)
		}

		resp.Title = "User registration failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if err := queue.SendSignUpVerificationEmail(u.ID); err != nil {
		db.Rollback()

		resp.Title = "User registration failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.FailedToEnqueueTask
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if err := db.Commit().Error; err != nil {
		resp.Title = "User registration failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Title = "User registration successful"
	resp.Data = u
	resp.Status = http.StatusCreated
	return resp.ServerJSON(ctx)
}

func emailVerification(ctx echo.Context) error {
	resp := core.Response{}

	email := ctx.QueryParam("email")
	token := ctx.QueryParam("token")

	db := app.DB().Begin()

	uc := data.NewUserRepository()

	u, err := uc.GetByEmail(db, email)
	if err != nil {
		ok := errors.IsRecordNotFoundError(err)
		if ok {
			resp.Title = "User not registered"
			resp.Status = http.StatusNotFound
			resp.Code = errors.UserNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Email verification failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if u.VerificationToken == nil || *u.VerificationToken != token {
		resp.Title = "Email verification failed"
		resp.Status = http.StatusForbidden
		resp.Code = errors.VerificationTokenIsInvalid
		return resp.ServerJSON(ctx)
	}

	if u.IsEmailVerified {
		resp.Title = "Email already verified"
		resp.Status = http.StatusOK
		return resp.ServerJSON(ctx)
	}

	u.VerificationToken = nil
	u.Status = models.UserActive
	u.IsEmailVerified = true
	if err := uc.Update(db, u); err != nil {
		resp.Title = "Email verification failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		return resp.ServerJSON(ctx)
	}

	if err := db.Commit().Error; err != nil {
		resp.Title = "Email verification failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		return resp.ServerJSON(ctx)
	}

	resp.Title = "Email has been verified"
	resp.Status = http.StatusOK
	return resp.ServerJSON(ctx)
}

func resetPasswordRequest(ctx echo.Context) error {
	resp := core.Response{}

	email := ctx.QueryParam("email")

	db := app.DB()

	uc := data.NewUserRepository()

	u, err := uc.GetByEmail(db, email)
	if err != nil {
		ok := errors.IsRecordNotFoundError(err)
		if ok {
			resp.Title = "Email not registered"
			resp.Status = http.StatusNotFound
			resp.Code = errors.UserNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Reset password failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if err := queue.SendPasswordResetRequestEmail(u.ID); err != nil {
		resp.Title = "Task queueing failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.FailedToEnqueueTask
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	resp.Title = "Reset password email sent"
	resp.Status = http.StatusOK
	return resp.ServerJSON(ctx)
}

func resetPasswordUpdate(ctx echo.Context) error {
	pld, err := validators.ValidateResetPassword(ctx)

	resp := core.Response{}

	if err != nil {
		resp.Title = "Invalid data"
		resp.Status = http.StatusUnprocessableEntity
		resp.Code = errors.ResetPasswordDataInvalid
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if pld.NewPassword != pld.NewPasswordAgain {
		resp.Title = "Password mismatch"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.ResetPasswordDataInvalid
		return resp.ServerJSON(ctx)
	}

	db := app.DB().Begin()

	uc := data.NewUserRepository()

	u, err := uc.GetByEmail(db, pld.Email)
	if err != nil {
		db.Rollback()

		ok := errors.IsRecordNotFoundError(err)
		if ok {
			resp.Title = "Email not registered"
			resp.Status = http.StatusNotFound
			resp.Code = errors.UserNotFound
			resp.Errors = err
			return resp.ServerJSON(ctx)
		}

		resp.Title = "Email verification failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if u.ResetPasswordToken == nil || pld.Token != *u.ResetPasswordToken {
		db.Rollback()

		resp.Title = "Invalid token"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.ResetPasswordDataInvalid
		return resp.ServerJSON(ctx)
	}

	now := time.Now().UTC()
	if now.After(*u.ResetPasswordTokenGeneratedAt) {
		db.Rollback()

		resp.Title = "Token expired"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.ResetPasswordDataInvalid
		return resp.ServerJSON(ctx)
	}

	pass, err := utils.GeneratePassword(pld.NewPassword)
	if err != nil {
		db.Rollback()

		resp.Title = "Failed to generate password hash"
		resp.Status = http.StatusBadRequest
		resp.Code = errors.PasswordEncryptionFailed
		return resp.ServerJSON(ctx)
	}

	u.ResetPasswordTokenGeneratedAt = nil
	u.ResetPasswordToken = nil
	u.Password = pass

	if err := uc.Update(db, u); err != nil {
		db.Rollback()

		resp.Title = "Password reset failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		return resp.ServerJSON(ctx)
	}

	if err := queue.SendPasswordResetConfirmationEmail(u.ID); err != nil {
		resp.Title = "Task queueing failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.FailedToEnqueueTask
		resp.Errors = err
		return resp.ServerJSON(ctx)
	}

	if err := db.Commit().Error; err != nil {
		resp.Title = "Password reset failed"
		resp.Status = http.StatusInternalServerError
		resp.Code = errors.DatabaseQueryFailed
		return resp.ServerJSON(ctx)
	}

	resp.Title = "Password has been changed"
	resp.Status = http.StatusOK
	return resp.ServerJSON(ctx)
}

func redirect(ctx echo.Context, url string) error {
	return ctx.Redirect(http.StatusPermanentRedirect, url)
}
