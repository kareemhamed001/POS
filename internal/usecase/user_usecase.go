package usecase

import (
	"context"
	"log"

	"github.com/kareemhamed001/POS/internal/entity"
	"github.com/kareemhamed001/POS/pkg/password"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	userRepository UserRepository
	tracer         trace.Tracer
}

func NewUserUsecase(userRepository UserRepository) *UserUsecase {
	return &UserUsecase{
		userRepository: userRepository,
		tracer:         otel.Tracer("user-usecase"),
	}
}

func (u *UserUsecase) ListUsers(ctx context.Context, search string, page, perPage int) (*[]entity.User, int, error) {
	ctx, span := u.tracer.Start(ctx, "UserUsecase.ListUsers")
	defer span.End()

	span.SetAttributes(
		attribute.String("query.search", search),
		attribute.Int("query.page", page),
		attribute.Int("query.per_page", perPage),
	)

	users, total, err := u.userRepository.ListUsers(ctx, search, page, perPage)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, 0, err
	}

	span.SetAttributes(attribute.Int("users.count", len(*users)))
	span.SetStatus(codes.Ok, "users retrieved")
	return users, total, nil
}

func (u *UserUsecase) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	ctx, span := u.tracer.Start(ctx, "UserUsecase.CreateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.email", user.Email),
		attribute.String("user.phone", user.Phone),
	)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	user.Password = string(hashedPassword)

	if err := u.userRepository.CreateUser(ctx, user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("user.id", int(user.ID)))
	span.SetStatus(codes.Ok, "user created")
	return user, nil
}
func (u *UserUsecase) GetUserByID(ctx context.Context, id uint) (*entity.User, error) {
	ctx, span := u.tracer.Start(ctx, "UserUsecase.GetUserByID")
	defer span.End()

	span.SetAttributes(attribute.Int("user.id", int(id)))

	user, err := u.userRepository.GetUserByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.String("user.email", user.Email))
	span.SetStatus(codes.Ok, "user retrieved")
	return user, nil
}
func (u *UserUsecase) UpdateUser(ctx context.Context, id uint, user *entity.User) error {
	ctx, span := u.tracer.Start(ctx, "UserUsecase.UpdateUser")
	defer span.End()

	span.SetAttributes(attribute.Int("user.id", int(id)))

	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
		if err != nil {
			log.Println("Error hashing password:", err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		user.Password = string(hashedPassword)
	}
	err := u.userRepository.UpdateUser(ctx, id, user)
	if err != nil {
		log.Println("Error updating user:", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	if err == nil {
		span.SetStatus(codes.Ok, "user updated")
	}
	return err
}
func (u *UserUsecase) DeleteUser(ctx context.Context, id uint) error {
	ctx, span := u.tracer.Start(ctx, "UserUsecase.DeleteUser")
	defer span.End()

	span.SetAttributes(attribute.Int("user.id", int(id)))

	if err := u.userRepository.DeleteUser(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "user deleted")
	return nil
}

// Register creates a new user account
func (u *UserUsecase) Register(ctx context.Context, user *entity.User) (*entity.User, error) {
	ctx, span := u.tracer.Start(ctx, "UserUsecase.Register")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.email", user.Email),
		attribute.String("user.phone", user.Phone),
	)

	hashedPassword, err := password.Hash(user.Password)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	user.Password = hashedPassword
	if user.Role == "" {
		user.Role = entity.RoleCustomer
	}

	if err := u.userRepository.CreateUser(ctx, user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("user.id", int(user.ID)))
	span.SetStatus(codes.Ok, "user registered")
	return user, nil
}

// Login verifies user credentials and returns the user if valid
func (u *UserUsecase) Login(ctx context.Context, email, pwd string) (*entity.User, error) {
	ctx, span := u.tracer.Start(ctx, "UserUsecase.Login")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	user, err := u.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if !password.Verify(user.Password, pwd) {
		err := bcrypt.ErrMismatchedHashAndPassword
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, bcrypt.ErrMismatchedHashAndPassword
	}

	span.SetAttributes(attribute.Int("user.id", int(user.ID)))
	span.SetStatus(codes.Ok, "user authenticated")
	return user, nil
}
