package usersgrpc

import (
	"context"
	"errors"
	"log/slog"
	"usersmanager/internal/domain/models"
	"usersmanager/internal/domain/profiles"
	serviceerros "usersmanager/internal/service"
	"usersmanager/pkg/lib/logger/sl"

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
	Log     *slog.Logger
	Service IUsersService
	umv1.UnimplementedUsersManagerServer
}

func Register(grpc *grpc.Server, log *slog.Logger, service IUsersService) {
	umv1.RegisterUsersManagerServer(grpc, &ServerAPI{Log: log, Service: service})
}

func (s *ServerAPI) GetUsers(ctx context.Context, req *umv1.GetUsersRequest) (*umv1.GetUsersResponse, error) {
	const op = "grpc.users.GetUsers"
	log := s.Log.With(
		"op", op,
	)

	select {
	case <-ctx.Done():
		log.Info("context is over")
		return nil, status.Error(codes.DeadlineExceeded, "context is over")
	default:
	}

	users, err := s.Service.GetUsers(ctx)
	if err != nil {
		log.Error("Error fetching users", sl.Err(err))
		return nil, status.Error(codes.Internal, "Error fetching users")
	}

	var pbUsers = make([]*umv1.User, 0, len(users))
	for _, user := range users {
		pbUsers = append(pbUsers, profiles.UsrToProtoUsr(user))
	}

	return &umv1.GetUsersResponse{
		Users: pbUsers,
	}, nil
}

func (s *ServerAPI) GetUserById(ctx context.Context, req *umv1.GetUserByIdRequest) (*umv1.GetUserByIdResponse, error) {
	const op = "grpc.users.GetUsers"
	log := s.Log.With(
		"op", op,
	)

	select {
	case <-ctx.Done():
		log.Info("context is over")
		return nil, status.Error(codes.DeadlineExceeded, "context is over")
	default:
	}

	uid, err := uuid.Parse(req.GetId())
	if err != nil {
		log.Error("invalid id", sl.Err(err))
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	user, err := s.Service.GetUserById(ctx, uid)
	if err != nil {
		if errors.Is(err, serviceerros.ErrNotFound) {
			log.Warn("User not found", sl.Err(serviceerros.ErrNotFound))
			return nil, status.Error(codes.NotFound, "User nt found")
		}

		log.Error("Error fetching users by id", sl.Err(err))
		return nil, status.Error(codes.Internal, "Error fetching users by id")
	}

	return &umv1.GetUserByIdResponse{
		User: profiles.UsrToProtoUsr(user),
	}, nil
}

func (s *ServerAPI) Insert(ctx context.Context, req *umv1.InsertRequest) (*umv1.InsertResponse, error) {
	const op = "grpc.users.Insert"
	log := s.Log.With(
		"op", op,
	)

	select {
	case <-ctx.Done():
		log.Info("context is over")
		return nil, status.Error(codes.DeadlineExceeded, "context is over")
	default:
	}

	userForInsert, err := profiles.ProtoUsrToUsr(req.GetUser())
	if err != nil {
		log.Error("invalid user", sl.Err(err))
		return nil, status.Error(codes.InvalidArgument, "invalid argument")
	}

	insertedUser, err := s.Service.Insert(ctx, userForInsert)
	if err != nil {
		if errors.Is(err, serviceerros.ErrAlreadyExists) {
			log.Warn("User already exists", sl.Err(serviceerros.ErrAlreadyExists))
			return nil, status.Error(codes.AlreadyExists, "User already exists")
		}

		log.Error("Error inserting user", sl.Err(err))
		return nil, status.Error(codes.Internal, "Error inserting user")
	}

	return &umv1.InsertResponse{
		User: profiles.UsrToProtoUsr(insertedUser),
	}, nil
}

func (s *ServerAPI) Update(ctx context.Context, req *umv1.UpdateRequest) (*umv1.UpdateResponse, error) {
	const op = "grpc.users.Update"
	log := s.Log.With(
		"op", op,
	)

	select {
	case <-ctx.Done():
		log.Info("context is over")
		return nil, status.Error(codes.DeadlineExceeded, "context is over")
	default:
	}

	idForUpdate, err := uuid.Parse(req.GetId())
	if err != nil {
		log.Error("invalid id", sl.Err(err))
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	userForUpdate, err := profiles.ProtoUsrToUsr(req.GetUser())
	if err != nil {
		log.Error("invalid user for update", sl.Err(err))
		return nil, status.Error(codes.InvalidArgument, "invalid user for update")
	}

	updatedUser, err := s.Service.Update(ctx, idForUpdate, userForUpdate)
	if err != nil {
		if errors.Is(err, serviceerros.ErrNotFound) {
			log.Warn("User not found", sl.Err(serviceerros.ErrNotFound))
			return nil, status.Error(codes.NotFound, "user not found")
		}

		log.Error("Error updating user", sl.Err(err))
		return nil, status.Error(codes.Internal, "error updating user")
	}

	return &umv1.UpdateResponse{
		User: profiles.UsrToProtoUsr(updatedUser),
	}, nil
}

func (s *ServerAPI) Delete(ctx context.Context, req *umv1.DeleteRequest) (*umv1.DeleteResponse, error) {
	const op = "grpc.users.Delete"
	log := s.Log.With(
		"op", op,
	)

	select {
	case <-ctx.Done():
		log.Info("context is over")
		return nil, status.Error(codes.DeadlineExceeded, "context is over")
	default:
	}

	idForDelete, err := uuid.Parse(req.GetId())
	if err != nil {
		log.Error("invalid id", sl.Err(err))
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	deletedUser, err := s.Service.Delete(ctx, idForDelete)
	if err != nil {
		if errors.Is(err, serviceerros.ErrNotFound) {
			log.Warn("User not found", sl.Err(serviceerros.ErrNotFound))
			return nil, status.Error(codes.NotFound, "user not found")
		}

		log.Error("Error deleting user", sl.Err(err))
		return nil, status.Error(codes.Internal, "error deleting user")
	}

	return &umv1.DeleteResponse{
		User: profiles.UsrToProtoUsr(deletedUser),
	}, nil
}
