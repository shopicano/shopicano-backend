package data

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/helpers"
	"github.com/shopicano/shopicano-backend/models"
)

type ProductRepositoryImpl struct {
}

var productRepository ProductRepository

func NewProductRepository() ProductRepository {
	if productRepository == nil {
		productRepository = &ProductRepositoryImpl{}
	}

	return productRepository
}

func (pu *ProductRepositoryImpl) Create(db *gorm.DB, p *models.Product) error {
	if err := db.Model(p).Create(p).Error; err != nil {
		return err
	}
	return nil
}

func (pu *ProductRepositoryImpl) Update(db *gorm.DB, p *models.Product) error {
	if err := db.Table(p.TableName()).
		Select("name, description, is_published, category_id, sku, slug, stock, unit, price, product_cost, max_quantity_count, image, is_shippable, is_digital, digital_download_link, updated_at").
		Where("id = ? AND store_id = ?", p.ID, p.StoreID).
		Updates(map[string]interface{}{
			"name":                  p.Name,
			"description":           p.Description,
			"is_published":          p.IsPublished,
			"category_id":           p.CategoryID,
			"sku":                   p.SKU,
			"slug":                  p.Slug,
			"stock":                 p.Stock,
			"unit":                  p.Unit,
			"price":                 p.Price,
			"image":                 p.Image,
			"is_shippable":          p.IsShippable,
			"is_digital":            p.IsDigital,
			"digital_download_link": p.DigitalDownloadLink,
			"product_cost":          p.ProductCost,
			"max_quantity_count":    p.MaxQuantityCount,
			"updated_at":            p.UpdatedAt,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (pu *ProductRepositoryImpl) IncreaseDownloadCounter(db *gorm.DB, pID, sID string) error {
	p := models.Product{}
	if err := db.Table(p.TableName()).
		Where("id = ? AND store_id = ?", pID, sID).
		Update("download_counter", gorm.Expr("download_counter + ?", 1)).Error; err != nil {
		return err
	}
	return nil
}

func (pu *ProductRepositoryImpl) IncreaseViewCounter(db *gorm.DB, pID, sID string) error {
	p := models.Product{}
	if err := db.Table(p.TableName()).
		Where("id = ? AND store_id = ?", pID, sID).
		Update("views", gorm.Expr("views + ?", 1)).Error; err != nil {
		return err
	}
	return nil
}

func (pu *ProductRepositoryImpl) List(db *gorm.DB, from, limit int) ([]models.ProductDetails, error) {
	var ps []models.ProductDetails
	p := models.Product{}
	if err := db.Table(p.TableName()).
		Select("products.id, products.stock, products.sku, products.slug, products.unit, products.store_id, s.name AS store_name, products.name, products.price, products.description, products.is_published, products.is_shippable, products.is_digital, c.id AS category_id, c.name AS category_name, products.image, products.created_at, products.updated_at").
		Joins("LEFT JOIN categories AS c ON products.category_id = c.id").
		Joins("LEFT JOIN stores AS s ON products.store_id = s.id").
		Where("products.is_published = ?", true).
		Offset(from).Limit(limit).
		Order("products.created_at DESC").Find(&ps).Error; err != nil {
		return nil, err
	}
	return ps, nil
}

func (pu *ProductRepositoryImpl) ListAsStoreStuff(db *gorm.DB, storeID string, from, limit int) ([]models.ProductDetailsInternal, error) {
	var ps []models.ProductDetailsInternal
	p := models.Product{}
	if err := db.Table(p.TableName()).
		Select("products.id, products.stock, products.sku, products.slug, products.unit, products.store_id, s.name AS store_name, products.name, products.price, products.description, products.is_published, products.is_shippable, products.is_digital, c.id AS category_id, c.name AS category_name, products.image, products.created_at, products.updated_at").
		Joins("LEFT JOIN categories AS c ON products.category_id = c.id").
		Joins("LEFT JOIN stores AS s ON products.store_id = s.id").
		Where("products.store_id = ?", storeID).
		Offset(from).Limit(limit).
		Order("created_at DESC").Find(&ps).Error; err != nil {
		return nil, err
	}
	return ps, nil
}

func (pu *ProductRepositoryImpl) Search(db *gorm.DB, query string, from, limit int) ([]models.ProductDetails, error) {
	var ps []models.ProductDetails
	p := models.Product{}
	if err := db.Table(p.TableName()).
		Select("products.id, products.name, products.sku, products.slug, products.unit, products.store_id, s.name AS store_name, products.stock, products.price, products.description, products.is_published, products.is_shippable, products.is_digital, c.id AS category_id, c.name AS category_name, products.image, products.created_at, products.updated_at").
		Joins("LEFT JOIN categories AS c ON products.category_id = c.id").
		Joins("LEFT JOIN collection_of_products AS cop ON products.id = cop.product_id").
		Joins("LEFT JOIN collections AS col ON cop.collection_id = col.id").
		Joins("LEFT JOIN stores AS s ON products.store_id = s.id").
		Where("products.is_published = ? AND (LOWER(products.name) LIKE ? OR LOWER(c.name) LIKE ? OR LOWER(col.name) LIKE ?)", true, "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%").
		Group("products.id, products.name, products.sku, products.unit, products.store_id, s.name, products.stock, products.price, products.description, products.is_published, products.is_shippable, products.is_digital, c.id, c.name, products.image, products.created_at, products.updated_at").
		Offset(from).Limit(limit).
		Order("created_at DESC").Find(&ps).Error; err != nil {
		return nil, err
	}
	return ps, nil
}

func (pu *ProductRepositoryImpl) SearchAsStoreStuff(db *gorm.DB, storeID, query string, from, limit int) ([]models.ProductDetailsInternal, error) {
	var ps []models.ProductDetailsInternal
	p := models.Product{}
	if err := db.Table(p.TableName()).
		Select("products.id, products.name, products.sku, products.slug, products.unit, products.store_id, s.name AS store_name, products.stock, products.price, products.description, products.is_published, products.is_shippable, products.is_digital, c.id AS category_id, c.name AS category_name, products.image, products.created_at, products.updated_at").
		Joins("LEFT JOIN categories AS c ON products.category_id = c.id").
		Joins("LEFT JOIN stores AS s ON products.store_id = s.id").
		Where("products.store_id = ? AND (LOWER(products.name) LIKE ? OR LOWER(c.name) LIKE ?)", storeID, "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%").
		Offset(from).Limit(limit).
		Order("created_at DESC").Find(&ps).Error; err != nil {
		return nil, err
	}
	return ps, nil
}

func (pu *ProductRepositoryImpl) Delete(db *gorm.DB, storeID, productID string) error {
	p := models.Product{}
	if err := db.Table(p.TableName()).
		Where("store_id = ? AND id = ?", storeID, productID).
		Delete(&p).Error; err != nil {
		return err
	}
	return nil
}

func (pu *ProductRepositoryImpl) Get(db *gorm.DB, productID string) (*models.Product, error) {
	p := models.Product{}
	if err := db.Table(fmt.Sprintf("%s", p.TableName())).
		Where("products.id = ? OR products.slug = ?", productID, productID).
		First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (pu *ProductRepositoryImpl) GetAsStoreStuff(db *gorm.DB, storeID, productID string) (*models.Product, error) {
	p := models.Product{}
	if err := db.Table(fmt.Sprintf("%s", p.TableName())).
		Where("(products.id = ? OR products.slug = ?) AND products.store_id = ?", productID, productID, storeID).
		First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (pu *ProductRepositoryImpl) GetDetails(db *gorm.DB, productID string) (*models.ProductDetails, error) {
	p := models.Product{}
	ps := models.ProductDetails{}
	cat := models.Category{}
	store := models.Store{}

	if err := db.Table(fmt.Sprintf("%s", p.TableName())).
		Select("products.id, s.id AS store_id, s.name AS store_name, products.max_quantity_count AS max_quantity_count, products.digital_download_link, products.price, products.product_cost, products.unit, products.stock, products.sku, products.name, products.slug, products.description, products.is_published, products.is_shippable, products.is_digital, c.id AS category_id, c.name AS category_name, products.image, products.created_at, products.updated_at").
		Joins(fmt.Sprintf("LEFT JOIN %s AS c ON products.category_id = c.id", cat.TableName())).
		Joins(fmt.Sprintf("LEFT JOIN %s AS s ON products.store_id = s.id", store.TableName())).
		Where("(products.id = ? OR products.slug = ?) AND products.is_published = ?", productID, productID, true).
		First(&ps).Error; err != nil {
		return nil, err
	}

	c := models.Collection{}
	cop := models.CollectionOfProduct{}
	var collections []models.Collection
	if err := db.Table(fmt.Sprintf("%s AS cop", cop.TableName())).
		Select("c.id, c.name, c.description").
		Joins(fmt.Sprintf("JOIN %s AS c ON cop.collection_id = c.id", c.TableName())).
		Where("cop.product_id = ?", ps.ID).
		Scan(&collections).Error; err != nil {
		return nil, err
	}
	ps.Collections = collections

	attributes, err := pu.ListAttributes(db, ps.ID)
	if err != nil {
		return nil, err
	}
	ps.Attributes = attributes

	additionalImages, err := pu.GetImages(db, ps.ID)
	if err != nil {
		return nil, err
	}
	ps.AdditionalImages = additionalImages

	return &ps, nil
}

func (pu *ProductRepositoryImpl) GetDetailsAsStoreStuff(db *gorm.DB, storeID, productID string) (*models.ProductDetailsInternal, error) {
	p := models.Product{}
	ps := models.ProductDetailsInternal{}
	cat := models.Category{}
	store := models.Store{}

	if err := db.Table(fmt.Sprintf("%s", p.TableName())).
		Select("products.id, s.id AS store_id, s.name AS store_name, products.max_quantity_count AS max_quantity_count, products.digital_download_link, products.price, products.product_cost, products.unit, products.stock, products.sku, products.name, products.slug, products.description, products.is_published, products.is_shippable, products.is_digital, c.id AS category_id, c.name AS category_name, products.image, products.created_at, products.updated_at").
		Joins(fmt.Sprintf("LEFT JOIN %s AS c ON products.category_id = c.id", cat.TableName())).
		Joins(fmt.Sprintf("LEFT JOIN %s AS s ON products.store_id = s.id", store.TableName())).
		Where("(products.id = ? OR products.slug = ?) AND products.store_id = ?", productID, productID, storeID).
		First(&ps).Error; err != nil {

		return nil, err
	}

	cop := models.CollectionOfProduct{}
	c := models.Collection{}
	var collections []models.Collection
	if err := db.Table(fmt.Sprintf("%s AS cop", cop.TableName())).
		Select("c.id, c.name, c.description").
		Joins(fmt.Sprintf("JOIN %s AS c ON cop.collection_id = c.id", c.TableName())).
		Where("c.store_id = ? AND cop.product_id = ?", storeID, productID).
		Scan(&collections).Error; err != nil {
		return nil, err
	}
	ps.Collections = collections

	attributes, err := pu.ListAttributes(db, ps.ID)
	if err != nil {
		return nil, err
	}
	ps.Attributes = attributes

	additionalImages, err := pu.GetImages(db, productID)
	if err != nil {
		return nil, err
	}
	ps.AdditionalImages = additionalImages

	return &ps, nil
}

func (pu *ProductRepositoryImpl) GetForOrder(db *gorm.DB, productID string, quantity int) (*models.Product, error) {
	p := models.Product{}

	if err := db.Table(p.TableName()).
		Where("id = ? AND (stock - ? >= 0 OR is_digital)", productID, quantity).
		Find(&p).Error; err != nil {
		return nil, err
	}

	if p.IsDigital {
		return &p, nil
	}

	if err := db.Table(p.TableName()).
		Where("id = ? AND stock - ? >= 0", productID, quantity).
		Update("stock", gorm.Expr("stock - ?", quantity)).Error; err != nil {
		return nil, err
	}

	return &p, nil
}

func (pu *ProductRepositoryImpl) Stats(db *gorm.DB, from, limit int) ([]helpers.ProductStats, error) {
	var stats []helpers.ProductStats

	p := models.Product{}
	oi := models.OrderedItem{}

	if err := db.Table(fmt.Sprintf("%s AS p", p.TableName())).
		Select("p.id AS id, p.name AS name, p.stock AS stock, p.price AS price, p.image AS image, p.description AS description, COALESCE(SUM(oi.quantity), 0) AS count_skip").
		Joins(fmt.Sprintf("LEFT JOIN %s AS oi ON p.id = oi.product_id", oi.TableName())).
		Group("p.id, p.name, p.stock, p.price, p.image, p.description").
		Order("count_skip DESC").
		Offset(from).
		Limit(limit).
		Find(&stats).Error; err != nil {
		return nil, err
	}

	if stats == nil {
		stats = []helpers.ProductStats{}
	}

	return stats, nil
}

func (pu *ProductRepositoryImpl) StatsAsStoreStaff(db *gorm.DB, storeID string, from, limit int) ([]helpers.ProductStats, error) {
	var stats []helpers.ProductStats

	p := models.Product{}
	oi := models.OrderedItem{}

	if err := db.Table(fmt.Sprintf("%s AS p", p.TableName())).
		Select("p.id AS id, p.name AS name, p.stock AS stock, p.price AS price, p.image AS image, p.description AS description, COALESCE(SUM(oi.quantity), 0) AS count").
		Joins(fmt.Sprintf("LEFT JOIN %s AS oi ON p.id = oi.product_id", oi.TableName())).
		Group("p.id, p.name, p.stock, p.price, p.image, p.description").
		Order("count DESC").
		Offset(from).
		Limit(limit).
		Find(&stats, "p.store_id = ?", storeID).Error; err != nil {
		return nil, err
	}

	if stats == nil {
		stats = []helpers.ProductStats{}
	}

	return stats, nil
}

func (pu *ProductRepositoryImpl) AddAttribute(db *gorm.DB, v *models.ProductAttribute) error {
	if err := db.Table(v.TableName()).Create(v).Error; err != nil {
		return err
	}
	return nil
}

func (pu *ProductRepositoryImpl) RemoveAttribute(db *gorm.DB, productID, attributeID string) error {
	v := models.ProductAttribute{}
	if err := db.Table(v.TableName()).Delete(&v, "product_id = ? AND id = ?", productID, attributeID).Error; err != nil {
		return err
	}
	return nil
}

func (pu *ProductRepositoryImpl) ListAttributes(db *gorm.DB, productID string) (map[string][]models.ProductKV, error) {
	v := models.ProductAttribute{}

	var attributes []models.ProductAttribute

	if err := db.Table(v.TableName()).Find(&attributes, "product_id = ?", productID).Error; err != nil {
		return nil, err
	}

	sortedAttributes := map[string][]models.ProductKV{}

	for _, a := range attributes {
		if _, ok := sortedAttributes[a.Key]; ok {
			sortedAttributes[a.Key] = append(sortedAttributes[a.Key], models.ProductKV{
				ID:    a.ID,
				Value: a.Value,
				Image: a.Image,
			})
		} else {
			sortedAttributes[a.Key] = []models.ProductKV{
				{
					ID:    a.ID,
					Value: a.Value,
					Image: a.Image,
				},
			}
		}
	}

	return sortedAttributes, nil
}

func (pu *ProductRepositoryImpl) GetAttribute(db *gorm.DB, productID, ID string) (*models.ProductAttribute, error) {
	v := models.ProductAttribute{}

	if err := db.Table(v.TableName()).Find(&v, "product_id = ? AND id = ?", productID, ID).Error; err != nil {
		return nil, err
	}

	return &v, nil
}

func (pu *ProductRepositoryImpl) ListByCollection(db *gorm.DB, collectionID string, from, limit int) ([]models.ProductDetails, error) {
	var ps []models.ProductDetails
	p := models.Product{}
	if err := db.Table(p.TableName()).
		Select("products.id, products.stock, products.sku, products.slug, products.unit, products.store_id, s.name AS store_name, products.name, products.price, products.description, products.is_published, products.is_shippable, products.is_digital, c.id AS category_id, c.name AS category_name, products.image, products.created_at, products.updated_at").
		Joins("LEFT JOIN categories AS c ON products.category_id = c.id").
		Joins("LEFT JOIN stores AS s ON products.store_id = s.id").
		Joins("LEFT JOIN collection_of_products AS cop ON cop.product_id = products.id").
		Where("products.is_published = ? AND cop.collection_id = ?", true, collectionID).
		Offset(from).Limit(limit).
		Order("products.created_at DESC").Find(&ps).Error; err != nil {
		return nil, err
	}
	return ps, nil
}

func (pu *ProductRepositoryImpl) ListByCollectionAsStoreStuff(db *gorm.DB, collectionID string, from, limit int) ([]models.ProductDetails, error) {
	var ps []models.ProductDetails
	p := models.Product{}
	if err := db.Table(p.TableName()).
		Select("products.id, products.stock, products.sku, products.slug, products.unit, products.store_id, s.name AS store_name, products.name, products.price, products.description, products.is_published, products.is_shippable, products.is_digital, c.id AS category_id, c.name AS category_name, products.image, products.created_at, products.updated_at").
		Joins("LEFT JOIN categories AS c ON products.category_id = c.id").
		Joins("LEFT JOIN stores AS s ON products.store_id = s.id").
		Joins("LEFT JOIN collection_of_products AS cop ON cop.product_id = products.id").
		Where("cop.collection_id = ?", collectionID).
		Offset(from).Limit(limit).
		Order("products.created_at DESC").Find(&ps).Error; err != nil {
		return nil, err
	}
	return ps, nil
}

func (pu *ProductRepositoryImpl) AddImage(db *gorm.DB, productID, imagePath string) error {
	pi := models.ProductImage{
		ProductID: productID,
		ImagePath: imagePath,
	}

	if err := db.Table(pi.TableName()).Create(&pi).Error; err != nil {
		return err
	}
	return nil
}

func (pu *ProductRepositoryImpl) GetImages(db *gorm.DB, productID string) ([]string, error) {
	var images []models.ProductImage
	pi := models.ProductImage{}
	if err := db.Table(pi.TableName()).Where("product_id = ?", productID).Find(&images).Error; err != nil {
		return nil, err
	}

	var filteredImages []string
	for _, v := range images {
		filteredImages = append(filteredImages, v.ImagePath)
	}
	return filteredImages, nil
}

func (pu *ProductRepositoryImpl) RemoveImage(db *gorm.DB, productID string) error {
	pi := models.ProductImage{}
	if err := db.Table(pi.TableName()).Where("product_id = ?", productID).Delete(&pi).Error; err != nil {
		return err
	}
	return nil
}
