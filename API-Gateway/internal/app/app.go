package app

import (
	"apigateway/internal/domain/models"
	"context"
	"log/slog"

	"github.com/google/uuid"
)

type IUserStorage interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	GetUserById(ctx context.Context, uid uuid.UUID) (models.User, error)
	Insert(ctx context.Context, user models.User) (models.User, error)
	Update(ctx context.Context, uid uuid.UUID, user models.User) (models.User, error)
	Delete(ctx context.Context, uid uuid.UUID) (models.User, error)
}

type App struct {
	log     *slog.Logger
	port    int
	storage IUserStorage
}

func New(log *slog.Logger, port int, storage IUserStorage) *App {
	return &App{
		log:     log,
		port:    port,
		storage: storage,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	return nil
}
