package grpcapp

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"usersmanager/internal/domain/models"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

type IUsersService interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	GetUserById(ctx context.Context, uid uuid.UUID) (models.User, error)
	Insert(ctx context.Context, user models.User) (models.User, error)
	Update(ctx context.Context, uid uuid.UUID, user models.User) (models.User, error)
	Delete(ctx context.Context, uid uuid.UUID) (models.User, error)
}

func New(log *slog.Logger, usersService IUsersService, port int) *App {
	gRPCServer := grpc.NewServer()

	// usersgrpc.Register(gRPCServer, usersService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"
	log := a.log.With("op", op)

	l, err := net.Listen(
		"tcp",
		fmt.Sprintf(":%d", a.port),
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Starting grpc server")

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	a.gRPCServer.GracefulStop()
}
