package usersgrpc

import (
	"context"
	"errors"
	"log/slog"
	"usersmanager/internal/domain/models"
	"usersmanager/internal/domain/profiles"
	serviceerrors "usersmanager/internal/service"
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
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return nil, status.Error(codes.Canceled, "context is over")
	default:
	}

	users, err := s.Service.GetUsers(ctx)
	if err != nil {
		log.Error("Failed to fetch users", sl.Err(err))
		return nil, status.Error(codes.Internal, "failed to fetch users")
	}

	var pbUsers = make([]*umv1.User, 0, len(users))
	for _, user := range users {
		pbUsers = append(pbUsers, profiles.UsrToProtoUsr(user))
	}

	log.Info("Users fetched successfully")
	return &umv1.GetUsersResponse{
		Users: pbUsers,
	}, nil
}

func (s *ServerAPI) GetUserById(ctx context.Context, req *umv1.GetUserByIdRequest) (*umv1.GetUserByIdResponse, error) {
	const op = "grpc.users.GetUserById"
	log := s.Log.With(
		"op", op,
	)

	select {
	case <-ctx.Done():
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return nil, status.Error(codes.Canceled, "context is over")
	default:
	}

	uid, err := uuid.Parse(req.GetId())
	if err != nil {
		log.Error("Invalid user ID format", sl.Err(err))
		return nil, status.Error(codes.InvalidArgument, "invalid id format")
	}

	user, err := s.Service.GetUserById(ctx, uid)
	if err != nil {
		if errors.Is(err, serviceerrors.ErrNotFound) {
			log.Warn("User not found", sl.Err(serviceerrors.ErrNotFound))
			return nil, status.Error(codes.NotFound, "user not found")
		}

		log.Error("Failed to fetch user by ID", sl.Err(err))
		return nil, status.Error(codes.Internal, "failed to fetch user by id")
	}

	log.Info("User fetched successfully", slog.String("user_id", user.Id.String()))
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
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return nil, status.Error(codes.Canceled, "context is over")
	default:
	}

	userForInsert, err := profiles.ProtoUsrToUsr(req.GetUser())
	if err != nil {
		log.Error("Invalid user data for insertion", sl.Err(err))
		return nil, status.Error(codes.InvalidArgument, "invalid user data")
	}

	insertedUser, err := s.Service.Insert(ctx, userForInsert)
	if err != nil {
		if errors.Is(err, serviceerrors.ErrAlreadyExists) {
			log.Warn("User with given ID or login already exists", sl.Err(serviceerrors.ErrAlreadyExists))
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		log.Error("Failed to insert user", sl.Err(err))
		return nil, status.Error(codes.Internal, "failed to insert user")
	}

	log.Info("User inserted successfully", slog.String("user_id", insertedUser.Id.String()))
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
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return nil, status.Error(codes.Canceled, "context is over")
	default:
	}

	idForUpdate, err := uuid.Parse(req.GetId())
	if err != nil {
		log.Error("Invalid user ID format for update", sl.Err(err))
		return nil, status.Error(codes.InvalidArgument, "invalid id format for update")
	}

	userForUpdate, err := profiles.ProtoUsrToUsr(req.GetUser())
	if err != nil {
		log.Error("Invalid user data for update", sl.Err(err))
		return nil, status.Error(codes.InvalidArgument, "invalid user data for update")
	}

	updatedUser, err := s.Service.Update(ctx, idForUpdate, userForUpdate)
	if err != nil {
		if errors.Is(err, serviceerrors.ErrNotFound) {
			log.Warn("User not found for update", sl.Err(serviceerrors.ErrNotFound))
			return nil, status.Error(codes.NotFound, "user not found for update")
		}

		log.Error("Failed to update user", sl.Err(err))
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	log.Info("User updated successfully", slog.String("user_id", updatedUser.Id.String()))
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
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return nil, status.Error(codes.Canceled, "context is over")
	default:
	}

	idForDelete, err := uuid.Parse(req.GetId())
	if err != nil {
		log.Error("Invalid user ID format for deletion", sl.Err(err))
		return nil, status.Error(codes.InvalidArgument, "invalid id format for deletion")
	}

	deletedUser, err := s.Service.Delete(ctx, idForDelete)
	if err != nil {
		if errors.Is(err, serviceerrors.ErrNotFound) {
			log.Warn("User not found for deletion", sl.Err(serviceerrors.ErrNotFound))
			return nil, status.Error(codes.NotFound, "user not found for deletion")
		}

		log.Error("Failed to delete user", sl.Err(err))
		return nil, status.Error(codes.Internal, "failed to delete user")
	}

	log.Info("User deleted successfully", slog.String("user_id", deletedUser.Id.String()))
	return &umv1.DeleteResponse{
		User: profiles.UsrToProtoUsr(deletedUser),
	}, nil
}
