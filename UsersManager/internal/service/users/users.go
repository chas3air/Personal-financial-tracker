package usersservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"usersmanager/internal/domain/models"
	serviceerrors "usersmanager/internal/service"
	storageerrors "usersmanager/internal/storage"
	"usersmanager/pkg/lib/logger/sl"

	"github.com/google/uuid"
)

type IUsersStorage interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	GetUserById(ctx context.Context, uid uuid.UUID) (models.User, error)
	Insert(ctx context.Context, user models.User) (models.User, error)
	Update(ctx context.Context, uid uuid.UUID, user models.User) (models.User, error)
	Delete(ctx context.Context, uid uuid.UUID) (models.User, error)
}

type UsersService struct {
	log     *slog.Logger
	storage IUsersStorage
}

func New(log *slog.Logger, storage IUsersStorage) *UsersService {
	return &UsersService{
		log:     log,
		storage: storage,
	}
}

// GetUsers implements grpcapp.IUsersService.
func (u *UsersService) GetUsers(ctx context.Context) ([]models.User, error) {
	const op = "service.users.GetUsers"
	log := u.log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return nil, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	users, err := u.storage.GetUsers(ctx)
	if err != nil {
		log.Error("Failed to fetch users", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Users fetched successfully", slog.Int("count", len(users)))
	return users, nil
}

// GetUserById implements grpcapp.IUsersService.
func (u *UsersService) GetUserById(ctx context.Context, uid uuid.UUID) (models.User, error) {
	const op = "service.users.GetUserById"
	log := u.log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	user, err := u.storage.GetUserById(ctx, uid)
	if err != nil {
		if errors.Is(err, storageerrors.ErrNotFound) {
			log.Warn("User not found", sl.Err(storageerrors.ErrNotFound), slog.String("user_id", uid.String()))
			return models.User{}, fmt.Errorf("%s: %w", op, serviceerrors.ErrNotFound)
		}

		log.Error("Failed to fetch user by id", sl.Err(err), slog.String("user_id", uid.String()))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("User fetched successfully", slog.String("user_id", user.Id.String()))
	return user, nil
}

// Insert implements grpcapp.IUsersService.
func (u *UsersService) Insert(ctx context.Context, userForInsert models.User) (models.User, error) {
	const op = "service.users.Insert"
	log := u.log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	insertedUser, err := u.storage.Insert(ctx, userForInsert)
	if err != nil {
		if errors.Is(err, storageerrors.ErrAlreadyExists) {
			log.Warn("User already exists", sl.Err(storageerrors.ErrAlreadyExists), slog.String("user_id", userForInsert.Id.String()))
			return models.User{}, fmt.Errorf("%s: %w", op, serviceerrors.ErrAlreadyExists)
		}

		log.Error("Failed to insert user", sl.Err(err), slog.String("user_id", userForInsert.Id.String()))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("User inserted successfully", slog.String("user_id", insertedUser.Id.String()))
	return insertedUser, nil
}

// Update implements grpcapp.IUsersService.
func (u *UsersService) Update(ctx context.Context, uid uuid.UUID, userForUpdate models.User) (models.User, error) {
	const op = "service.users.Update"
	log := u.log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	updatedUser, err := u.storage.Update(ctx, uid, userForUpdate)
	if err != nil {
		if errors.Is(err, storageerrors.ErrNotFound) {
			log.Warn("User not found for update", sl.Err(storageerrors.ErrNotFound), slog.String("user_id", uid.String()))
			return models.User{}, fmt.Errorf("%s: %w", op, serviceerrors.ErrNotFound)
		}

		log.Error("Failed to update user", sl.Err(err), slog.String("user_id", uid.String()))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("User updated successfully", slog.String("user_id", updatedUser.Id.String()))
	return updatedUser, nil
}

// Delete implements grpcapp.IUsersService.
func (u *UsersService) Delete(ctx context.Context, uid uuid.UUID) (models.User, error) {
	const op = "service.users.Delete"
	log := u.log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	deletedUser, err := u.storage.Delete(ctx, uid)
	if err != nil {
		if errors.Is(err, storageerrors.ErrNotFound) {
			log.Warn("User not found for deletion", sl.Err(storageerrors.ErrNotFound), slog.String("user_id", uid.String()))
			return models.User{}, fmt.Errorf("%s: %w", op, serviceerrors.ErrNotFound)
		}

		log.Error("Failed to delete user", sl.Err(err), slog.String("user_id", uid.String()))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("User deleted successfully", slog.String("user_id", deletedUser.Id.String()))
	return deletedUser, nil
}
