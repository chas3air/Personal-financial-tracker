package grpchelper

import (
	storageerrors "apigateway/internal/storage"
	"apigateway/pkg/lib/logger/sl"
	"fmt"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GrpcErrorHelper(log *slog.Logger, op string, err error) error {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Canceled:
			log.Warn("Context cancelled", sl.Err(err))
			return fmt.Errorf("%s: %w", op, storageerrors.ErrContextCanceled)

		case codes.DeadlineExceeded:
			log.Warn("Deadline exeeced", sl.Err(err))
			return fmt.Errorf("%s: %w", op, storageerrors.ErrDeadlineExeeced)

		case codes.InvalidArgument:
			log.Warn("Invalid arguments", sl.Err(err))
			return fmt.Errorf("%s: %w", op, storageerrors.ErrInvalidArgument)

		case codes.AlreadyExists:
			log.Warn("Record with given ID already exists", sl.Err(err))
			return fmt.Errorf("%s: %w", op, storageerrors.ErrAlreadyExists)

		case codes.NotFound:
			log.Warn("Record not found", sl.Err(err))
			return fmt.Errorf("%s: %w", op, storageerrors.ErrNotFound)

		default:
			log.Error("Failed to fetch record by ID", sl.Err(err))
			return fmt.Errorf("%s: %w", op, storageerrors.ErrInternal)
		}
	} else {
		return fmt.Errorf("%s: %w", op, storageerrors.ErrInternal)
	}
}
