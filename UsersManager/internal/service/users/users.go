package usersservice

import (
	"context"
	"log/slog"
	"usersmanager/internal/domain/models"

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
	panic("unimplemented")
}

// GetUserById implements grpcapp.IUsersService.
func (u *UsersService) GetUserById(ctx context.Context, uid uuid.UUID) (models.User, error) {
	panic("unimplemented")
}

// Insert implements grpcapp.IUsersService.
func (u *UsersService) Insert(ctx context.Context, user models.User) (models.User, error) {
	panic("unimplemented")
}

// Update implements grpcapp.IUsersService.
func (u *UsersService) Update(ctx context.Context, uid uuid.UUID, user models.User) (models.User, error) {
	panic("unimplemented")
}

// Delete implements grpcapp.IUsersService.
func (u *UsersService) Delete(ctx context.Context, uid uuid.UUID) (models.User, error) {
	panic("unimplemented")
}
