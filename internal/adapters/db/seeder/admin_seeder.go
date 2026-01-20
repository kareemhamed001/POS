package seeder

import (
	"context"
	"errors"

	entity "github.com/kareemhamed001/POS/internal/core/domain"
	"github.com/kareemhamed001/POS/internal/adapters/persistence/postgres"
	"github.com/kareemhamed001/POS/pkg/logger"
	"github.com/kareemhamed001/POS/pkg/password"
)

// SeedAdmin ensures a default admin user exists; creates one if missing.
func SeedAdmin(ctx context.Context, repo *postgres.UserRepository, name, email, phone, plainPassword string) error {
	if email == "" || plainPassword == "" {
		logger.Warn("Skipping admin seeding: ADMIN_EMAIL or ADMIN_PASSWORD is empty")
		return nil
	}

	// Check if admin already exists
	_, err := repo.GetUserByEmail(ctx, email)
	if err == nil {
		logger.Info("Admin user already exists, skipping seeding")
		return nil
	}
	if !errors.Is(err, postgres.ErrUserNotFound) {
		return err
	}

	hashed, err := password.Hash(plainPassword)
	if err != nil {
		return err
	}

	admin := &entity.User{
		Name:     name,
		Email:    email,
		Phone:    phone,
		Password: hashed,
		Role:     entity.RoleAdmin,
	}

	if err := repo.CreateUser(ctx, admin); err != nil {
		// If another process created it concurrently, ignore duplicate error
		if err == postgres.ErrDuplicatedKey {
			logger.Info("Admin user created by another process, skipping")
			return nil
		}
		return err
	}

	logger.Info("Seeded default admin user")
	return nil
}
