package usershandlers

import (
	"apigateway/internal/domain/models"
	serviceerrors "apigateway/internal/service"
	"apigateway/pkg/lib/logger/sl"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type IUsersService interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	GetUserById(ctx context.Context, uid uuid.UUID) (models.User, error)
	Insert(ctx context.Context, user models.User) (models.User, error)
	Update(ctx context.Context, uid uuid.UUID, user models.User) (models.User, error)
	Delete(ctx context.Context, uid uuid.UUID) (models.User, error)
}

type UsersHandler struct {
	log     *slog.Logger
	service IUsersService
}

func New(log *slog.Logger, service IUsersService) *UsersHandler {
	return &UsersHandler{
		log:     log,
		service: service,
	}
}
func (u *UsersHandler) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.users.GetUsersHandler"
	log := u.log.With("op", op)

	select {
	case <-r.Context().Done():
		log.Info("Request cancelled", sl.Err(r.Context().Err()))
		http.Error(w, "Request timeout", http.StatusRequestTimeout)
		return
	default:
	}

	users, err := u.service.GetUsers(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, serviceerrors.ErrContextCanceled):
			log.Warn("Context cancelled", sl.Err(err))
			http.Error(w, "Request timeout", http.StatusRequestTimeout)
			return
		default:
			log.Error("Failed to fetch users", sl.Err(err))
			http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.Error("Failed to encode users", sl.Err(err))
		http.Error(w, "Failed to encode users", http.StatusInternalServerError)
		return
	}
}

