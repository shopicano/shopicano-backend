package tasks

import (
	"fmt"
	"git.cloudbro.net/michaelfigg/yallawebsites/app"
	"git.cloudbro.net/michaelfigg/yallawebsites/config"
	"git.cloudbro.net/michaelfigg/yallawebsites/data"
	"git.cloudbro.net/michaelfigg/yallawebsites/log"
	"git.cloudbro.net/michaelfigg/yallawebsites/services"
	"git.cloudbro.net/michaelfigg/yallawebsites/templates"
	"git.cloudbro.net/michaelfigg/yallawebsites/utils"
	"github.com/RichardKnop/machinery/v1/tasks"
	"time"
)

const (
	SendSignUpVerificationEmailTaskName = "send_sign_up_verification_email"
)

func SendSignUpVerificationEmailFn(userID string) error {
	db := app.DB().Begin()

	adminDao := data.NewMarketplaceRepository()
	settings, err := adminDao.GetSettings(db)
	if err != nil {
		db.Rollback()

		log.Log().Errorln(err)
		return tasks.NewErrRetryTaskLater(err.Error(), time.Second*30)
	}

	userDao := data.NewUserRepository()
	u, err := userDao.Get(db, userID)
	if err != nil {
		db.Rollback()

		log.Log().Errorln(err)
		return tasks.NewErrRetryTaskLater(err.Error(), time.Second*30)
	}

	t := utils.NewToken()
	u.VerificationToken = &t

	if err := userDao.Update(db, u); err != nil {
		db.Rollback()

		log.Log().Errorln(err)
		return tasks.NewErrRetryTaskLater(err.Error(), time.Second*30)
	}

	verificationUrl := fmt.Sprintf("%s%s&email=%s&token=%s",
		config.App().FrontStoreUrl, config.PathMappingCfg()["after_account_verification"], u.Email, *u.VerificationToken)

	params := map[string]interface{}{
		"platformName":    settings.Name,
		"platformWebsite": settings.Website,
		"verificationUrl": verificationUrl,
		"userName":        u.Name,
		"assetsUrl":       fmt.Sprintf("%s/assets/", settings.Website),
	}

	body, err := templates.GenerateActivateAccountEmailHTML(params)
	if err != nil {
		db.Rollback()

		log.Log().Errorln(err)
		return tasks.NewErrRetryTaskLater(err.Error(), time.Second*30)
	}

	if err := services.SendSignUpVerificationEmail(u.Email, "Please verify your account", body); err != nil {
		db.Rollback()

		log.Log().Errorln(err)
		return tasks.NewErrRetryTaskLater(err.Error(), time.Second*30)
	}

	if err := db.Commit().Error; err != nil {
		return tasks.NewErrRetryTaskLater(err.Error(), time.Second*30)
	}

	return nil
}
