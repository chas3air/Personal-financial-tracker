package app

import (
	"context"
	"log/slog"
	grpcapp "usersmanager/internal/app/grpc"
	"usersmanager/internal/domain/models"

	"github.com/google/uuid"
)

type App struct {
	GRPCApp *grpcapp.App
}

type IUsersStorage interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	GetUserById(ctx context.Context, uid uuid.UUID) (models.User, error)
	Insert(ctx context.Context, user models.User) (models.User, error)
	Update(ctx context.Context, uid uuid.UUID, user models.User) (models.User, error)
	Delete(ctx context.Context, uid uuid.UUID) (models.User, error)
}

func New(log *slog.Logger, port int, usersStorage IUsersStorage) *App {
	usersService := usersservice.New(log, usersStorage)
	grpcApp := grpcapp.New(log, usersService, port)

	return &App{
		GRPCApp: grpcApp,
	}
}
