package data

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/utils"
	"time"
)

type UserRepositoryImpl struct {
}

var userRepository UserRepository

func NewUserRepository() UserRepository {
	if userRepository == nil {
		userRepository = &UserRepositoryImpl{}
	}
	return userRepository
}

func (uu *UserRepositoryImpl) Register(db *gorm.DB, u *models.User) error {
	if err := db.Model(u).Create(u).Error; err != nil {
		return err
	}
	return nil
}

func (uu *UserRepositoryImpl) Login(db *gorm.DB, email string) (*models.User, error) {
	u := models.User{}

	if err := db.Model(&u).Where("email = ?", email).Find(&u).Error; err != nil {
		return nil, err
	}

	return &u, nil
}

func (uu *UserRepositoryImpl) CreateSession(db *gorm.DB, s *models.Session) error {
	if err := db.Model(&s).Create(&s).Error; err != nil {
		return err
	}
	return nil
}

func (uu *UserRepositoryImpl) Logout(db *gorm.DB, token string) error {
	s := models.Session{}

	if err := db.Model(&s).Where("access_token = ?", token).Delete(&s).Error; err != nil {
		return err
	}
	return nil
}

func (uu *UserRepositoryImpl) RefreshToken(db *gorm.DB, token string) (*models.Session, error) {
	os := models.Session{}

	if err := db.Model(&os).Where("refresh_token = ?", token).First(&os).Error; err != nil {
		return nil, err
	}

	s := models.Session{
		ID:           utils.NewUUID(),
		UserID:       os.UserID,
		AccessToken:  utils.NewToken(),
		RefreshToken: utils.NewToken(),
		CreatedAt:    time.Now().UTC(),
		ExpireOn:     time.Now().Add(time.Hour * 48).Unix(),
	}

	if err := db.Model(&s).Create(&s).Error; err != nil {
		return nil, err
	}

	if err := db.Model(&os).Where("refresh_token = ?", token).Delete(&os).Error; err != nil {
		return nil, err
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}

	return &s, nil
}

func (uu *UserRepositoryImpl) Update(db *gorm.DB, u *models.User) error {
	if err := db.Table(u.TableName()).
		Where("id = ?", u.ID).
		Select("name, profile_picture, phone, password, reset_password_token, reset_password_token_generated_at, verification_token, is_email_verified, status, permission_id, updated_at").
		Updates(map[string]interface{}{
			"name":                              u.Name,
			"profile_picture":                   u.ProfilePicture,
			"phone":                             u.Phone,
			"password":                          u.Password,
			"verification_token":                u.VerificationToken,
			"reset_password_token":              u.ResetPasswordToken,
			"reset_password_token_generated_at": u.ResetPasswordTokenGeneratedAt,
			"is_email_verified":                 u.IsEmailVerified,
			"status":                            u.Status,
			"permission_id":                     u.PermissionID,
			"updated_at":                        u.UpdatedAt,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (uu *UserRepositoryImpl) GetPermission(db *gorm.DB, token string) (string, *models.Permission, error) {
	s := models.Session{}
	u := models.User{}
	up := models.UserPermission{}

	result := struct {
		ID         string             `json:"id"`
		Permission *models.Permission `json:"permission"`
	}{}

	if err := db.Table(fmt.Sprintf("%s AS s", s.TableName())).
		Select("u.id, up.permission").
		Joins(fmt.Sprintf("JOIN %s AS u ON u.id = s.user_id", u.TableName())).
		Joins(fmt.Sprintf("JOIN %s AS up ON u.permission_id = up.id", up.TableName())).
		Where("s.access_token = ? AND u.status = ?", token, models.UserActive).Scan(&result).Error; err != nil {
		return "", nil, err
	}
	return result.ID, result.Permission, nil
}

func (uu *UserRepositoryImpl) GetPermissionByUserID(db *gorm.DB, userID string) (string, *models.Permission, error) {
	u := models.User{}
	up := models.UserPermission{}

	result := struct {
		ID         string             `json:"id"`
		Permission *models.Permission `json:"permission"`
	}{}

	if err := db.Table(fmt.Sprintf("%s AS u", u.TableName())).
		Select("u.id, up.permission").
		Joins(fmt.Sprintf("JOIN %s AS up ON u.permission_id = up.id", up.TableName())).
		Where("u.id = ? AND u.status = ?", userID, models.UserActive).Scan(&result).Error; err != nil {
		return "", nil, err
	}
	return result.ID, result.Permission, nil
}

func (uu *UserRepositoryImpl) Get(db *gorm.DB, userID string) (*models.User, error) {
	u := models.User{}

	if err := db.Model(&u).Where("id = ?", userID).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (uu *UserRepositoryImpl) GetByEmail(db *gorm.DB, email string) (*models.User, error) {
	u := models.User{}

	if err := db.Model(&u).Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (uu *UserRepositoryImpl) IsSignUpEnabled(db *gorm.DB) (bool, error) {
	s := models.Settings{}
	if err := db.Model(&s).Where("id = ?", "1").First(&s).Error; err != nil {
		return false, err
	}
	return s.IsSignUpEnabled, nil
}

func (uu *UserRepositoryImpl) IsStoreCreationEnabled(db *gorm.DB) (bool, error) {
	s := models.Settings{}
	if err := db.Model(&s).Where("id = ?", "1").First(&s).Error; err != nil {
		return false, err
	}
	return s.IsStoreCreationEnabled, nil
}

func (uu *UserRepositoryImpl) List(db *gorm.DB, from, limit int) ([]models.User, error) {
	var users []models.User

	u := models.User{}
	if err := db.Table(u.TableName()).Offset(from).Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (uu *UserRepositoryImpl) Search(db *gorm.DB, query string, from, limit int) ([]models.User, error) {
	var users []models.User

	u := models.User{}
	if err := db.Table(u.TableName()).
		Offset(from).
		Limit(limit).
		Where("email LIKE ? OR phone LIKE ?", "%"+query+"%", "%"+query+"%").
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
