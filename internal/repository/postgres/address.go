package postgres

import (
	"context"
	"errors"

	"github.com/kareemhamed001/POS/internal/entity"
	"gorm.io/gorm"
)

var (
	ErrAddressNotFound = errors.New("Address not found")
)

type AddressRepository struct {
	db *gorm.DB
}

func NewAddressRepository(db *gorm.DB) *AddressRepository {
	return &AddressRepository{db: db}
}

func (r *AddressRepository) AddAddress(ctx context.Context, address *entity.Address) error {
	return gorm.G[entity.Address](r.db).Create(ctx, address)
}

func (r *AddressRepository) GetAddressesByUserID(ctx context.Context, userID uint) ([]entity.Address, error) {
	addresses, err := gorm.G[entity.Address](r.db).
		Where("user_id = ?", userID).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	return addresses, nil
}

func (r *AddressRepository) UpdateAddress(ctx context.Context, address *entity.Address) error {
	rowsAffected, err := gorm.G[entity.Address](r.db).
		Where("id = ?", address.ID).
		Updates(ctx, *address)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrAddressNotFound
	}
	return nil
}

func (r *AddressRepository) DeleteAddress(ctx context.Context, id uint) error {
	rowsAffected, err := gorm.G[entity.Address](r.db).
		Where("id = ?", id).
		Delete(ctx)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrAddressNotFound
	}
	return nil
}