func (u *UsersHandler) GetUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.users.GetUserByIdHandler"
	log := u.log.With("op", op)

	select {
	case <-r.Context().Done():
		log.Info("Request cancelled", sl.Err(r.Context().Err()))
		http.Error(w, "Request timeout", http.StatusRequestTimeout)
		return
	default:
	}

	uid, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		log.Error("Invalid user ID", sl.Err(err))
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	user, err := u.service.GetUserById(r.Context(), uid)
	if err != nil {
		switch {
		case errors.Is(err, serviceerrors.ErrContextCanceled):
			log.Warn("Request cancelled", sl.Err(err))
			http.Error(w, "Request timeout", http.StatusRequestTimeout)
			return
		case errors.Is(err, serviceerrors.ErrInvalidArgument):
			log.Warn("Invalid argument", sl.Err(err))
			http.Error(w, "Invalid argument", http.StatusBadRequest)
			return
		case errors.Is(err, serviceerrors.ErrNotFound):
			log.Warn("User not found", sl.Err(err), slog.String("user_id", uid.String()))
			http.Error(w, "User not found", http.StatusNotFound)
			return
		default:
			log.Error("Failed to fetch user by id", sl.Err(err), slog.String("user_id", uid.String()))
			http.Error(w, "Failed to fetch user by id", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Error("Failed to encode user", sl.Err(err))
		http.Error(w, "Failed to encode user", http.StatusInternalServerError)
		return
	}
}

func (u *UsersHandler) InsertHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.users.InsertHandler"
	log := u.log.With("op", op)

	select {
	case <-r.Context().Done():
		log.Info("Request cancelled", sl.Err(r.Context().Err()))
		http.Error(w, "Request timeout", http.StatusRequestTimeout)
		return
	default:
	}

	validate := validator.New()
	var userFromRequest models.User
	if err := json.NewDecoder(r.Body).Decode(&userFromRequest); err != nil {
		log.Error("Failed to read request body", sl.Err(err))
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(userFromRequest); err != nil {
		log.Error("Failed to validate requested user", sl.Err(err))
		http.Error(w, "Failed to validate user", http.StatusBadRequest)
		return
	}

	insertedUser, err := u.service.Insert(r.Context(), userFromRequest)
	if err != nil {
		switch {
		case errors.Is(err, serviceerrors.ErrContextCanceled):
			log.Warn("Request cancelled", sl.Err(err))
			http.Error(w, "Request timeout", http.StatusRequestTimeout)
			return
		case errors.Is(err, serviceerrors.ErrInvalidArgument):
			log.Warn("Invalid argument", sl.Err(err))
			http.Error(w, "Invalid argument", http.StatusBadRequest)
			return
		case errors.Is(err, serviceerrors.ErrAlreadyExists):
			log.Warn("User already exists", sl.Err(err))
			http.Error(w, "User already exists", http.StatusConflict)
			return
		default:
			log.Error("Failed to insert user", sl.Err(err))
			http.Error(w, "Failed to insert user", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(insertedUser); err != nil {
		log.Error("Failed to encode user", sl.Err(err))
		http.Error(w, "Failed to encode user", http.StatusInternalServerError)
		return
	}
}

func (u *UsersHandler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.users.UpdateHandler"
	log := u.log.With("op", op)

	select {
	case <-r.Context().Done():
		log.Info("Request cancelled", sl.Err(r.Context().Err()))
		http.Error(w, "Request timeout", http.StatusRequestTimeout)
		return
	default:
	}

	uid, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		log.Error("Invalid user ID", sl.Err(err))
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	validate := validator.New()
	var userFromRequest models.User
	if err := json.NewDecoder(r.Body).Decode(&userFromRequest); err != nil {
		log.Error("Failed to read request body", sl.Err(err))
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(userFromRequest); err != nil {
		log.Error("Failed to validate requested user", sl.Err(err))
		http.Error(w, "Failed to validate user", http.StatusBadRequest)
		return
	}

	updatedUser, err := u.service.Update(r.Context(), uid, userFromRequest)
	if err != nil {
		switch {
		case errors.Is(err, serviceerrors.ErrContextCanceled):
			log.Warn("Request cancelled", sl.Err(err))
			http.Error(w, "Request timeout", http.StatusRequestTimeout)
			return
		case errors.Is(err, serviceerrors.ErrInvalidArgument):
			log.Warn("Invalid argument", sl.Err(err))
			http.Error(w, "Invalid argument", http.StatusBadRequest)
			return
		case errors.Is(err, serviceerrors.ErrNotFound):
			log.Warn("User not found", sl.Err(err), slog.String("user_id", uid.String()))
			http.Error(w, "User not found", http.StatusNotFound)
			return
		default:
			log.Error("Failed to update user", sl.Err(err), slog.String("user_id", uid.String()))
			http.Error(w, "Failed to update user", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(updatedUser); err != nil {
		log.Error("Failed to encode user", sl.Err(err))
		http.Error(w, "Failed to encode user", http.StatusInternalServerError)
		return
	}
}

func (u *UsersHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.users.DeleteHandler"
	log := u.log.With("op", op)

	select {
	case <-r.Context().Done():
		log.Info("Request cancelled", sl.Err(r.Context().Err()))
		http.Error(w, "Request timeout", http.StatusRequestTimeout)
		return
	default:
	}

	uid, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		log.Error("Invalid user ID", sl.Err(err))
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	deletedUser, err := u.service.Delete(r.Context(), uid)
	if err != nil {
		switch {
		case errors.Is(err, serviceerrors.ErrContextCanceled):
			log.Warn("Request cancelled", sl.Err(err))
			http.Error(w, "Request timeout", http.StatusRequestTimeout)
			return
		case errors.Is(err, serviceerrors.ErrInvalidArgument):
			log.Warn("Invalid argument", sl.Err(err))
			http.Error(w, "Invalid argument", http.StatusBadRequest)
			return
		case errors.Is(err, serviceerrors.ErrNotFound):
			log.Warn("User not found", sl.Err(err), slog.String("user_id", uid.String()))
			http.Error(w, "User not found", http.StatusNotFound)
			return
		default:
			log.Error("Failed to delete user", sl.Err(err), slog.String("user_id", uid.String()))
			http.Error(w, "Failed to delete user", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(deletedUser); err != nil {
		log.Error("Failed to encode user", sl.Err(err))
		http.Error(w, "Failed to encode user", http.StatusInternalServerError)
		return
	}
}
