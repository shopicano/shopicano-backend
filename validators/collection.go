package validators

import (
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"github.com/shopicano/shopicano-backend/errors"
	"github.com/shopicano/shopicano-backend/models"
	"github.com/shopicano/shopicano-backend/utils"
	"time"
)

func ValidateCreateCollection(ctx echo.Context) (*models.Collection, error) {
	pld := struct {
		Name        string `json:"name" valid:"required,stringlength(1|100)"`
		Description string `json:"description" valid:"required,stringlength(1|500)"`
		Image       string `json:"image"`
		IsPublished bool   `json:"is_published"`
	}{}

	if err := ctx.Bind(&pld); err != nil {
		return nil, err
	}

	ok, err := govalidator.ValidateStruct(&pld)
	if ok {
		return &models.Collection{

			ID:          utils.NewUUID(),
			Name:        pld.Name,
			Description: pld.Description,
			Image:       pld.Image,
			IsPublished: pld.IsPublished,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}, nil
	}

	ve := errors.ValidationError{}

	for k, v := range govalidator.ErrorsByField(err) {
		ve.Add(k, v)
	}

	return nil, &ve
}

type ReqCollectionUpdate struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Image       *string `json:"image"`
	IsPublished *bool   `json:"is_published"`
}

func ValidateUpdateCollection(ctx echo.Context) (*ReqCollectionUpdate, error) {
	pld := ReqCollectionUpdate{}

	if err := ctx.Bind(&pld); err != nil {
		return nil, err
	}

	ve := errors.ValidationError{}
	if pld.Name != nil {
		ok := len(*pld.Name) >= 1 && len(*pld.Name) <= 100
		if !ok {
			ve.Add("name", "must be between 1 to 100 characters")
		}
	}
	if pld.Description != nil {
		ok := len(*pld.Name) >= 1 && len(*pld.Name) <= 500
		if !ok {
			ve.Add("description", "must be between 1 to 500 characters")
		}
	}

	if len(ve) == 0 {
		return &pld, nil
	}

	return nil, &ve
}
