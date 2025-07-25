package app

import (
	"apigateway/internal/domain/models"
	usershandlers "apigateway/internal/handlers/users"
	usersservice "apigateway/internal/service/users"
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
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
	r := mux.NewRouter()

	usersService := usersservice.New(a.log, a.storage)
	usersHandler := usershandlers.New(a.log, usersService)

	r.HandleFunc("/api/v1/login", nil).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/register", nil).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/refresh", nil).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/logout", nil).Methods(http.MethodPost)

	r.HandleFunc("/api/v1/users", usersHandler.GetUsersHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/users/{id}", usersHandler.GetUserByIdHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/users", usersHandler.InsertHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/users/{id}", usersHandler.UpdateHandler).Methods(http.MethodPut)
	r.HandleFunc("/api/v1/users/{id}", usersHandler.DeleteHandler).Methods(http.MethodDelete)

	if err := http.ListenAndServe(
		fmt.Sprintf(":%d", a.port),
		r,
	); err != nil {
		panic(err)
	}

	return nil
}
