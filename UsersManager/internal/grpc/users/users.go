package usersgrpc

import (
	"context"
	"log/slog"
	"usersmanager/internal/domain/models"

	umv1 "github.com/chas3air/protos/gen/go/usersManager"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IUsersService interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	GetUserById(ctx context.Context, uid uuid.UUID) (models.User, error)
	Insert(ctx context.Context, user models.User) (models.User, error)
	Update(ctx context.Context, uid uuid.UUID, user models.User) (models.User, error)
	Delete(ctx context.Context, uid uuid.UUID) (models.User, error)
}

type ServerAPI struct {
	log     *slog.Logger
	Service IUsersService
	umv1.UnimplementedUsersManagerServer
}

func Register(grpc *grpc.Server, log *slog.Logger, service IUsersService) {
	umv1.RegisterUsersManagerServer(grpc, &ServerAPI{log: log, Service: service})
}

func (s *ServerAPI) GetUsers(context.Context, *umv1.GetUsersRequest) (*umv1.GetUsersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUsers not implemented")
}

func (s *ServerAPI) GetUserById(context.Context, *umv1.GetUserByIdRequest) (*umv1.GetUserByIdResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserById not implemented")
}

func (s *ServerAPI) Insert(context.Context, *umv1.InsertRequest) (*umv1.InsertResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Insert not implemented")
}

func (s *ServerAPI) Update(context.Context, *umv1.UpdateRequest) (*umv1.UpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}

func (s *ServerAPI) Delete(context.Context, *umv1.DeleteRequest) (*umv1.DeleteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
