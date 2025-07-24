package usersservice_test

import (
	"context"
	"errors"
	"testing"

	"apigateway/internal/domain/models"
	serviceerrors "apigateway/internal/service"
	usersservice "apigateway/internal/service/users"
	storageerrors "apigateway/internal/storage"
	"apigateway/pkg/lib/logger/handler/slogdiscard"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock IUsersStorage
type mockUsersStorage struct {
	mock.Mock
}

func (m *mockUsersStorage) GetUsers(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *mockUsersStorage) GetUserById(ctx context.Context, uid uuid.UUID) (models.User, error) {
	args := m.Called(ctx, uid)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *mockUsersStorage) Insert(ctx context.Context, user models.User) (models.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *mockUsersStorage) Update(ctx context.Context, uid uuid.UUID, user models.User) (models.User, error) {
	args := m.Called(ctx, uid, user)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *mockUsersStorage) Delete(ctx context.Context, uid uuid.UUID) (models.User, error) {
	args := m.Called(ctx, uid)
	return args.Get(0).(models.User), args.Error(1)
}

func newTestService(t *testing.T) (*usersservice.UsersService, *mockUsersStorage) {
	mockStorage := new(mockUsersStorage)
	logger := slogdiscard.NewDiscardLogger()
	svc := usersservice.New(logger, mockStorage)
	return svc, mockStorage
}

func TestUsersService_GetUsers(t *testing.T) {
	svc, mockStorage := newTestService(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		users := []models.User{
			{Id: uuid.New(), Login: "user1"},
			{Id: uuid.New(), Login: "user2"},
		}
		mockStorage.On("GetUsers", ctx).Return(users, nil).Once()

		fetchedUsers, err := svc.GetUsers(ctx)
		assert.NoError(t, err)
		assert.Len(t, fetchedUsers, 2)
		mockStorage.AssertExpectations(t)
	})

	t.Run("context canceled", func(t *testing.T) {
		ctxCanceled, cancel := context.WithCancel(ctx)
		cancel()

		_, err := svc.GetUsers(ctxCanceled)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled))
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage context canceled error", func(t *testing.T) {
		mockStorage.On("GetUsers", ctx).Return(nil, storageerrors.ErrContextCanceled).Once()

		_, err := svc.GetUsers(ctx)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrContextCanceled))
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage deadline exceeded error", func(t *testing.T) {
		mockStorage.On("GetUsers", ctx).Return(nil, storageerrors.ErrDeadlineExeeced).Once()

		_, err := svc.GetUsers(ctx)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrDeadlineExeeced))
		mockStorage.AssertExpectations(t)
	})

	t.Run("other storage error", func(t *testing.T) {
		someErr := errors.New("something went wrong in storage")
		mockStorage.On("GetUsers", ctx).Return(nil, someErr).Once()

		_, err := svc.GetUsers(ctx)
		assert.Error(t, err)
		assert.False(t, errors.Is(err, serviceerrors.ErrInternal))
		assert.True(t, errors.Is(err, someErr))
		mockStorage.AssertExpectations(t)
	})
}

