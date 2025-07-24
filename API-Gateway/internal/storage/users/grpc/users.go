package usersgrpcstorage

import (
	"context"
	"fmt"
	"log/slog"

	"apigateway/internal/domain/models"
	"apigateway/internal/domain/profiles"
	grpchelper "apigateway/pkg/lib/grpc/helper"
	"apigateway/pkg/lib/logger/sl"

	umv1 "github.com/chas3air/protos/gen/go/usersManager"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCUsersStorage struct {
	Log  *slog.Logger
	Conn *grpc.ClientConn
}

func New(log *slog.Logger, host string, port int) *GRPCUsersStorage {
	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("Failed to connect to gRPC server", sl.Err(err))
		panic(err)
	}

	return &GRPCUsersStorage{
		Log:  log,
		Conn: conn,
	}
}

func (g *GRPCUsersStorage) Close() {
	if err := g.Conn.Close(); err != nil {
		panic(err)
	}
}

func (s *GRPCUsersStorage) GetUsers(ctx context.Context) ([]models.User, error) {
	const op = "storage.users.grpc.GetUsers"
	log := s.Log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return nil, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	client := umv1.NewUsersManagerClient(s.Conn)
	res, err := client.GetUsers(ctx, &umv1.GetUsersRequest{})
	if err != nil {
		err = grpchelper.GrpcErrorHelper(log, op, err)
		return nil, err
	}

	usersForRet := make([]models.User, 0, len(res.GetUsers()))

	for _, pbUser := range res.GetUsers() {
		tmpUser, err := profiles.ProtoUsrToUsr(pbUser)
		if err != nil {
			log.Warn("Wrong user format", sl.Err(err))
			continue
		}

		usersForRet = append(usersForRet, tmpUser)
	}

	log.Info("Users fetched successfully", slog.Int("count", len(usersForRet)))
	return usersForRet, nil
}

func (s *GRPCUsersStorage) GetUserById(ctx context.Context, uid uuid.UUID) (models.User, error) {
	const op = "storage.users.grpc.GetUserById"
	log := s.Log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	client := umv1.NewUsersManagerClient(s.Conn)
	res, err := client.GetUserById(ctx, &umv1.GetUserByIdRequest{Id: uid.String()})
	if err != nil {
		err = grpchelper.GrpcErrorHelper(log, op, err)
		return models.User{}, err
	}

	user, err := profiles.ProtoUsrToUsr(res.GetUser())
	if err != nil {
		log.Error("Wrong user format", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("User fetched successfully", slog.String("user_id", user.Id.String()))
	return user, nil
}

func (s *GRPCUsersStorage) Insert(ctx context.Context, userForInsert models.User) (models.User, error) {
	const op = "storage.users.grpc.Insert"
	log := s.Log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	pbUserForInsert := profiles.UsrToProtoUsr(userForInsert)

	client := umv1.NewUsersManagerClient(s.Conn)
	res, err := client.Insert(ctx, &umv1.InsertRequest{User: pbUserForInsert})
	if err != nil {
		err = grpchelper.GrpcErrorHelper(log, op, err)
		return models.User{}, err
	}

	insertedUser, err := profiles.ProtoUsrToUsr(res.GetUser())
	if err != nil {
		log.Error("Wrong user format", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("User inserted successfully", slog.String("user_id", insertedUser.Id.String()))
	return insertedUser, nil
}

func (s *GRPCUsersStorage) Update(ctx context.Context, uid uuid.UUID, userForUpdate models.User) (models.User, error) {
	const op = "storage.users.grpc.Update"
	log := s.Log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	pbUserForUpdate := profiles.UsrToProtoUsr(userForUpdate)

	client := umv1.NewUsersManagerClient(s.Conn)
	res, err := client.Update(ctx, &umv1.UpdateRequest{
		Id:   uid.String(),
		User: pbUserForUpdate,
	})
	if err != nil {
		err = grpchelper.GrpcErrorHelper(log, op, err)
		return models.User{}, err
	}

	updatedUser, err := profiles.ProtoUsrToUsr(res.GetUser())
	if err != nil {
		log.Error("Wrong user format", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("User updated successfully", slog.String("user_id", updatedUser.Id.String()))
	return updatedUser, nil
}

func (s *GRPCUsersStorage) Delete(ctx context.Context, uid uuid.UUID) (models.User, error) {
	const op = "storage.users.grpc.Delete"
	log := s.Log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	client := umv1.NewUsersManagerClient(s.Conn)
	res, err := client.Delete(ctx, &umv1.DeleteRequest{Id: uid.String()})
	if err != nil {
		err = grpchelper.GrpcErrorHelper(log, op, err)
		return models.User{}, err
	}

	deletedUser, err := profiles.ProtoUsrToUsr(res.GetUser())
	if err != nil {
		log.Error("Wrong user format", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("User deleted successfully", slog.String("user_id", deletedUser.Id.String()))
	return deletedUser, nil
}
