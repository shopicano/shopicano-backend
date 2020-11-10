package data

import (
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/models"
)

type AddressRepositoryImpl struct {
}

var addressRepository *AddressRepositoryImpl

func NewAddressRepository() *AddressRepositoryImpl {
	if addressRepository == nil {
		addressRepository = &AddressRepositoryImpl{}
	}
	return addressRepository
}

func (au *AddressRepositoryImpl) CreateAddress(db *gorm.DB, a *models.Address) error {
	if err := db.Table(a.TableName()).Create(a).Error; err != nil {
		return err
	}
	return nil
}

func (au *AddressRepositoryImpl) UpdateAddress(db *gorm.DB, a *models.Address) error {
	if err := db.Table(a.TableName()).
		Where("id = ? AND user_id = ?", a.ID, a.UserID).Updates(a).Error; err != nil {
		return err
	}
	return nil
}

func (au *AddressRepositoryImpl) GetAddress(db *gorm.DB, userID, addressID string) (*models.AddressView, error) {
	a := models.AddressView{}
	if err := db.Table(a.TableName()).
		Where("id = ? AND user_id = ?", addressID, userID).Find(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (au *AddressRepositoryImpl) GetAddressByID(db *gorm.DB, addressID string) (*models.AddressView, error) {
	a := models.AddressView{}
	if err := db.Table(a.TableName()).
		Where("id = ?", addressID).Find(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (au *AddressRepositoryImpl) GetRawAddressByID(db *gorm.DB, addressID string) (*models.Address, error) {
	a := models.Address{}
	if err := db.Table(a.TableName()).
		Where("id = ?", addressID).Find(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (au *AddressRepositoryImpl) ListAddresses(db *gorm.DB, userId string, from, limit int) ([]models.AddressView, error) {
	var addresses []models.AddressView
	address := models.AddressView{}
	if err := db.Table(address.TableName()).
		Offset(from).Limit(limit).
		Where("user_id = ?", userId).
		Order("created_at DESC").Find(&addresses).Error; err != nil {
		return nil, err
	}
	return addresses, nil
}

func (au *AddressRepositoryImpl) DeleteAddress(db *gorm.DB, userID, addressID string) error {
	address := models.Address{}
	if err := db.Table(address.TableName()).
		Where("user_id = ? AND id = ?", userID, addressID).
		Delete(&address).Error; err != nil {
		return err
	}
	return nil
}
