package usersgrpcstorage

import (
	"apigateway/internal/domain/models"
	"apigateway/internal/domain/profiles"
	grpchelper "apigateway/pkg/lib/grpc/helper"
	"apigateway/pkg/lib/logger/sl"
	"context"
	"fmt"
	"log/slog"

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
		log.Error("failed to connect to gRPC server", sl.Err(err))
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
		log.Info("context is over")
		return nil, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	c := umv1.NewUsersManagerClient(s.Conn)
	res, err := c.GetUsers(ctx, nil)
	if err != nil {
		err := grpchelper.GrpcErrorHelper(s.Log, op, err)
		return nil, err
	}

	var usersForRet = make([]models.User, 0, len(res.GetUsers()))

	for _, pbUser := range res.GetUsers() {
		tmpUser, err := profiles.ProtoUsrToUsr(pbUser)
		if err != nil {
			log.Warn("Wrong user format", sl.Err(err))
			continue
		}

		usersForRet = append(usersForRet, tmpUser)
	}

	return usersForRet, nil
}

func (s *GRPCUsersStorage) GetUserById(ctx context.Context, uid uuid.UUID) (models.User, error) {
	const op = "storage.users.grpc.GetUserById"
	log := s.Log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("context is over")
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	c := umv1.NewUsersManagerClient(s.Conn)
	res, err := c.GetUserById(ctx, &umv1.GetUserByIdRequest{Id: uid.String()})
	if err != nil {
		err := grpchelper.GrpcErrorHelper(s.Log, op, err)
		return models.User{}, err
	}

	user, err := profiles.ProtoUsrToUsr(res.GetUser())
	if err != nil {
		log.Error("Wrong user format", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *GRPCUsersStorage) Insert(ctx context.Context, userForInsert models.User) (models.User, error) {
	const op = "storage.users.grpc.Insert"
	log := s.Log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("context is over")
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	pbUserForInsert := profiles.UsrToProtoUsr(userForInsert)

	c := umv1.NewUsersManagerClient(s.Conn)
	res, err := c.Insert(ctx, &umv1.InsertRequest{User: pbUserForInsert})
	if err != nil {
		err := grpchelper.GrpcErrorHelper(s.Log, op, err)
		return models.User{}, err
	}

	insertedUser, err := profiles.ProtoUsrToUsr(res.GetUser())
	if err != nil {
		log.Error("Wrong user format", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return insertedUser, nil
}

func (s *GRPCUsersStorage) Update(ctx context.Context, uid uuid.UUID, userForUpdate models.User) (models.User, error) {
	const op = "storage.users.grpc.Update"
	log := s.Log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("context is over")
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	pbUserForUpdate := profiles.UsrToProtoUsr(userForUpdate)
	c := umv1.NewUsersManagerClient(s.Conn)
	res, err := c.Update(ctx,
		&umv1.UpdateRequest{
			Id:   uid.String(),
			User: pbUserForUpdate,
		})
	if err != nil {
		err := grpchelper.GrpcErrorHelper(s.Log, op, err)
		return models.User{}, err
	}

	updatedUser, err := profiles.ProtoUsrToUsr(res.GetUser())
	if err != nil {
		log.Error("Wrong user format", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return updatedUser, nil
}

func (s *GRPCUsersStorage) Delete(ctx context.Context, uid uuid.UUID) (models.User, error) {
	const op = "storage.users.grpc.Delete"
	log := s.Log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("context is over")
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	c := umv1.NewUsersManagerClient(s.Conn)
	res, err := c.Delete(ctx, &umv1.DeleteRequest{Id: uid.String()})
	if err != nil {
		err := grpchelper.GrpcErrorHelper(s.Log, op, err)
		return models.User{}, err
	}

	deletedUser, err := profiles.ProtoUsrToUsr(res.GetUser())
	if err != nil {
		log.Error("Wrong user format", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return deletedUser, nil
}
