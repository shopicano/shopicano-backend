package validators

import (
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/utils"
	"github.com/shopicano/shopicano-backend/values"
	"time"
)

func ValidateCreateStore(ctx echo.Context) (*models.Store, error) {
	pld := struct {
		Name        string `json:"name" valid:"required,stringlength(1|100)"`
		AddressID   string `json:"address_id" valid:"required"`
		Description string `json:"description" valid:"required,stringlength(1|1000)"`
		LogoImage   string `json:"logo_image"`
		CoverImage  string `json:"cover_image"`
	}{}

	if err := ctx.Bind(&pld); err != nil {
		return nil, err
	}

	ok, err := govalidator.ValidateStruct(&pld)
	if ok {
		return &models.Store{
			ID:                       utils.NewUUID(),
			Name:                     pld.Name,
			Status:                   models.StoreRegistered,
			Description:              pld.Description,
			CoverImage:               pld.CoverImage,
			LogoImage:                pld.LogoImage,
			IsOrderCreationEnabled:   false,
			IsProductCreationEnabled: false,
			AddressID:                pld.AddressID,
			CreatedAt:                time.Now().UTC(),
			UpdatedAt:                time.Now().UTC(),
		}, nil
	}

	ve := errors.ValidationError{}

	for k, v := range govalidator.ErrorsByField(err) {
		ve.Add(k, v)
	}

	return nil, &ve
}

type reqStoreUpdate struct {
	Name                     *string `json:"name"`
	Address                  *string `json:"address"`
	City                     *string `json:"city"`
	State                    *string `json:"state"`
	CountryID                *int64  `json:"country_id"`
	Postcode                 *string `json:"postcode"`
	Email                    *string `json:"email"`
	Phone                    *string `json:"phone"`
	Description              *string `json:"description"`
	LogoImage                *string `json:"logo_image"`
	CoverImage               *string `json:"cover_image"`
	IsProductCreationEnabled *bool   `json:"is_product_creation_enabled"`
	IsOrderCreationEnabled   *bool   `json:"is_order_creation_enabled"`
	IsAutoConfirmEnabled     *bool   `json:"is_auto_confirm_enabled"`
}

func ValidateUpdateStore(ctx echo.Context) (*reqStoreUpdate, error) {
	pld := reqStoreUpdate{}

	if err := ctx.Bind(&pld); err != nil {
		return nil, err
	}

	ok, err := govalidator.ValidateStruct(&pld)
	if ok {
		return &pld, nil
	}

	ve := errors.ValidationError{}

	for k, v := range govalidator.ErrorsByField(err) {
		ve.Add(k, v)
	}

	return nil, &ve
}

func ValidateCreateStoreStaff(ctx echo.Context) (*string, *string, error) {
	pld := struct {
		Email        string `json:"email" valid:"required,email"`
		PermissionID string `json:"permission_id" valid:"required"`
	}{}

	if err := ctx.Bind(&pld); err != nil {
		return nil, nil, err
	}

	ve := errors.ValidationError{}

	ok, err := govalidator.ValidateStruct(&pld)
	if ok {
		if pld.PermissionID == values.AdminGroupID || pld.PermissionID == values.ManagerGroupID {
			return &pld.Email, &pld.PermissionID, nil
		}

		ve.Add("permission_id", "is invalid")
	}

	for k, v := range govalidator.ErrorsByField(err) {
		ve.Add(k, v)
	}

	return nil, nil, &ve
}

func ValidateUpdateStoreStaff(ctx echo.Context) (*string, error) {
	pld := struct {
		PermissionID string `json:"permission_id" valid:"required"`
	}{}

	if err := ctx.Bind(&pld); err != nil {
		return nil, err
	}

	ve := errors.ValidationError{}

	ok, err := govalidator.ValidateStruct(&pld)
	if ok {
		if pld.PermissionID == values.AdminGroupID || pld.PermissionID == values.ManagerGroupID {
			return &pld.PermissionID, nil
		}

		ve.Add("permission_id", "is invalid")
	}

	for k, v := range govalidator.ErrorsByField(err) {
		ve.Add(k, v)
	}

	return nil, &ve
}

func ValidateUpdateStoreStatus(ctx echo.Context) (*models.StoreStatus, *int64, error) {
	pld := struct {
		Status         *models.StoreStatus `json:"status"`
		CommissionRate *int64              `json:"commission_rate"`
	}{}

	if err := ctx.Bind(&pld); err != nil {
		return nil, nil, err
	}

	ve := errors.ValidationError{}

	if pld.Status != nil && !pld.Status.IsValid() {
		ve.Add("status", "is invalid")
	}
	if pld.CommissionRate != nil && (*pld.CommissionRate < 0 || *pld.CommissionRate > 100) {
		ve.Add("commission_rate", "is invalid")
	}

	if len(ve) > 0 {
		return nil, nil, &ve
	}

	return pld.Status, pld.CommissionRate, nil
}