func TestUsersService_GetUserById(t *testing.T) {
	svc, mockStorage := newTestService(t)
	ctx := context.Background()
	testID := uuid.New()
	testUser := models.User{Id: testID, Login: "test"}

	t.Run("success", func(t *testing.T) {
		mockStorage.On("GetUserById", ctx, testID).Return(testUser, nil).Once()

		user, err := svc.GetUserById(ctx, testID)
		assert.NoError(t, err)
		assert.Equal(t, testID, user.Id)
		mockStorage.AssertExpectations(t)
	})

	t.Run("context canceled", func(t *testing.T) {
		ctxCanceled, cancel := context.WithCancel(ctx)
		cancel()

		_, err := svc.GetUserById(ctxCanceled, testID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled))
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage context canceled error", func(t *testing.T) {
		mockStorage.On("GetUserById", ctx, testID).Return(models.User{}, storageerrors.ErrContextCanceled).Once()

		_, err := svc.GetUserById(ctx, testID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrContextCanceled))
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage deadline exceeded error", func(t *testing.T) {
		mockStorage.On("GetUserById", ctx, testID).Return(models.User{}, storageerrors.ErrDeadlineExeeced).Once()

		_, err := svc.GetUserById(ctx, testID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrDeadlineExeeced))
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage invalid argument error", func(t *testing.T) {
		mockStorage.On("GetUserById", ctx, testID).Return(models.User{}, storageerrors.ErrInvalidArgument).Once()

		_, err := svc.GetUserById(ctx, testID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrInvalidArgument))
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage not found error", func(t *testing.T) {
		mockStorage.On("GetUserById", ctx, testID).Return(models.User{}, storageerrors.ErrNotFound).Once()

		_, err := svc.GetUserById(ctx, testID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrNotFound))
		mockStorage.AssertExpectations(t)
	})

	t.Run("other storage error", func(t *testing.T) {
		someErr := errors.New("connection failed")
		mockStorage.On("GetUserById", ctx, testID).Return(models.User{}, someErr).Once()

		_, err := svc.GetUserById(ctx, testID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrInternal))
		mockStorage.AssertExpectations(t)
	})
}

func TestUsersService_Insert(t *testing.T) {
	svc, mockStorage := newTestService(t)
	ctx := context.Background()
	testUser := models.User{Id: uuid.New(), Login: "newuser"}

	t.Run("success", func(t *testing.T) {
		mockStorage.On("Insert", ctx, testUser).Return(testUser, nil).Once()

		insertedUser, err := svc.Insert(ctx, testUser)
		assert.NoError(t, err)
		assert.Equal(t, testUser.Id, insertedUser.Id)
		mockStorage.AssertExpectations(t)
	})

	t.Run("context canceled", func(t *testing.T) {
		ctxCanceled, cancel := context.WithCancel(ctx)
		cancel()

		_, err := svc.Insert(ctxCanceled, testUser)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled))
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage context canceled error", func(t *testing.T) {
		mockStorage.On("Insert", ctx, testUser).Return(models.User{}, storageerrors.ErrContextCanceled).Once()

		_, err := svc.Insert(ctx, testUser)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrContextCanceled))
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage deadline exceeded error", func(t *testing.T) {
		mockStorage.On("Insert", ctx, testUser).Return(models.User{}, storageerrors.ErrDeadlineExeeced).Once()

		_, err := svc.Insert(ctx, testUser)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrDeadlineExeeced))
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage invalid argument error", func(t *testing.T) {
		mockStorage.On("Insert", ctx, testUser).Return(models.User{}, storageerrors.ErrInvalidArgument).Once()

		_, err := svc.Insert(ctx, testUser)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrInvalidArgument))
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage already exists error", func(t *testing.T) {
		mockStorage.On("Insert", ctx, testUser).Return(models.User{}, storageerrors.ErrAlreadyExists).Once()

		_, err := svc.Insert(ctx, testUser)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrAlreadyExists))
		mockStorage.AssertExpectations(t)
	})

	t.Run("other storage error", func(t *testing.T) {
		someErr := errors.New("unique constraint violation")
		mockStorage.On("Insert", ctx, testUser).Return(models.User{}, someErr).Once()

		_, err := svc.Insert(ctx, testUser)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrInternal))
		mockStorage.AssertExpectations(t)
	})
}

