package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kareemhamed001/POS/internal/entity"
	"github.com/kareemhamed001/POS/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var (
	ErrDuplicatedKey = errors.New("User with same email or phone already exists")
	ErrUserNotFound  = errors.New("User not found")
)

type UserRepository struct {
	db     *gorm.DB
	tracer trace.Tracer
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db:     db,
		tracer: otel.Tracer("user-repo"),
	}
}

func (u *UserRepository) ListUsers(ctx context.Context, search string, page, perPage int) (*[]entity.User, int, error) {
	ctx, span := u.tracer.Start(ctx, "UserRepository.ListUsers")
	defer span.End()

	span.SetAttributes(
		attribute.String("query.search", search),
		attribute.Int("query.page", page),
		attribute.Int("query.per_page", perPage),
	)

	// Use LOWER for case-insensitive search (works with both PostgreSQL and SQLite)
	searchPattern := "%" + search + "%"
	total, err := gorm.G[entity.User](u.db).
		Where("LOWER(name) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?) OR LOWER(phone) LIKE LOWER(?)", searchPattern, searchPattern, searchPattern).
		Count(ctx, "id")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, 0, err
	}
	skip := (page - 1) * perPage
	users, err := gorm.G[entity.User](u.db).
		Where("LOWER(name) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?) OR LOWER(phone) LIKE LOWER(?)", searchPattern, searchPattern, searchPattern).
		Offset(skip).
		Limit(perPage).
		Find(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, 0, err
	}

	span.SetAttributes(attribute.Int("users.count", len(users)))
	span.SetStatus(codes.Ok, "users listed")

	return &users, int(total), nil
}
func (u *UserRepository) CreateUser(ctx context.Context, user *entity.User) error {
	ctx, span := u.tracer.Start(ctx, "UserRepository.CreateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.email", user.Email),
		attribute.String("user.phone", user.Phone),
	)

	err := gorm.G[entity.User](u.db).Create(ctx, user)
	if err != nil {
		// Check for PostgreSQL unique constraint violation
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation
				span.SetStatus(codes.Error, ErrDuplicatedKey.Error())
				return ErrDuplicatedKey
			}
		}
		// Check for SQLite unique constraint violation
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			span.SetStatus(codes.Error, ErrDuplicatedKey.Error())
			return ErrDuplicatedKey
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(attribute.Int("user.id", int(user.ID)))
	span.SetStatus(codes.Ok, "user created")
	return nil
}
func (u *UserRepository) GetUserByID(ctx context.Context, id uint) (*entity.User, error) {
	ctx, span := u.tracer.Start(ctx, "UserRepository.GetUserByID")
	defer span.End()

	span.SetAttributes(attribute.Int("user.id", int(id)))

	user, err := gorm.G[entity.User](u.db).Where("id=?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			span.SetStatus(codes.Error, ErrUserNotFound.Error())
			return nil, ErrUserNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err

	}
	span.SetStatus(codes.Ok, "user retrieved")
	return &user, nil
}

func (u *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	ctx, span := u.tracer.Start(ctx, "UserRepository.GetUserByEmail")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	user, err := gorm.G[entity.User](u.db).Where("email=?", email).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			span.SetStatus(codes.Error, ErrUserNotFound.Error())
			return nil, ErrUserNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	span.SetStatus(codes.Ok, "user retrieved")
	return &user, nil
}
func (u *UserRepository) UpdateUser(ctx context.Context, id uint, user *entity.User) error {
	ctx, span := u.tracer.Start(ctx, "UserRepository.UpdateUser")
	defer span.End()

	span.SetAttributes(attribute.Int("user.id", int(id)))

	rowsAffected, err := gorm.G[*entity.User](u.db).Where("id=?", id).Updates(ctx, user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			span.SetStatus(codes.Error, ErrUserNotFound.Error())
			return ErrUserNotFound
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			span.SetStatus(codes.Error, ErrDuplicatedKey.Error())
			return ErrDuplicatedKey
		}
		logger.Errorf("Failed to update user: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return err

	}
	if rowsAffected <= 0 {
		span.SetStatus(codes.Error, ErrUserNotFound.Error())
		return ErrUserNotFound
	}

	span.SetStatus(codes.Ok, "user updated")
	return nil

}
func (u *UserRepository) DeleteUser(ctx context.Context, id uint) error {
	ctx, span := u.tracer.Start(ctx, "UserRepository.DeleteUser")
	defer span.End()

	span.SetAttributes(attribute.Int("user.id", int(id)))

	rowsAffected, err := gorm.G[entity.User](u.db).Where("id=?", id).Delete(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			span.SetStatus(codes.Error, ErrUserNotFound.Error())
			return ErrUserNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err

	}
	if rowsAffected <= 0 {
		span.SetStatus(codes.Error, ErrUserNotFound.Error())
		return ErrUserNotFound
	}

	span.SetStatus(codes.Ok, "user deleted")
	return nil
}
