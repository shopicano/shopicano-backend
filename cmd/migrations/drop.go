package migration

import (
	"github.com/shopicano/shopicano-backend/app"
	"github.com/shopicano/shopicano-backend/core"
	"github.com/shopicano/shopicano-backend/log"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/spf13/cobra"
)

var MigDropCmd = &cobra.Command{
	Use:   "drop",
	Short: "drop drops database tables",
	Run:   drop,
}

func drop(cmd *cobra.Command, args []string) {
	tx := app.DB().Begin()

	var tables []core.Table
	tables = append(tables, &models.CouponUsage{}, &models.CouponFor{}, &models.Coupon{}, &models.Review{}, &models.OrderedItemAttribute{})
	tables = append(tables, &models.ProductAttribute{}, &models.OrderLog{}, &models.ProductImage{})
	tables = append(tables, &models.OrderedItem{}, &models.Order{})
	tables = append(tables, &models.CollectionOfProduct{}, &models.Product{}, &models.Category{}, &models.Collection{})
	tables = append(tables, &models.ShippingMethod{}, &models.PaymentMethod{}, &models.Settings{})
	tables = append(tables, &models.Staff{}, &models.StorePermission{}, &models.Store{})
	tables = append(tables, &models.Address{}, &models.Session{}, &models.User{}, &models.UserPermission{})
	tables = append(tables, &models.Location{}, &models.Log{})

	for _, t := range tables {
		if err := tx.DropTableIfExists(t).Error; err != nil {
			tx.Rollback()
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Log().Errorln(err)
	}

	log.Log().Infoln("Migration drop completed")
}
