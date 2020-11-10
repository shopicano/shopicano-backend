package data

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/models"
	"strings"
)

type CustomerRepositoryImpl struct {
}

var customerRepository CustomerRepository

func NewCustomerRepository() CustomerRepository {
	if customerRepository == nil {
		customerRepository = &CustomerRepositoryImpl{}
	}
	return customerRepository
}

func (cr *CustomerRepositoryImpl) List(db *gorm.DB, storeID string, offset, limit int) ([]models.Customer, error) {
	var customers []models.Customer

	u := models.User{}
	odv := models.OrderDetailsView{}

	if err := db.Table(fmt.Sprintf("%s AS u", u.TableName())).
		Select("u.id AS id, u.name AS name, u.email AS email, u.profile_picture AS profile_picture, u.phone AS phone,"+
			" u.is_email_verified AS is_email_verified, odv.store_id AS store_id, COUNT(odv.id) AS number_of_purchases").
		Joins(fmt.Sprintf("JOIN %s AS odv ON u.id = odv.user_id", odv.TableName())).
		Group("u.id, odv.store_id").
		Where("store_id = ?", storeID).
		Find(&customers).Error; err != nil {
		return nil, err
	}
	return customers, nil
}

func (cr *CustomerRepositoryImpl) Search(db *gorm.DB, query, storeID string, offset, limit int) ([]models.Customer, error) {
	var customers []models.Customer

	u := models.User{}
	odv := models.OrderDetailsView{}

	if err := db.Table(fmt.Sprintf("%s AS u", u.TableName())).
		Select("u.id AS id, u.name AS name, u.email AS email, u.profile_picture AS profile_picture, u.phone AS phone,"+
			" u.is_email_verified AS is_email_verified, odv.store_id AS store_id, COUNT(odv.id) AS number_of_purchases").
		Joins(fmt.Sprintf("JOIN %s AS odv ON u.id = odv.user_id", odv.TableName())).
		Group("u.id, odv.store_id").
		Where("store_id = ? AND (LOWER(name) LIKE ? OR LOWER(email) LIKE ?)", storeID, "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%").
		Find(&customers).Error; err != nil {
		return nil, err
	}
	return customers, nil
}
