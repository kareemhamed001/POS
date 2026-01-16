package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kareemhamed001/POS/internal/entity"
	"github.com/kareemhamed001/POS/pkg/logger"
	"gorm.io/gorm"
)

var (
	ErrDuplicatedKey = errors.New("User with same email or phone already exists")
	ErrUserNotFound  = errors.New("User not found")
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (u *UserRepository) ListUsers(ctx context.Context, search string, page, perPage int) (*[]entity.User, int, error) {
	// Use LOWER for case-insensitive search (works with both PostgreSQL and SQLite)
	searchPattern := "%" + search + "%"
	total, err := gorm.G[entity.User](u.db).
		Where("LOWER(name) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?) OR LOWER(phone) LIKE LOWER(?)", searchPattern, searchPattern, searchPattern).
		Count(ctx, "id")
	if err != nil {
		return nil, 0, err
	}
	skip := (page - 1) * perPage
	users, err := gorm.G[entity.User](u.db).
		Where("LOWER(name) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?) OR LOWER(phone) LIKE LOWER(?)", searchPattern, searchPattern, searchPattern).
		Offset(skip).
		Limit(perPage).
		Find(ctx)
	if err != nil {
		return nil, 0, err
	}

	return &users, int(total), nil
}
func (u *UserRepository) CreateUser(ctx context.Context, user *entity.User) error {
	err := gorm.G[entity.User](u.db).Create(ctx, user)
	if err != nil {
		// Check for PostgreSQL unique constraint violation
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation
				return ErrDuplicatedKey
			}
		}
		// Check for SQLite unique constraint violation
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrDuplicatedKey
		}
		return err
	}
	return nil
}
func (u *UserRepository) GetUserByID(ctx context.Context, id uint) (*entity.User, error) {

	user, err := gorm.G[entity.User](u.db).Where("id=?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err

	}
	return &user, nil
}
func (u *UserRepository) UpdateUser(ctx context.Context, id uint, user *entity.User) error {
	rowsAffected, err := gorm.G[*entity.User](u.db).Where("id=?", id).Updates(ctx, user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrDuplicatedKey
		}
		logger.Errorf("Failed to update user: %v", err)

		return err

	}
	if rowsAffected <= 0 {
		return ErrUserNotFound
	}
	return nil

}
func (u *UserRepository) DeleteUser(ctx context.Context, id uint) error {
	rowsAffected, err := gorm.G[entity.User](u.db).Where("id=?", id).Delete(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err

	}
	if rowsAffected <= 0 {
		return ErrUserNotFound
	}
	return nil
}