func TestUsersService_Update(t *testing.T) {
	svc, mockStorage := newTestService(t)
	ctx := context.Background()
	testID := uuid.New()
	testUser := models.User{Id: testID, Login: "updateduser"}

	t.Run("success", func(t *testing.T) {
		mockStorage.On("Update", ctx, testID, testUser).Return(testUser, nil).Once()

		updatedUser, err := svc.Update(ctx, testID, testUser)
		assert.NoError(t, err)
		assert.Equal(t, testID, updatedUser.Id)
		mockStorage.AssertExpectations(t)
	})

	t.Run("context canceled", func(t *testing.T) {
		ctxCanceled, cancel := context.WithCancel(ctx)
		cancel()

		_, err := svc.Update(ctxCanceled, testID, testUser)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled))
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage context canceled error", func(t *testing.T) {
		mockStorage.On("Update", ctx, testID, testUser).Return(models.User{}, storageerrors.ErrContextCanceled).Once()

		_, err := svc.Update(ctx, testID, testUser)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrContextCanceled))
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage deadline exceeded error", func(t *testing.T) {
		mockStorage.On("Update", ctx, testID, testUser).Return(models.User{}, storageerrors.ErrDeadlineExeeced).Once()

		_, err := svc.Update(ctx, testID, testUser)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrDeadlineExeeced))
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage invalid argument error", func(t *testing.T) {
		mockStorage.On("Update", ctx, testID, testUser).Return(models.User{}, storageerrors.ErrInvalidArgument).Once()

		_, err := svc.Update(ctx, testID, testUser)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrInvalidArgument))
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage not found error", func(t *testing.T) {
		mockStorage.On("Update", ctx, testID, testUser).Return(models.User{}, storageerrors.ErrNotFound).Once()

		_, err := svc.Update(ctx, testID, testUser)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrNotFound))
		mockStorage.AssertExpectations(t)
	})

	t.Run("other storage error", func(t *testing.T) {
		someErr := errors.New("database connection lost")
		mockStorage.On("Update", ctx, testID, testUser).Return(models.User{}, someErr).Once()

		_, err := svc.Update(ctx, testID, testUser)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrInternal))
		mockStorage.AssertExpectations(t)
	})
}

func TestUsersService_Delete(t *testing.T) {
	svc, mockStorage := newTestService(t)
	ctx := context.Background()
	testID := uuid.New()
	testUser := models.User{Id: testID, Login: "deleteduser"}

	t.Run("success", func(t *testing.T) {
		mockStorage.On("Delete", ctx, testID).Return(testUser, nil).Once()

		deletedUser, err := svc.Delete(ctx, testID)
		assert.NoError(t, err)
		assert.Equal(t, testID, deletedUser.Id)
		mockStorage.AssertExpectations(t)
	})

	t.Run("context canceled", func(t *testing.T) {
		ctxCanceled, cancel := context.WithCancel(ctx)
		cancel()

		_, err := svc.Delete(ctxCanceled, testID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled))
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage context canceled error", func(t *testing.T) {
		mockStorage.On("Delete", ctx, testID).Return(models.User{}, storageerrors.ErrContextCanceled).Once()

		_, err := svc.Delete(ctx, testID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrContextCanceled))
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage deadline exceeded error", func(t *testing.T) {
		mockStorage.On("Delete", ctx, testID).Return(models.User{}, storageerrors.ErrDeadlineExeeced).Once()

		_, err := svc.Delete(ctx, testID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrDeadlineExeeced))
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage invalid argument error", func(t *testing.T) {
		mockStorage.On("Delete", ctx, testID).Return(models.User{}, storageerrors.ErrInvalidArgument).Once()

		_, err := svc.Delete(ctx, testID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrInvalidArgument))
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage not found error", func(t *testing.T) {
		mockStorage.On("Delete", ctx, testID).Return(models.User{}, storageerrors.ErrNotFound).Once()

		_, err := svc.Delete(ctx, testID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrNotFound))
		mockStorage.AssertExpectations(t)
	})

	t.Run("other storage error", func(t *testing.T) {
		someErr := errors.New("network issue during delete")
		mockStorage.On("Delete", ctx, testID).Return(models.User{}, someErr).Once()

		_, err := svc.Delete(ctx, testID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, serviceerrors.ErrInternal))
		mockStorage.AssertExpectations(t)
	})
}
