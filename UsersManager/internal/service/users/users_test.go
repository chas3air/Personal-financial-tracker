package usersservice_test

import (
	"context"
	"testing"
	"usersmanager/internal/domain/models"
	serviceerros "usersmanager/internal/service"
	usersservice "usersmanager/internal/service/users"
	storageerrors "usersmanager/internal/storage"
	"usersmanager/pkg/lib/logger/handler/slogdiscard"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock for IUsersStorage ---

type MockUsersStorage struct {
	mock.Mock
}

func (m *MockUsersStorage) GetUsers(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUsersStorage) GetUserById(ctx context.Context, uid uuid.UUID) (models.User, error) {
	args := m.Called(ctx, uid)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUsersStorage) Insert(ctx context.Context, user models.User) (models.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUsersStorage) Update(ctx context.Context, uid uuid.UUID, user models.User) (models.User, error) {
	args := m.Called(ctx, uid, user)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUsersStorage) Delete(ctx context.Context, uid uuid.UUID) (models.User, error) {
	args := m.Called(ctx, uid)
	return args.Get(0).(models.User), args.Error(1)
}

// --- Tests ---

func newTestService(storage *MockUsersStorage) *usersservice.UsersService {
	logger := slogdiscard.NewDiscardLogger()
	return usersservice.New(logger, storage)
}

func TestGetUser_Success(t *testing.T) {
	mockStorage := new(MockUsersStorage)
	users := []models.User{
		{Id: uuid.New()},
		{Id: uuid.New()},
	}

	mockStorage.On("GetUsers", mock.Anything).Return(users, nil)

	svc := newTestService(mockStorage)
	got, err := svc.GetUsers(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, users, got)

	mockStorage.AssertExpectations(t)
}

func TestGetUsers_ContextCanceled(t *testing.T) {
	mockStorage := new(MockUsersStorage)
	svc := newTestService(mockStorage)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	users, err := svc.GetUsers(ctx)
	assert.Nil(t, users)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestGetUserById_Success(t *testing.T) {
	mockStorage := new(MockUsersStorage)
	id := uuid.New()
	user := models.User{Id: id, Login: "user1"}
	mockStorage.On("GetUserById", mock.Anything, id).Return(user, nil)

	svc := newTestService(mockStorage)
	got, err := svc.GetUserById(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, user, got)
	mockStorage.AssertExpectations(t)
}

func TestGetUserById_NotFound(t *testing.T) {
	mockStorage := new(MockUsersStorage)
	id := uuid.New()
	mockStorage.On("GetUserById", mock.Anything, id).Return(models.User{}, storageerrors.ErrNotFound)

	svc := newTestService(mockStorage)
	_, err := svc.GetUserById(context.Background(), id)

	assert.ErrorIs(t, err, serviceerros.ErrNotFound)
	mockStorage.AssertExpectations(t)
}

func TestInsert_Success(t *testing.T) {
	mockStorage := new(MockUsersStorage)
	user := models.User{Id: uuid.New(), Login: "user1"}
	mockStorage.On("Insert", mock.Anything, user).Return(user, nil)

	svc := newTestService(mockStorage)
	got, err := svc.Insert(context.Background(), user)

	assert.NoError(t, err)
	assert.Equal(t, user, got)
	mockStorage.AssertExpectations(t)
}

func TestInsert_AlreadyExists(t *testing.T) {
	mockStorage := new(MockUsersStorage)
	user := models.User{Id: uuid.New(), Login: "user1"}
	mockStorage.On("Insert", mock.Anything, user).Return(models.User{}, storageerrors.ErrAlreadyExists)

	svc := newTestService(mockStorage)
	_, err := svc.Insert(context.Background(), user)

	assert.ErrorIs(t, err, serviceerros.ErrAlreadyExists)
	mockStorage.AssertExpectations(t)
}

func TestUpdate_Success(t *testing.T) {
	mockStorage := new(MockUsersStorage)
	id := uuid.New()
	user := models.User{Id: id, Login: "user1"}
	mockStorage.On("Update", mock.Anything, id, user).Return(user, nil)

	svc := newTestService(mockStorage)
	got, err := svc.Update(context.Background(), id, user)

	assert.NoError(t, err)
	assert.Equal(t, user, got)
	mockStorage.AssertExpectations(t)
}

func TestUpdate_NotFound(t *testing.T) {
	mockStorage := new(MockUsersStorage)
	id := uuid.New()
	user := models.User{Id: id, Login: "user1"}
	mockStorage.On("Update", mock.Anything, id, user).Return(models.User{}, storageerrors.ErrNotFound)

	svc := newTestService(mockStorage)
	_, err := svc.Update(context.Background(), id, user)

	assert.ErrorIs(t, err, serviceerros.ErrNotFound)
	mockStorage.AssertExpectations(t)
}

func TestDelete_Success(t *testing.T) {
	mockStorage := new(MockUsersStorage)
	id := uuid.New()
	user := models.User{Id: id, Login: "user1"}
	mockStorage.On("Delete", mock.Anything, id).Return(user, nil)

	svc := newTestService(mockStorage)
	got, err := svc.Delete(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, user, got)
	mockStorage.AssertExpectations(t)
}

func TestDelete_NotFound(t *testing.T) {
	mockStorage := new(MockUsersStorage)
	id := uuid.New()
	mockStorage.On("Delete", mock.Anything, id).Return(models.User{}, storageerrors.ErrNotFound)

	svc := newTestService(mockStorage)
	_, err := svc.Delete(context.Background(), id)

	assert.ErrorIs(t, err, serviceerros.ErrNotFound)
	mockStorage.AssertExpectations(t)
}
