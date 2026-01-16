package postgres

import (
	"context"
	"testing"

	"github.com/kareemhamed001/POS/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "failed to connect to test database")

	// Migrate the schema
	err = db.AutoMigrate(&entity.User{}, &entity.Address{}, &entity.Order{})
	require.NoError(t, err, "failed to migrate test database")

	return db
}

func createTestUser(t *testing.T, db *gorm.DB, email, phone string) *entity.User {
	user := &entity.User{
		Name:     "Test User",
		Email:    email,
		Phone:    phone,
		Password: "password123",
		Role:     entity.RoleCustomer,
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

func TestUserRepository_CreateUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("successfully create user", func(t *testing.T) {
		user := &entity.User{
			Name:     "John Doe",
			Email:    "john@example.com",
			Phone:    "+201234567890",
			Password: "securepass",
			Role:     entity.RoleCustomer,
		}

		err := repo.CreateUser(ctx, user)
		assert.NoError(t, err)
		assert.NotZero(t, user.ID)
	})

	t.Run("fail on duplicate email", func(t *testing.T) {
		user1 := &entity.User{
			Name:     "User One",
			Email:    "duplicate@example.com",
			Phone:    "+201111111111",
			Password: "pass1",
			Role:     entity.RoleCustomer,
		}
		err := repo.CreateUser(ctx, user1)
		require.NoError(t, err)

		user2 := &entity.User{
			Name:     "User Two",
			Email:    "duplicate@example.com",
			Phone:    "+202222222222",
			Password: "pass2",
			Role:     entity.RoleCustomer,
		}
		err = repo.CreateUser(ctx, user2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email")
	})

	t.Run("fail on duplicate phone", func(t *testing.T) {
		user1 := &entity.User{
			Name:     "User Three",
			Email:    "user3@example.com",
			Phone:    "+203333333333",
			Password: "pass3",
			Role:     entity.RoleCustomer,
		}
		err := repo.CreateUser(ctx, user1)
		require.NoError(t, err)

		user2 := &entity.User{
			Name:     "User Four",
			Email:    "user4@example.com",
			Phone:    "+203333333333",
			Password: "pass4",
			Role:     entity.RoleCustomer,
		}
		err = repo.CreateUser(ctx, user2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "phone")
	})
}

func TestUserRepository_GetUserByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("successfully get user by id", func(t *testing.T) {
		testUser := createTestUser(t, db, "getuser@example.com", "+204444444444")

		user, err := repo.GetUserByID(ctx, testUser.ID)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.Email, user.Email)
		assert.Equal(t, testUser.Phone, user.Phone)
	})

	t.Run("return error when user not found", func(t *testing.T) {
		user, err := repo.GetUserByID(ctx, 99999)
		assert.ErrorIs(t, err, ErrUserNotFound)
		assert.Nil(t, user)
	})
}

func TestUserRepository_ListUsers(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create test users
	createTestUser(t, db, "alice@example.com", "+205555555555")
	createTestUser(t, db, "bob@example.com", "+206666666666")
	createTestUser(t, db, "charlie@example.com", "+207777777777")

	t.Run("list all users without search", func(t *testing.T) {
		users, total, err := repo.ListUsers(ctx, "", 1, 10)
		assert.NoError(t, err)
		assert.NotNil(t, users)
		assert.Equal(t, 3, total)
		assert.Len(t, *users, 3)
	})

	t.Run("search users by name", func(t *testing.T) {
		_, total, err := repo.ListUsers(ctx, "Test User", 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, 3, total) // All users have "Test User" as name
	})

	t.Run("search users by email", func(t *testing.T) {
		_, total, err := repo.ListUsers(ctx, "alice", 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, 1, total) // Only alice@example.com matches
	})

	t.Run("pagination works correctly", func(t *testing.T) {
		// Page 1 with 2 items per page
		users, total, err := repo.ListUsers(ctx, "", 1, 2)
		assert.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, *users, 2)

		// Page 2 with 2 items per page
		users, total, err = repo.ListUsers(ctx, "", 2, 2)
		assert.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, *users, 1)
	})
}

func TestUserRepository_UpdateUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("successfully update user", func(t *testing.T) {
		testUser := createTestUser(t, db, "update@example.com", "+208888888888")

		updateData := &entity.User{
			Name:  "Updated Name",
			Email: "updated@example.com",
		}

		err := repo.UpdateUser(ctx, testUser.ID, updateData)
		assert.NoError(t, err)

		// Verify the update
		updatedUser, err := repo.GetUserByID(ctx, testUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", updatedUser.Name)
		assert.Equal(t, "updated@example.com", updatedUser.Email)
	})

	t.Run("return error when updating non-existent user", func(t *testing.T) {
		updateData := &entity.User{
			Name: "Non Existent",
		}

		err := repo.UpdateUser(ctx, 99999, updateData)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("partial update should work", func(t *testing.T) {
		testUser := createTestUser(t, db, "partial@example.com", "+209999999999")
		originalEmail := testUser.Email

		updateData := &entity.User{
			Name: "Only Name Updated",
		}

		err := repo.UpdateUser(ctx, testUser.ID, updateData)
		assert.NoError(t, err)

		updatedUser, err := repo.GetUserByID(ctx, testUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Only Name Updated", updatedUser.Name)
		assert.Equal(t, originalEmail, updatedUser.Email)
	})
}

func TestUserRepository_DeleteUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("successfully delete user", func(t *testing.T) {
		testUser := createTestUser(t, db, "delete@example.com", "+201000000000")

		err := repo.DeleteUser(ctx, testUser.ID)
		assert.NoError(t, err)

		// Verify deletion
		user, err := repo.GetUserByID(ctx, testUser.ID)
		assert.ErrorIs(t, err, ErrUserNotFound)
		assert.Nil(t, user)
	})

	t.Run("return error when deleting non-existent user", func(t *testing.T) {
		err := repo.DeleteUser(ctx, 99999)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("soft delete should work", func(t *testing.T) {
		testUser := createTestUser(t, db, "softdelete@example.com", "+201111111110")

		err := repo.DeleteUser(ctx, testUser.ID)
		assert.NoError(t, err)

		// Check that the record exists but is soft deleted
		var deletedUser entity.User
		err = db.Unscoped().First(&deletedUser, testUser.ID).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedUser.DeletedAt)
	})
}

func TestUserRepository_ConcurrentOperations(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("concurrent creates should handle duplicates", func(t *testing.T) {
		done := make(chan bool, 2)
		errors := make(chan error, 2)

		createUser := func(email string) {
			user := &entity.User{
				Name:     "Concurrent User",
				Email:    email,
				Phone:    "+201234567899",
				Password: "pass",
				Role:     entity.RoleCustomer,
			}
			err := repo.CreateUser(ctx, user)
			errors <- err
			done <- true
		}

		go createUser("concurrent@example.com")
		go createUser("concurrent@example.com")

		<-done
		<-done

		err1 := <-errors
		err2 := <-errors

		// One should succeed, one should fail with duplicate key
		successCount := 0
		errorCount := 0

		if err1 == nil {
			successCount++
		} else {
			errorCount++
		}

		if err2 == nil {
			successCount++
		} else {
			errorCount++
		}

		assert.Equal(t, 1, successCount, "exactly one create should succeed")
		assert.Equal(t, 1, errorCount, "exactly one create should fail")
	})
}
