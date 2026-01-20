package request

import (
	"context"
	"errors"
	"mime/multipart"
	"path/filepath"
	"time"

	entity "github.com/kareemhamed001/POS/internal/core/domain"
	"github.com/kareemhamed001/POS/pkg/file"
)

type CreateProductRequest struct {
	Name              string                `form:"name" validate:"required,min=2,max=100"`
	ShortDescription  *string               `form:"short_description" validate:"omitempty,min=2,max=255"`
	Description       string                `form:"description" validate:"required"`
	Price             float32               `form:"price" validate:"required,gt=0"`
	DiscountType      string                `form:"discount_type" validate:"omitempty,discount_type_valid"`
	DiscountValue     float32               `form:"discount_value" validate:"omitempty,gte=0"`
	DiscountStartDate *string               `form:"discount_start_date" validate:"omitempty,datetime=2006-01-02 15:04:05"`
	DiscountEndDate   *string               `form:"discount_end_date" validate:"omitempty,datetime=2006-01-02 15:04:05"`
	Image             *multipart.FileHeader `form:"image" validate:"omitempty" `
	Quantity          int                   `form:"quantity" validate:"required,gte=0"`
}

type UpdateProductRequest struct {
	Name              *string               `form:"name" validate:"omitempty,min=2,max=100"`
	ShortDescription  *string               `form:"short_description" validate:"omitempty,min=2,max=255"`
	Description       *string               `form:"description" validate:"omitempty"`
	Price             *float32              `form:"price" validate:"omitempty,gt=0"`
	DiscountType      *string               `form:"discount_type" validate:"omitempty,discount_type_valid_ptr"`
	DiscountValue     *float32              `form:"discount_value" validate:"omitempty,gte=0"`
	DiscountStartDate *string               `form:"discount_start_date" validate:"omitempty,datetime=2006-01-02 15:04:05"`
	DiscountEndDate   *string               `form:"discount_end_date" validate:"omitempty,datetime=2006-01-02 15:04:05"`
	Image             *multipart.FileHeader `form:"image" validate:"omitempty"`
	Quantity          *int                  `form:"quantity" validate:"omitempty,gte=0"`
}

func getFormatedTimePtr(dateStr *string) (*time.Time, error) {
	if dateStr == nil {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02 15:04:05", *dateStr)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func ensureDiscountDatesValidity(startDate, endDate *time.Time) error {
	if startDate != nil && endDate != nil {
		if endDate.Before(*startDate) {
			return errors.New("discount_end_date must be after discount_start_date")
		}
	}
	return nil
}

// uploadProductImage validates and uploads a product image file
func uploadProductImage(ctx context.Context, fileHeader *multipart.FileHeader, fileStorage file.FileStorage) (string, error) {
	if fileHeader == nil {
		return "", nil
	}

	// Validate extension
	ext := filepath.Ext(fileHeader.Filename)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		// ok
	default:
		return "", errors.New("unsupported image format")
	}

	// Generate unique filename and path
	filename := file.GenerateFileName(fileHeader.Filename)
	filePath := file.GenerateFilePath(filename)

	// Upload file using the storage interface
	imageURL, err := fileStorage.Upload(ctx, fileHeader, filePath)
	if err != nil {
		return "", errors.New("failed to save image: " + err.Error())
	}

	return imageURL, nil
}

func (r *CreateProductRequest) ToProduct(ctx context.Context, fileStorage file.FileStorage) (*entity.Product, error) {

	discountStartDatePtr, err := getFormatedTimePtr(r.DiscountStartDate)
	if err != nil {
		return nil, errors.New("invalid discount_start_date format")
	}

	discountEndDatePtr, err := getFormatedTimePtr(r.DiscountEndDate)
	if err != nil {
		return nil, errors.New("invalid discount_end_date format")
	}

	if err := ensureDiscountDatesValidity(discountStartDatePtr, discountEndDatePtr); err != nil {
		return nil, err
	}

	// Handle image file upload (optional)
	imageURL, err := uploadProductImage(ctx, r.Image, fileStorage)
	if err != nil {
		return nil, err
	}

	var imageUrlPtr *string
	if imageURL != "" {
		imageUrlPtr = &imageURL
	}

	return &entity.Product{
		Name:              r.Name,
		ShortDescription:  r.ShortDescription,
		Description:       r.Description,
		Price:             r.Price,
		DiscountType:      entity.DiscountType(r.DiscountType),
		DiscountValue:     r.DiscountValue,
		DiscountStartDate: discountStartDatePtr,
		DiscountEndDate:   discountEndDatePtr,
		ImageUrl:          imageUrlPtr,
		Quantity:          r.Quantity,
	}, nil
}

func (r *UpdateProductRequest) ToProduct(existingProduct *entity.Product, ctx context.Context, fileStorage file.FileStorage) (*entity.Product, error) {
	updatedProduct := *existingProduct // Start with existing product

	if r.Name != nil {
		updatedProduct.Name = *r.Name
	}
	if r.ShortDescription != nil {
		updatedProduct.ShortDescription = r.ShortDescription
	}
	if r.Description != nil {
		updatedProduct.Description = *r.Description
	}
	if r.Price != nil {
		updatedProduct.Price = *r.Price
	}
	if r.DiscountType != nil {
		updatedProduct.DiscountType = entity.DiscountType(*r.DiscountType)
	}
	if r.DiscountValue != nil {
		updatedProduct.DiscountValue = *r.DiscountValue
	}
	if r.DiscountStartDate != nil {
		startDate, err := getFormatedTimePtr(r.DiscountStartDate)
		if err != nil {
			return nil, errors.New("invalid discount_start_date format")
		}
		updatedProduct.DiscountStartDate = startDate
	}
	if r.DiscountEndDate != nil {
		endDate, err := getFormatedTimePtr(r.DiscountEndDate)
		if err != nil {
			return nil, errors.New("invalid discount_end_date format")
		}
		updatedProduct.DiscountEndDate = endDate
	}

	if err := ensureDiscountDatesValidity(updatedProduct.DiscountStartDate, updatedProduct.DiscountEndDate); err != nil {
		return nil, err
	}

	if r.Image != nil {
		imageURL, err := uploadProductImage(ctx, r.Image, fileStorage)
		if err != nil {
			return nil, err
		}
		if imageURL != "" {
			updatedProduct.ImageUrl = &imageURL
		}
	}
	if r.Quantity != nil {
		updatedProduct.Quantity = *r.Quantity
	}

	return &updatedProduct, nil
}
