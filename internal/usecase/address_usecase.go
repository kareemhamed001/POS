package usecase

import (
	"context"

	"github.com/kareemhamed001/POS/internal/entity"
)

type AddressUsecase struct {
	addressRepo AddressRepository
}

func NewAddressUsecase(addressRepo AddressRepository) *AddressUsecase {
	return &AddressUsecase{
		addressRepo: addressRepo,
	}
}

func (u *AddressUsecase) AddAddress(ctx context.Context, address *entity.Address) error {
	return u.addressRepo.AddAddress(ctx, address)
}

func (u *AddressUsecase) GetUserAddresses(ctx context.Context, userID uint) ([]entity.Address, error) {
	return u.addressRepo.GetAddressesByUserID(ctx, userID)
}

func (u *AddressUsecase) UpdateAddress(ctx context.Context, address *entity.Address) error {
	return u.addressRepo.UpdateAddress(ctx, address)
}

func (u *AddressUsecase) DeleteAddress(ctx context.Context, id uint) error {
	return u.addressRepo.DeleteAddress(ctx, id)
}
