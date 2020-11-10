package data

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/log"
	"github.com/shopicano/shopicano-backend/models"
	"time"
)

type OrderRepositoryImpl struct {
}

var orderRepository OrderRepository

func NewOrderRepository() OrderRepository {
	if orderRepository == nil {
		orderRepository = &OrderRepositoryImpl{}
	}
	return orderRepository
}

func (os *OrderRepositoryImpl) Create(db *gorm.DB, o *models.Order) error {
	if err := db.Table(o.TableName()).Create(o).Error; err != nil {
		return err
	}
	return nil
}

func (os *OrderRepositoryImpl) CreateLog(db *gorm.DB, ol *models.OrderLog) error {
	if err := db.Table(ol.TableName()).Create(ol).Error; err != nil {
		return err
	}
	return nil
}

func (os *OrderRepositoryImpl) UpdatePaymentInfo(db *gorm.DB, o *models.OrderDetailsView) error {
	order := models.Order{}
	if err := db.Table(order.TableName()).
		Where("id = ?", o.ID).
		Select("nonce, transaction_id, payment_status").
		Updates(map[string]interface{}{
			"nonce":          o.Nonce,
			"transaction_id": o.TransactionID,
			"payment_status": o.PaymentStatus,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (os *OrderRepositoryImpl) UpdateStatus(db *gorm.DB, o *models.Order) error {
	order := models.Order{}
	if err := db.Table(order.TableName()).
		Where("id = ?", o.ID).
		Select("status").
		Updates(map[string]interface{}{
			"status": o.Status,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (os *OrderRepositoryImpl) UpdatePaymentStatus(db *gorm.DB, o *models.Order) error {
	order := models.Order{}
	if err := db.Table(order.TableName()).
		Where("id = ?", o.ID).
		Select("payment_status").
		Updates(map[string]interface{}{
			"payment_status": o.PaymentStatus,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (os *OrderRepositoryImpl) AddOrderedItem(db *gorm.DB, oi *models.OrderedItem) error {
	if err := db.Table(oi.TableName()).Create(oi).Error; err != nil {
		return err
	}
	return nil
}

func (os *OrderRepositoryImpl) AddOrderedItemAttribute(db *gorm.DB, attr *models.OrderedItemAttribute) error {
	if err := db.Table(attr.TableName()).Create(attr).Error; err != nil {
		return err
	}
	return nil
}

func (os *OrderRepositoryImpl) GetOrderedItemAttribute(db *gorm.DB, orderedItemID string) (*models.OrderedItemAttribute, error) {
	oi := models.OrderedItemAttribute{}
	if err := db.Table(oi.TableName()).Find(&oi, "ordered_item_id = ?", orderedItemID).Error; err != nil {
		return nil, err
	}
	return &oi, nil
}

func (os *OrderRepositoryImpl) GetOrderedItem(db *gorm.DB, orderID, productID string) (*models.OrderedItem, error) {
	oi := models.OrderedItem{}
	if err := db.Table(oi.TableName()).Find(&oi, "order_id = ? AND product_id = ?", orderID, productID).Error; err != nil {
		return nil, err
	}
	return &oi, nil
}

func (os *OrderRepositoryImpl) List(db *gorm.DB, userID string, offset, limit int) ([]models.OrderDetailsViewExternal, error) {
	order := models.OrderDetailsViewExternal{}
	var orders []models.OrderDetailsViewExternal

	if err := db.Table(order.TableName()).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&orders, "user_id = ?", userID).Error; err != nil {
		log.Log().Errorln(err)
		return nil, err
	}

	for i, v := range orders {
		oiv := models.OrderedItemViewExternal{}

		var items []models.OrderedItemViewExternal
		if err := db.Table(oiv.TableName()).Find(&items, "order_id = ?", v.ID).Error; err != nil {
			log.Log().Errorln(err)
			return nil, err
		}

		for i := range items {
			pu := NewProductRepository()
			images, err := pu.GetImages(db, items[i].ProductID)
			if err != nil {
				return nil, err
			}
			items[i].AdditionalImages = images
		}

		if len(items) == 0 {
			items = []models.OrderedItemViewExternal{}
		}

		orders[i].Items = items
	}

	if len(orders) == 0 {
		orders = []models.OrderDetailsViewExternal{}
	}
	return orders, nil
}

func (os *OrderRepositoryImpl) ListAsStoreStuff(db *gorm.DB, storeID string, offset, limit int) ([]models.OrderDetailsViewExternal, error) {
	order := models.OrderDetailsViewExternal{}
	var orders []models.OrderDetailsViewExternal

	if err := db.Model(&order).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&orders, "store_id = ?", storeID).Error; err != nil {
		log.Log().Errorln(err)
		return nil, err
	}

	for i, v := range orders {
		oiv := models.OrderedItemViewExternal{}

		var items []models.OrderedItemViewExternal
		if err := db.Model(oiv).Find(&items, "order_id = ?", v.ID).Error; err != nil {
			log.Log().Errorln(err)
			return nil, err
		}

		for i := range items {
			pu := NewProductRepository()
			images, err := pu.GetImages(db, items[i].ProductID)
			if err != nil {
				return nil, err
			}
			items[i].AdditionalImages = images
		}

		if len(items) == 0 {
			items = []models.OrderedItemViewExternal{}
		}

		orders[i].Items = items
	}

	if len(orders) == 0 {
		orders = []models.OrderDetailsViewExternal{}
	}
	return orders, nil
}

func (os *OrderRepositoryImpl) Search(db *gorm.DB, query, userID string, offset, limit int) ([]models.OrderDetailsView, error) {
	order := models.OrderDetailsView{}
	var orders []models.OrderDetailsView

	if err := db.Model(&order).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&orders, "user_id = ? AND hash = ?", userID, query).Error; err != nil {
		log.Log().Errorln(err)
		return nil, err
	}

	for i, v := range orders {
		oiv := models.OrderedItemView{}

		var items []models.OrderedItemView
		if err := db.Model(oiv).Find(&items, "order_id = ?", v.ID).Error; err != nil {
			log.Log().Errorln(err)
			return nil, err
		}

		for i := range items {
			pu := NewProductRepository()
			images, err := pu.GetImages(db, items[i].ProductID)
			if err != nil {
				return nil, err
			}
			items[i].AdditionalImages = images
		}

		if len(items) == 0 {
			items = []models.OrderedItemView{}
		}

		orders[i].Items = items
	}

	if len(orders) == 0 {
		orders = []models.OrderDetailsView{}
	}
	return orders, nil
}

func (os *OrderRepositoryImpl) SearchAsStoreStuff(db *gorm.DB, query, storeID string, offset, limit int) ([]models.OrderDetailsView, error) {
	order := models.OrderDetailsView{}
	var orders []models.OrderDetailsView

	if err := db.Model(&order).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&orders, "store_id = ? AND hash = ?", storeID, query).Error; err != nil {
		log.Log().Errorln(err)
		return nil, err
	}

	for i, v := range orders {
		oiv := models.OrderedItemView{}

		var items []models.OrderedItemView
		if err := db.Model(oiv).Find(&items, "order_id = ?", v.ID).Error; err != nil {
			log.Log().Errorln(err)
			return nil, err
		}

		for i := range items {
			pu := NewProductRepository()
			images, err := pu.GetImages(db, items[i].ProductID)
			if err != nil {
				return nil, err
			}
			items[i].AdditionalImages = images
		}

		if len(items) == 0 {
			items = []models.OrderedItemView{}
		}

		orders[i].Items = items
	}

	if len(orders) == 0 {
		orders = []models.OrderDetailsView{}
	}
	return orders, nil
}

func (os *OrderRepositoryImpl) GetDetails(db *gorm.DB, orderID string) (*models.OrderDetailsView, error) {
	order := models.OrderDetailsView{}
	if err := db.Model(&order).First(&order, "id = ?", orderID).Error; err != nil {
		log.Log().Errorln(err)
		return nil, err
	}

	oiv := models.OrderedItemView{}

	var items []models.OrderedItemView
	if err := db.Model(oiv).Find(&items, "order_id = ?", orderID).Error; err != nil {
		log.Log().Errorln(err)
		return nil, err
	}

	for i := range items {
		pu := NewProductRepository()
		images, err := pu.GetImages(db, items[i].ProductID)
		if err != nil {
			return nil, err
		}
		items[i].AdditionalImages = images
	}

	order.Items = items
	return &order, nil
}

func (os *OrderRepositoryImpl) GetDetailsAsStoreStuff(db *gorm.DB, storeID, orderID string) (*models.OrderDetailsView, error) {
	order := models.OrderDetailsView{}
	if err := db.Model(&order).First(&order, "id = ? AND store_id = ?", orderID, storeID).Error; err != nil {
		log.Log().Errorln(err)
		return nil, err
	}

	oiv := models.OrderedItemView{}

	var items []models.OrderedItemView
	if err := db.Model(oiv).Find(&items, "order_id = ?", orderID).Error; err != nil {
		log.Log().Errorln(err)
		return nil, err
	}

	for i, it := range items {
		attribute := models.OrderedItemAttribute{}
		var attributes []models.OrderedItemAttribute
		if err := db.Table(attribute.TableName()).Find(&attributes, "ordered_item_id = ?", it.ID).Error; err != nil {
			if errors.IsRecordNotFoundError(err) {
				continue
			}
			return nil, err
		}

		for _, attr := range attributes {
			items[i].Attributes = append(items[i].Attributes, models.OrderItemAttributeKV{
				Key:   attr.AttributeKey,
				Value: attr.AttributeValue,
			})
		}
	}

	for i := range items {
		pu := NewProductRepository()
		images, err := pu.GetImages(db, items[i].ProductID)
		if err != nil {
			return nil, err
		}
		items[i].AdditionalImages = images
	}

	order.Items = items
	return &order, nil
}

func (os *OrderRepositoryImpl) GetAsStoreStuff(db *gorm.DB, storeID, orderID string) (*models.Order, error) {
	order := models.Order{}
	if err := db.Model(&order).First(&order, "id = ? AND store_id = ?", orderID, storeID).Error; err != nil {
		log.Log().Errorln(err)
		return nil, err
	}
	return &order, nil
}

func (os *OrderRepositoryImpl) GetDetailsAsUser(db *gorm.DB, userID, orderID string) (*models.OrderDetailsViewExternal, error) {
	order := models.OrderDetailsViewExternal{}
	if err := db.Model(&order).First(&order, "id = ? AND user_id = ?", orderID, userID).Error; err != nil {
		log.Log().Errorln(err)
		return nil, err
	}

	oiv := models.OrderedItemView{}

	var items []models.OrderedItemViewExternal
	if err := db.Model(oiv).Find(&items, "order_id = ?", orderID).Error; err != nil {
		log.Log().Errorln(err)
		return nil, err
	}

	for i, it := range items {
		attribute := models.OrderedItemAttribute{}
		var attributes []models.OrderedItemAttribute
		if err := db.Table(attribute.TableName()).Find(&attributes, "ordered_item_id = ?", it.ID).Error; err != nil {
			if errors.IsRecordNotFoundError(err) {
				continue
			}
			return nil, err
		}

		for _, attr := range attributes {
			items[i].Attributes = append(items[i].Attributes, models.OrderItemAttributeKV{
				Key:   attr.AttributeKey,
				Value: attr.AttributeValue,
			})
		}
	}

	for i := range items {
		pu := NewProductRepository()
		images, err := pu.GetImages(db, items[i].ProductID)
		if err != nil {
			return nil, err
		}
		items[i].AdditionalImages = images
	}

	order.Items = items
	return &order, nil
}

func (os *OrderRepositoryImpl) StoreSummary(db *gorm.DB, storeID string) (*models.Summary, error) {
	o := models.Order{}
	oi := models.OrderedItem{}

	sum := models.Summary{}

	if err := db.Table(fmt.Sprintf("%s AS o", o.TableName())).
		Select("COUNT(o.id) AS total_orders, SUM(oi.price * oi.quantity) AS earnings, SUM(oi.product_cost * oi.quantity) AS expenses,"+
			"SUM(oi.price * oi.quantity) - SUM(oi.product_cost * oi.quantity) AS profits, SUM(o.discounted_amount) AS discounts,"+
			"COUNT(DISTINCT (o.user_id)) AS customers").
		Joins(fmt.Sprintf("JOIN %s AS oi ON o.id = oi.order_id", oi.TableName())).
		Where("o.store_id = ? AND o.status = ? AND o.payment_status = ?", storeID, models.OrderDelivered, models.PaymentCompleted).
		Find(&sum).Error; err != nil {
		return nil, err
	}
	return &sum, nil
}

func (os *OrderRepositoryImpl) StoreSummaryAny(db *gorm.DB, storeID string) (*models.Summary, error) {
	o := models.Order{}
	oi := models.OrderedItem{}

	sum := models.Summary{}

	if err := db.Table(fmt.Sprintf("%s AS o", o.TableName())).
		Select("COUNT(o.id) AS total_orders, SUM(oi.price * oi.quantity) AS earnings, SUM(oi.product_cost * oi.quantity) AS expenses,"+
			"SUM(oi.price * oi.quantity) - SUM(oi.product_cost * oi.quantity) AS profits, SUM(o.discounted_amount) AS discounts,"+
			"COUNT(DISTINCT (o.user_id)) AS customers").
		Joins(fmt.Sprintf("LEFT JOIN %s AS oi ON o.id = oi.order_id", oi.TableName())).
		Where("o.store_id = ?", storeID).
		Find(&sum).Error; err != nil {
		return nil, err
	}
	return &sum, nil
}

func (os *OrderRepositoryImpl) StoreSummaryByTime(db *gorm.DB, storeID string, from, end time.Time) (*models.Summary, error) {
	o := models.Order{}
	oi := models.OrderedItem{}

	sum := models.Summary{}

	if err := db.Table(fmt.Sprintf("%s AS o", o.TableName())).
		Select("COUNT(o.id) AS total_orders, SUM(oi.price * oi.quantity) AS earnings, SUM(oi.product_cost * oi.quantity) AS expenses,"+
			"SUM(oi.price * oi.quantity) - SUM(oi.product_cost * oi.quantity) AS profits, SUM(o.discounted_amount) AS discounts,"+
			"COUNT(DISTINCT (o.user_id)) AS customers").
		Joins(fmt.Sprintf("JOIN %s AS oi ON o.id = oi.order_id", oi.TableName())).
		Where("o.store_id = ? AND o.created_at >= ? AND o.created_At <= ? AND o.status = ? AND o.payment_status = ?",
			storeID, from, end, models.OrderDelivered, models.PaymentCompleted).
		Find(&sum).Error; err != nil {
		return nil, err
	}
	return &sum, nil
}

func (os *OrderRepositoryImpl) CountByStatus(db *gorm.DB, storeID string, from, end time.Time) ([]models.StatusReport, error) {
	o := models.Order{}

	var stats []models.StatusReport

	if err := db.Table(fmt.Sprintf("%s AS o", o.TableName())).
		Select("o.status AS key, COUNT(o.status) AS value").
		Group("o.status").
		Where("o.store_id = ? AND created_at >= ? AND created_at <= ?", storeID, from, end).
		Scan(&stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

func (os *OrderRepositoryImpl) EarningsByStatus(db *gorm.DB, storeID string, from, end time.Time) ([]models.StatusReport, error) {
	o := models.Order{}

	var stats []models.StatusReport

	if err := db.Table(fmt.Sprintf("%s AS o", o.TableName())).
		Select("o.payment_status AS key, SUM(o.grand_total) AS value").
		Group("o.payment_status").
		Where("o.store_id = ? AND created_at >= ? AND created_at <= ?", storeID, from, end).
		Scan(&stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

func (os *OrderRepositoryImpl) CreateReview(db *gorm.DB, r *models.Review) error {
	if err := db.Table(r.TableName()).Create(r).Error; err != nil {
		return err
	}
	return nil
}
