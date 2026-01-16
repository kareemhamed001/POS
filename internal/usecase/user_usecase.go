package usecase

import (
	"context"
	"log"

	"github.com/kareemhamed001/POS/internal/entity"
	"github.com/kareemhamed001/POS/pkg/password"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	userRepository UserRepository
}

func NewUserUsecase(userRepository UserRepository) *UserUsecase {
	return &UserUsecase{
		userRepository: userRepository,
	}
}

func (u *UserUsecase) ListUsers(ctx context.Context, search string, page, perPage int) (*[]entity.User, int, error) {
	return u.userRepository.ListUsers(ctx, search, page, perPage)
}

func (u *UserUsecase) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		return nil, err
	}

	user.Password = string(hashedPassword)

	if err := u.userRepository.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}
func (u *UserUsecase) GetUserByID(ctx context.Context, id uint) (*entity.User, error) {
	return u.userRepository.GetUserByID(ctx, id)
}
func (u *UserUsecase) UpdateUser(ctx context.Context, id uint, user *entity.User) error {
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
		if err != nil {
			log.Println("Error hashing password:", err)
			return err
		}
		user.Password = string(hashedPassword)
	}
	err := u.userRepository.UpdateUser(ctx, id, user)
	if err != nil {
		log.Println("Error updating user:", err)
	}
	return err
}
func (u *UserUsecase) DeleteUser(ctx context.Context, id uint) error {
	return u.userRepository.DeleteUser(ctx, id)
}

// Register creates a new user account
func (u *UserUsecase) Register(ctx context.Context, user *entity.User) (*entity.User, error) {
	hashedPassword, err := password.Hash(user.Password)
	if err != nil {
		return nil, err
	}

	user.Password = hashedPassword
	if user.Role == "" {
		user.Role = entity.RoleCustomer
	}

	if err := u.userRepository.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// Login verifies user credentials and returns the user if valid
func (u *UserUsecase) Login(ctx context.Context, email, pwd string) (*entity.User, error) {
	// Get user by email
	user, err := u.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Verify password
	if !password.Verify(user.Password, pwd) {
		return nil, bcrypt.ErrMismatchedHashAndPassword
	}

	return user, nil
}
