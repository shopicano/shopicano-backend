package data

import (
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/models"
	"time"
)

type OrderRepository interface {
	Create(db *gorm.DB, o *models.Order) error
	CreateLog(db *gorm.DB, ol *models.OrderLog) error
	AddOrderedItem(db *gorm.DB, item *models.OrderedItem) error
	AddOrderedItemAttribute(db *gorm.DB, attr *models.OrderedItemAttribute) error
	GetOrderedItem(db *gorm.DB, orderID, productID string) (*models.OrderedItem, error)
	GetDetailsAsUser(db *gorm.DB, userID, orderID string) (*models.OrderDetailsViewExternal, error)
	GetDetailsAsStoreStuff(db *gorm.DB, storeID, orderID string) (*models.OrderDetailsView, error)
	GetAsStoreStuff(db *gorm.DB, storeID, orderID string) (*models.Order, error)
	GetDetails(db *gorm.DB, orderID string) (*models.OrderDetailsView, error)
	UpdatePaymentInfo(db *gorm.DB, o *models.OrderDetailsView) error
	UpdateStatus(db *gorm.DB, o *models.Order) error
	UpdatePaymentStatus(db *gorm.DB, o *models.Order) error
	List(db *gorm.DB, userID string, offset, limit int) ([]models.OrderDetailsViewExternal, error)
	ListAsStoreStuff(db *gorm.DB, storeID string, offset, limit int) ([]models.OrderDetailsViewExternal, error)
	Search(db *gorm.DB, query, userID string, offset, limit int) ([]models.OrderDetailsView, error)
	SearchAsStoreStuff(db *gorm.DB, query, storeID string, offset, limit int) ([]models.OrderDetailsView, error)
	CreateReview(db *gorm.DB, review *models.Review) error

	// Report functionality
	StoreSummary(db *gorm.DB, storeID string) (*models.Summary, error)
	StoreSummaryAny(db *gorm.DB, storeID string) (*models.Summary, error)
	StoreSummaryByTime(db *gorm.DB, storeID string, from, end time.Time) (*models.Summary, error)
	CountByStatus(db *gorm.DB, storeID string, from, end time.Time) ([]models.StatusReport, error)
	EarningsByStatus(db *gorm.DB, storeID string, from, end time.Time) ([]models.StatusReport, error)
}
