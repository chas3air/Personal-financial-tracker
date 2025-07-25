package usersgrpcstorage

import (
	"auth/pkg/lib/logger/sl"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCUsersStorage struct {
	Log  *slog.Logger
	Conn *grpc.ClientConn
}

// New creates a new GRPCUsersStorage instance.
// It establishes a gRPC connection to the given host and port using insecure credentials.
// Panics if the connection cannot be established.
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
