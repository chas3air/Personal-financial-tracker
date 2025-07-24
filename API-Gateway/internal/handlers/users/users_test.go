package usershandlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"apigateway/internal/domain/models"
	usershandlers "apigateway/internal/handlers/users"
	serviceerrors "apigateway/internal/service"
	"apigateway/pkg/lib/logger/handler/slogdiscard"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Мок сервиса пользователей
type mockUsersService struct {
	mock.Mock
}

func (m *mockUsersService) GetUsers(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *mockUsersService) GetUserById(ctx context.Context, uid uuid.UUID) (models.User, error) {
	args := m.Called(ctx, uid)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *mockUsersService) Insert(ctx context.Context, user models.User) (models.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *mockUsersService) Update(ctx context.Context, uid uuid.UUID, user models.User) (models.User, error) {
	args := m.Called(ctx, uid, user)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *mockUsersService) Delete(ctx context.Context, uid uuid.UUID) (models.User, error) {
	args := m.Called(ctx, uid)
	return args.Get(0).(models.User), args.Error(1)
}

func newTestHandler(t *testing.T) (*usershandlers.UsersHandler, *mockUsersService) {
	mockService := new(mockUsersService)
	logger := slogdiscard.NewDiscardLogger()
	handler := usershandlers.New(logger, mockService)
	return handler, mockService
}

