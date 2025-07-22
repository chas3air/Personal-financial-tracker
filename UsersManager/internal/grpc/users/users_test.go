package usersgrpc_test

import (
	"context"
	"errors"
	"testing"

	"usersmanager/internal/domain/models"
	"usersmanager/internal/domain/profiles"
	usersgrpc "usersmanager/internal/grpc/users"
	serviceerrors "usersmanager/internal/service"
	"usersmanager/pkg/lib/logger/handler/slogdiscard"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	umv1 "github.com/chas3air/protos/gen/go/usersManager"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Mock сервиса пользователей
type mockUsersService struct {
	mock.Mock
}

func (m *mockUsersService) GetUsers(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
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

func newServerAPI(t *testing.T) (*usersgrpc.ServerAPI, *mockUsersService) {
	svc := new(mockUsersService)
	logger := slogdiscard.NewDiscardLogger()
	server := &usersgrpc.ServerAPI{Log: logger, Service: svc}
	return server, svc
}

func TestServerAPI_GetUsers(t *testing.T) {
	server, svc := newServerAPI(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		users := []models.User{
			{Id: uuid.New(), Login: "user1", Password: "p1", Role: "admin"},
			{Id: uuid.New(), Login: "user2", Password: "p2", Role: "user"},
		}
		svc.On("GetUsers", ctx).Return(users, nil).Once()

		resp, err := server.GetUsers(ctx, &umv1.GetUsersRequest{})
		assert.NoError(t, err)
		assert.Len(t, resp.Users, 2)
		svc.AssertExpectations(t)
	})

	t.Run("error fetching users", func(t *testing.T) {
		svc.On("GetUsers", ctx).Return([]models.User{}, errors.New("db error")).Once()

		_, err := server.GetUsers(ctx, &umv1.GetUsersRequest{})
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		svc.AssertExpectations(t)
	})

	t.Run("context done", func(t *testing.T) {
		ctxCanceled, cancel := context.WithCancel(ctx)
		cancel()

		_, err := server.GetUsers(ctxCanceled, &umv1.GetUsersRequest{})
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.DeadlineExceeded, st.Code())
	})
}

func TestServerAPI_GetUserById(t *testing.T) {
	server, svc := newServerAPI(t)
	ctx := context.Background()
	id := uuid.New()
	req := &umv1.GetUserByIdRequest{Id: id.String()}

	user := models.User{Id: id, Login: "user1", Password: "pass", Role: "admin"}

	t.Run("success", func(t *testing.T) {
		svc.On("GetUserById", ctx, id).Return(user, nil).Once()

		resp, err := server.GetUserById(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, user.Id.String(), resp.User.Id)
		svc.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		svc.On("GetUserById", ctx, id).Return(models.User{}, serviceerrors.ErrNotFound).Once()

		_, err := server.GetUserById(ctx, req)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		svc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		svc.On("GetUserById", ctx, id).Return(models.User{}, errors.New("db error")).Once()

		_, err := server.GetUserById(ctx, req)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		svc.AssertExpectations(t)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		badReq := &umv1.GetUserByIdRequest{Id: "invalid-uuid"}
		_, err := server.GetUserById(ctx, badReq)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		svc.AssertExpectations(t)
	})
}

func TestServerAPI_Insert(t *testing.T) {
	server, svc := newServerAPI(t)
	ctx := context.Background()
	user := models.User{Id: uuid.New(), Login: "u1", Password: "p1", Role: "admin"}
	pbUser := profiles.UsrToProtoUsr(user)
	pbInsertReq := &umv1.InsertRequest{User: pbUser}

	t.Run("success", func(t *testing.T) {
		svc.On("Insert", ctx, user).Return(user, nil).Once()

		resp, err := server.Insert(ctx, pbInsertReq)
		assert.NoError(t, err)
		assert.Equal(t, pbUser.Id, resp.User.Id)
		svc.AssertExpectations(t)
	})

	t.Run("user exists error", func(t *testing.T) {
		svc.On("Insert", ctx, user).Return(models.User{}, serviceerrors.ErrAlreadyExists).Once()

		_, err := server.Insert(ctx, pbInsertReq)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, st.Code())
		svc.AssertExpectations(t)
	})

	t.Run("invalid user in proto", func(t *testing.T) {
		// Создадим некорректный protobuf-пользователь, который вызовет ошибку преобразования
		badReq := &umv1.InsertRequest{User: &umv1.User{Id: "invalid-uuid"}}
		_, err := server.Insert(ctx, badReq)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestServerAPI_Update(t *testing.T) {
	server, svc := newServerAPI(t)
	ctx := context.Background()
	user := models.User{Id: uuid.New(), Login: "u1", Password: "p1", Role: "admin"}
	pbUser := profiles.UsrToProtoUsr(user)
	req := &umv1.UpdateRequest{
		Id:   user.Id.String(),
		User: pbUser,
	}

	t.Run("success", func(t *testing.T) {
		svc.On("Update", ctx, user.Id, user).Return(user, nil).Once()

		resp, err := server.Update(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, pbUser.Id, resp.User.Id)
		svc.AssertExpectations(t)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		badReq := &umv1.UpdateRequest{Id: "not-uuid", User: pbUser}
		_, err := server.Update(ctx, badReq)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("invalid user proto", func(t *testing.T) {
		badReq := &umv1.UpdateRequest{Id: user.Id.String(), User: &umv1.User{Id: "bad-uuid"}}
		_, err := server.Update(ctx, badReq)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("user not found", func(t *testing.T) {
		svc.On("Update", ctx, user.Id, user).Return(models.User{}, serviceerrors.ErrNotFound).Once()

		_, err := server.Update(ctx, req)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		svc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		svc.On("Update", ctx, user.Id, user).Return(models.User{}, errors.New("db error")).Once()

		_, err := server.Update(ctx, req)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		svc.AssertExpectations(t)
	})
}

func TestServerAPI_Delete(t *testing.T) {
	server, svc := newServerAPI(t)
	ctx := context.Background()
	user := models.User{Id: uuid.New(), Login: "u1", Password: "p1", Role: "admin"}
	req := &umv1.DeleteRequest{Id: user.Id.String()}

	t.Run("success", func(t *testing.T) {
		svc.On("Delete", ctx, user.Id).Return(user, nil).Once()

		resp, err := server.Delete(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, user.Id.String(), resp.User.Id)
		svc.AssertExpectations(t)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		badReq := &umv1.DeleteRequest{Id: "not-uuid"}
		_, err := server.Delete(ctx, badReq)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("user not found", func(t *testing.T) {
		svc.On("Delete", ctx, user.Id).Return(models.User{}, serviceerrors.ErrNotFound).Once()

		_, err := server.Delete(ctx, req)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		svc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		svc.On("Delete", ctx, user.Id).Return(models.User{}, errors.New("db error")).Once()

		_, err := server.Delete(ctx, req)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		svc.AssertExpectations(t)
	})
}
