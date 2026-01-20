package usecase

import (
	"context"

	"github.com/kareemhamed001/POS/internal/entity"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type AddressUsecase struct {
	addressRepo AddressRepository
	tracer      trace.Tracer
}

func NewAddressUsecase(addressRepo AddressRepository) *AddressUsecase {
	return &AddressUsecase{
		addressRepo: addressRepo,
		tracer:      otel.Tracer("address-usecase"),
	}
}

func (u *AddressUsecase) AddAddress(ctx context.Context, address *entity.Address) error {
	ctx, span := u.tracer.Start(ctx, "AddressUsecase.AddAddress")
	defer span.End()

	span.SetAttributes(
		attribute.Int("address.user_id", int(address.UserID)),
		attribute.String("address.city", address.City),
	)

	if err := u.addressRepo.AddAddress(ctx, address); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "address added")
	return nil
}

func (u *AddressUsecase) GetUserAddresses(ctx context.Context, userID uint) ([]entity.Address, error) {
	ctx, span := u.tracer.Start(ctx, "AddressUsecase.GetUserAddresses")
	defer span.End()

	span.SetAttributes(attribute.Int("address.user_id", int(userID)))

	addresses, err := u.addressRepo.GetAddressesByUserID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("address.count", len(addresses)))
	span.SetStatus(codes.Ok, "addresses retrieved")
	return addresses, nil
}

func (u *AddressUsecase) UpdateAddress(ctx context.Context, address *entity.Address) error {
	ctx, span := u.tracer.Start(ctx, "AddressUsecase.UpdateAddress")
	defer span.End()

	span.SetAttributes(attribute.Int("address.id", int(address.ID)))

	if err := u.addressRepo.UpdateAddress(ctx, address); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "address updated")
	return nil
}

func (u *AddressUsecase) DeleteAddress(ctx context.Context, id uint) error {
	ctx, span := u.tracer.Start(ctx, "AddressUsecase.DeleteAddress")
	defer span.End()

	span.SetAttributes(attribute.Int("address.id", int(id)))

	if err := u.addressRepo.DeleteAddress(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "address deleted")
	return nil
}