func TestUsersHandler_GetUsersHandler(t *testing.T) {
	handler, service := newTestHandler(t)

	t.Run("success", func(t *testing.T) {
		users := []models.User{
			{Id: uuid.New(), Login: "user1"},
			{Id: uuid.New(), Login: "user2"},
		}
		service.On("GetUsers", mock.Anything).Return(users, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()

		handler.GetUsersHandler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var got []models.User
		err := json.NewDecoder(resp.Body).Decode(&got)
		assert.NoError(t, err)
		assert.Len(t, got, 2)
		service.AssertExpectations(t)
	})

	t.Run("context cancelled error", func(t *testing.T) {
		service.On("GetUsers", mock.Anything).Return(nil, serviceerrors.ErrContextCanceled).Once()

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()

		handler.GetUsersHandler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusRequestTimeout, resp.StatusCode)
		service.AssertExpectations(t)
	})

	t.Run("other error", func(t *testing.T) {
		service.On("GetUsers", mock.Anything).Return(nil, errors.New("some error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()

		handler.GetUsersHandler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		service.AssertExpectations(t)
	})
}

func TestUsersHandler_GetUserByIdHandler(t *testing.T) {
	handler, service := newTestHandler(t)

	validID := uuid.New()
	url := "/users/" + validID.String()

	t.Run("success", func(t *testing.T) {
		user := models.User{Id: validID, Login: "user1"}
		service.On("GetUserById", mock.Anything, validID).Return(user, nil).Once()

		req := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()

		// Используем mux, чтобы получить vars
		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", handler.GetUserByIdHandler)
		router.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var got models.User
		err := json.NewDecoder(resp.Body).Decode(&got)
		assert.NoError(t, err)
		assert.Equal(t, validID, got.Id)
		service.AssertExpectations(t)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/not-uuid", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", handler.GetUserByIdHandler)
		router.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("context cancelled error", func(t *testing.T) {
		service.On("GetUserById", mock.Anything, validID).Return(models.User{}, serviceerrors.ErrContextCanceled).Once()

		req := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", handler.GetUserByIdHandler)
		router.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusRequestTimeout, resp.StatusCode)
		service.AssertExpectations(t)
	})

	t.Run("not found error", func(t *testing.T) {
		service.On("GetUserById", mock.Anything, validID).Return(models.User{}, serviceerrors.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", handler.GetUserByIdHandler)
		router.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		service.AssertExpectations(t)
	})

	t.Run("other error", func(t *testing.T) {
		service.On("GetUserById", mock.Anything, validID).Return(models.User{}, errors.New("other error")).Once()

		req := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", handler.GetUserByIdHandler)
		router.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		service.AssertExpectations(t)
	})
}

func TestUsersHandler_InsertHandler(t *testing.T) {
	handler, service := newTestHandler(t)

	tUser := models.User{Id: uuid.New(), Login: "user1", Password: "pass1", Role: "user"}
	bodyBytes, _ := json.Marshal(tUser)

	t.Run("success", func(t *testing.T) {
		service.On("Insert", mock.Anything, tUser).Return(tUser, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()

		handler.InsertHandler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var got models.User
		err := json.NewDecoder(resp.Body).Decode(&got)
		assert.NoError(t, err)
		assert.Equal(t, tUser.Id, got.Id)
		service.AssertExpectations(t)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader("bad json"))
		w := httptest.NewRecorder()

		handler.InsertHandler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("validation error", func(t *testing.T) {
		// Empty user, validation should fail (assuming your user struct has required fields)
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{}`))
		w := httptest.NewRecorder()

		handler.InsertHandler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("context cancelled error", func(t *testing.T) {
		service.On("Insert", mock.Anything, mock.Anything).Return(models.User{}, serviceerrors.ErrContextCanceled).Once()

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()

		handler.InsertHandler(w, req)

		resp := w.Result()
		fmt.Println(resp.StatusCode)
		assert.Equal(t, http.StatusRequestTimeout, resp.StatusCode)
		service.AssertExpectations(t)
	})

	/*

		t.Run("context cancelled error", func(t *testing.T) {
			service.On("GetUsers", mock.Anything).Return(nil, serviceerrors.ErrContextCanceled).Once()

			req := httptest.NewRequest(http.MethodGet, "/users", nil)
			w := httptest.NewRecorder()

			handler.GetUsersHandler(w, req)

			resp := w.Result()
			assert.Equal(t, http.StatusRequestTimeout, resp.StatusCode)
			service.AssertExpectations(t)
		})
	*/

	t.Run("already exists error", func(t *testing.T) {
		service.On("Insert", mock.Anything, mock.Anything).Return(models.User{}, serviceerrors.ErrAlreadyExists).Once()

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()

		handler.InsertHandler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
		service.AssertExpectations(t)
	})

	t.Run("other error", func(t *testing.T) {
		service.On("Insert", mock.Anything, mock.Anything).Return(models.User{}, errors.New("some error")).Once()

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()

		handler.InsertHandler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		service.AssertExpectations(t)
	})
}

func TestUsersHandler_UpdateHandler(t *testing.T) {
	handler, service := newTestHandler(t)

	validID := uuid.New()
	url := "/users/" + validID.String()
	tUser := models.User{Id: validID, Login: "userUpdated", Password: "passUpdated", Role: "admin"}
	bodyBytes, _ := json.Marshal(tUser)

	t.Run("success", func(t *testing.T) {
		service.On("Update", mock.Anything, validID, tUser).Return(tUser, nil).Once()

		req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", handler.UpdateHandler)
		router.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var got models.User
		err := json.NewDecoder(resp.Body).Decode(&got)
		assert.NoError(t, err)
		assert.Equal(t, validID, got.Id)
		service.AssertExpectations(t)
	})

	t.Run("invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/users/not-uuid", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", handler.UpdateHandler)
		router.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, url, strings.NewReader("not json"))
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", handler.UpdateHandler)
		router.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("validation failure", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, url, strings.NewReader(`{}`))
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", handler.UpdateHandler)
		router.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("context cancelled error", func(t *testing.T) {
		service.On("Update", mock.Anything, validID, mock.Anything).Return(models.User{}, serviceerrors.ErrContextCanceled).Once()

		req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", handler.UpdateHandler)
		router.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusRequestTimeout, resp.StatusCode)
		service.AssertExpectations(t)
	})

	t.Run("not found error", func(t *testing.T) {
		service.On("Update", mock.Anything, validID, mock.Anything).Return(models.User{}, serviceerrors.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", handler.UpdateHandler)
		router.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		service.AssertExpectations(t)
	})

	t.Run("other error", func(t *testing.T) {
		service.On("Update", mock.Anything, validID, mock.Anything).Return(models.User{}, errors.New("other error")).Once()

		req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", handler.UpdateHandler)
		router.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		service.AssertExpectations(t)
	})
}

func TestUsersHandler_DeleteHandler(t *testing.T) {
	handler, service := newTestHandler(t)

	validID := uuid.New()
	url := "/users/" + validID.String()
	tUser := models.User{Id: validID, Login: "userToDelete"}

	t.Run("success", func(t *testing.T) {
		service.On("Delete", mock.Anything, validID).Return(tUser, nil).Once()

		req := httptest.NewRequest(http.MethodDelete, url, nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", handler.DeleteHandler)
		router.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var got models.User
		err := json.NewDecoder(resp.Body).Decode(&got)
		assert.NoError(t, err)
		assert.Equal(t, validID, got.Id)
		service.AssertExpectations(t)
	})

	t.Run("invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/users/not-uuid", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", handler.DeleteHandler)
		router.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("context cancelled error", func(t *testing.T) {
		service.On("Delete", mock.Anything, validID).Return(models.User{}, serviceerrors.ErrContextCanceled).Once()

		req := httptest.NewRequest(http.MethodDelete, url, nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", handler.DeleteHandler)
		router.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusRequestTimeout, resp.StatusCode)
		service.AssertExpectations(t)
	})

	t.Run("not found error", func(t *testing.T) {
		service.On("Delete", mock.Anything, validID).Return(models.User{}, serviceerrors.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodDelete, url, nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", handler.DeleteHandler)
		router.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		service.AssertExpectations(t)
	})

	t.Run("other error", func(t *testing.T) {
		service.On("Delete", mock.Anything, validID).Return(models.User{}, errors.New("other error")).Once()

		req := httptest.NewRequest(http.MethodDelete, url, nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/users/{id}", handler.DeleteHandler)
		router.ServeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		service.AssertExpectations(t)
	})
}
