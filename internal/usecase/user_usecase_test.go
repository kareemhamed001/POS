package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/kareemhamed001/POS/internal/entity"
)

type mockUserRepository struct {
	listUsersFn      func(ctx context.Context, search string, page, perPage int) (*[]entity.User, int, error) //int for total users for pagination
	createUserFn     func(ctx context.Context, user *entity.User) error
	getUserByIDFn    func(ctx context.Context, id uint) (*entity.User, error)
	getUserByEmailFn func(ctx context.Context, email string) (*entity.User, error)
	updateUserFn     func(ctx context.Context, id uint, user *entity.User) error
	deleteUserFn     func(ctx context.Context, id uint) error
}

func (m *mockUserRepository) ListUsers(ctx context.Context, search string, page, perPage int) (*[]entity.User, int, error) {
	return m.listUsersFn(ctx, search, page, perPage)
}

func (m *mockUserRepository) CreateUser(ctx context.Context, user *entity.User) error {
	return m.createUserFn(ctx, user)
}

func (m *mockUserRepository) GetUserByID(ctx context.Context, id uint) (*entity.User, error) {
	return m.getUserByIDFn(ctx, id)
}

func (m *mockUserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	return m.getUserByEmailFn(ctx, email)
}

func (m *mockUserRepository) UpdateUser(ctx context.Context, id uint, user *entity.User) error {
	return m.updateUserFn(ctx, id, user)
}

func (m *mockUserRepository) DeleteUser(ctx context.Context, id uint) error {
	return m.deleteUserFn(ctx, id)
}

func TestUserUsecase_CreateUser(t *testing.T) {
	mockRepo := &mockUserRepository{}
	usecase := NewUserUsecase(mockRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &entity.User{Name: "Test User"}
		mockRepo.createUserFn = func(ctx context.Context, u *entity.User) error {
			u.ID = 1
			return nil
		}

		createdUser, err := usecase.CreateUser(ctx, user)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if createdUser.ID != 1 {
			t.Fatalf("expected user ID 1, got %d", createdUser.ID)
		}
	})
	t.Run("failure", func(t *testing.T) {
		user := &entity.User{Name: "Test User"}
		expectedErr := errors.New("db error")
		mockRepo.createUserFn = func(ctx context.Context, u *entity.User) error {
			return expectedErr
		}

		_, err := usecase.CreateUser(ctx, user)
		if err != expectedErr {
			t.Fatalf("expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("Bcrypt Error", func(t *testing.T) {
		// Bcrypt fails if the password is too long (> 72 bytes)
		longPass := make([]byte, 80)
		_, err := usecase.CreateUser(ctx, &entity.User{Password: string(longPass)})
		if err == nil {
			t.Errorf("expected bcrypt error for long password")
		}
	})
}

func TestUserUsecase_GetUserByID(t *testing.T) {
	mockRepo := &mockUserRepository{}
	usecase := NewUserUsecase(mockRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expectedUser := &entity.User{Name: "Kareem"}
		mockRepo.getUserByIDFn = func(ctx context.Context, id uint) (*entity.User, error) {
			return expectedUser, nil
		}

		user, err := usecase.GetUserByID(ctx, 1)
		if err != nil || user.Name != "Kareem" {
			t.Errorf("expected Kareem, got %v", user)
		}
	})
}

func TestUserUsecase_UpdateUser(t *testing.T) {
	mockRepo := &mockUserRepository{}
	uc := NewUserUsecase(mockRepo)
	ctx := context.Background()

	t.Run("UpdateUser With Password Hash", func(t *testing.T) {
		mockRepo.updateUserFn = func(ctx context.Context, id uint, user *entity.User) error { return nil }
		err := uc.UpdateUser(ctx, 1, &entity.User{Password: "new-pass"})
		if err != nil {
			t.Errorf("failed update with pass")
		}
	})

	t.Run("UpdateUser Without Password", func(t *testing.T) {
		mockRepo.updateUserFn = func(ctx context.Context, id uint, user *entity.User) error { return nil }
		err := uc.UpdateUser(ctx, 1, &entity.User{Password: ""}) // hits the 'if password != ""' else branch
		if err != nil {
			t.Errorf("failed update without pass")
		}
	})

	t.Run("UpdateUser Bcrypt Error", func(t *testing.T) {
		longPass := make([]byte, 80)
		err := uc.UpdateUser(ctx, 1, &entity.User{Password: string(longPass)})
		if err == nil {
			t.Errorf("expected bcrypt error")
		}
	})

	t.Run("UpdateUser Repo Error", func(t *testing.T) {
		mockRepo.updateUserFn = func(ctx context.Context, id uint, user *entity.User) error {
			return errors.New("db update error")
		}
		err := uc.UpdateUser(ctx, 1, &entity.User{Password: "123"})
		if err == nil {
			t.Errorf("expected error branch for repo failure")
		}
	})

}

func TestUserUsecase_ListUsers(t *testing.T) {
	mockRepo := &mockUserRepository{}
	usecase := NewUserUsecase(mockRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo.listUsersFn = func(ctx context.Context, search string, page, perPage int) (*[]entity.User, int, error) {
			return &[]entity.User{{Name: "User1"}}, 1, nil
		}

		users, total, err := usecase.ListUsers(ctx, "", 1, 10)
		if err != nil || total != 1 || len(*users) != 1 {
			t.Errorf("list failed: %v", err)
		}
	})
}

func TestUserUsecase_DeleteUser(t *testing.T) {
	mockRepo := &mockUserRepository{}
	usecase := NewUserUsecase(mockRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo.deleteUserFn = func(ctx context.Context, id uint) error {
			return nil
		}
		err := usecase.DeleteUser(ctx, 1)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})
}
