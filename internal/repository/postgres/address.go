package postgres

import (
	"context"
	"errors"

	"github.com/kareemhamed001/POS/internal/entity"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var (
	ErrAddressNotFound = errors.New("Address not found")
)

type AddressRepository struct {
	db     *gorm.DB
	tracer trace.Tracer
}

func NewAddressRepository(db *gorm.DB) *AddressRepository {
	return &AddressRepository{
		db:     db,
		tracer: otel.Tracer("address-repo"),
	}
}

func (r *AddressRepository) AddAddress(ctx context.Context, address *entity.Address) error {
	ctx, span := r.tracer.Start(ctx, "AddressRepository.AddAddress")
	defer span.End()

	span.SetAttributes(
		attribute.Int("address.user_id", int(address.UserID)),
		attribute.String("address.city", address.City),
	)

	if err := gorm.G[entity.Address](r.db).Create(ctx, address); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "address created")
	return nil
}

func (r *AddressRepository) GetAddressesByUserID(ctx context.Context, userID uint) ([]entity.Address, error) {
	ctx, span := r.tracer.Start(ctx, "AddressRepository.GetAddressesByUserID")
	defer span.End()

	span.SetAttributes(attribute.Int("address.user_id", int(userID)))

	addresses, err := gorm.G[entity.Address](r.db).
		Where("user_id = ?", userID).
		Find(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("address.count", len(addresses)))
	span.SetStatus(codes.Ok, "addresses retrieved")
	return addresses, nil
}

func (r *AddressRepository) UpdateAddress(ctx context.Context, address *entity.Address) error {
	ctx, span := r.tracer.Start(ctx, "AddressRepository.UpdateAddress")
	defer span.End()

	span.SetAttributes(attribute.Int("address.id", int(address.ID)))

	rowsAffected, err := gorm.G[entity.Address](r.db).
		Where("id = ?", address.ID).
		Updates(ctx, *address)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if rowsAffected == 0 {
		span.SetStatus(codes.Error, ErrAddressNotFound.Error())
		return ErrAddressNotFound
	}

	span.SetStatus(codes.Ok, "address updated")
	return nil
}

func (r *AddressRepository) DeleteAddress(ctx context.Context, id uint) error {
	ctx, span := r.tracer.Start(ctx, "AddressRepository.DeleteAddress")
	defer span.End()

	span.SetAttributes(attribute.Int("address.id", int(id)))

	rowsAffected, err := gorm.G[entity.Address](r.db).
		Where("id = ?", id).
		Delete(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if rowsAffected == 0 {
		span.SetStatus(codes.Error, ErrAddressNotFound.Error())
		return ErrAddressNotFound
	}

	span.SetStatus(codes.Ok, "address deleted")
	return nil
}
