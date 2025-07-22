package userspsqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"usersmanager/internal/domain/models"
	storageerrors "usersmanager/internal/storage"
	"usersmanager/pkg/lib/logger/sl"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type UsersPsqlStorage struct {
	Log       *slog.Logger
	DB        *sql.DB
	TableName string
}

func New(log *slog.Logger, connStr string, tableName string) *UsersPsqlStorage {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	wd, _ := os.Getwd()
	migrationPath := filepath.Join(wd, "app", "migrations")
	if err := goose.Up(db, migrationPath); err != nil {
		panic(err)
	}

	return &UsersPsqlStorage{
		Log:       log,
		DB:        db,
		TableName: tableName,
	}
}

func (u *UsersPsqlStorage) Close() {
	if err := u.DB.Close(); err != nil {
		panic(err)
	}
}

// GetUsers implements app.IUsersStorage.
func (u *UsersPsqlStorage) GetUsers(ctx context.Context) ([]models.User, error) {
	const op = "storage.users.psql.GetUsers"
	log := u.Log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return nil, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	query := fmt.Sprintf("SELECT * FROM %s;", u.TableName)
	rows, err := u.DB.QueryContext(ctx, query)
	if err != nil {
		log.Error("Error getting rows", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var bufUser models.User
	users := make([]models.User, 0, 10)
	for rows.Next() {
		if err := rows.Scan(&bufUser.Id, &bufUser.Login, &bufUser.Password, &bufUser.Role); err != nil {
			log.Warn("Error scanning row", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		users = append(users, bufUser)
	}

	log.Info("Users fetched successfully", slog.Int("count", len(users)))
	return users, nil
}

// GetUserById implements app.IUsersStorage.
func (u *UsersPsqlStorage) GetUserById(ctx context.Context, uid uuid.UUID) (models.User, error) {
	const op = "storage.users.psql.GetUserById"
	log := u.Log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	var user models.User
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1;", u.TableName)
	err := u.DB.QueryRowContext(ctx, query, uid).Scan(&user.Id, &user.Login, &user.Password, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User doesn't exist", sl.Err(storageerrors.ErrNotFound), slog.String("user_id", uid.String()))
			return models.User{}, fmt.Errorf("%s: %w", op, storageerrors.ErrNotFound)
		}

		log.Error("Error scanning row", sl.Err(err), slog.String("user_id", uid.String()))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("User fetched successfully", slog.String("user_id", user.Id.String()))
	return user, nil
}

// Insert implements app.IUsersStorage.
func (u *UsersPsqlStorage) Insert(ctx context.Context, user models.User) (models.User, error) {
	const op = "storage.users.psql.Insert"
	log := u.Log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	query := fmt.Sprintf("INSERT INTO %s (id, login, password, role) VALUES ($1, $2, $3, $4);", u.TableName)
	_, err := u.DB.ExecContext(ctx, query, user.Id, user.Login, user.Password, user.Role)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			log.Warn("User already exists", sl.Err(storageerrors.ErrAlreadyExists), slog.String("user_id", user.Id.String()))
			return models.User{}, fmt.Errorf("%s: %w", op, storageerrors.ErrAlreadyExists)
		}

		log.Error("Error inserting user", sl.Err(err), slog.String("user_id", user.Id.String()))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("User inserted successfully", slog.String("user_id", user.Id.String()))
	return user, nil
}

// Update implements app.IUsersStorage.
func (u *UsersPsqlStorage) Update(ctx context.Context, uid uuid.UUID, user models.User) (models.User, error) {
	const op = "storage.users.psql.Update"
	log := u.Log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	query := fmt.Sprintf("UPDATE %s SET login = $1, password = $2, role = $3 WHERE id = $4;", u.TableName)
	result, err := u.DB.ExecContext(ctx, query, user.Login, user.Password, user.Role, uid)
	if err != nil {
		log.Error("Error updating user", sl.Err(err), slog.String("user_id", uid.String()))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Error("Zero users affected", slog.String("user_id", uid.String()))
		return models.User{}, fmt.Errorf("%s: %w", op, storageerrors.ErrNotFound)
	}

	log.Info("User updated successfully", slog.String("user_id", uid.String()))
	return user, nil
}

// Delete implements app.IUsersStorage.
func (u *UsersPsqlStorage) Delete(ctx context.Context, uid uuid.UUID) (models.User, error) {
	const op = "storage.users.psql.Delete"
	log := u.Log.With("op", op)

	select {
	case <-ctx.Done():
		log.Info("Context cancelled", sl.Err(ctx.Err()))
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	userForReturn, err := u.GetUserById(ctx, uid)
	if err != nil {
		if errors.Is(err, storageerrors.ErrNotFound) {
			log.Error("User doesn't exist", sl.Err(storageerrors.ErrNotFound), slog.String("user_id", uid.String()))
			return models.User{}, fmt.Errorf("%s: %w", op, storageerrors.ErrNotFound)
		}

		log.Error("Error retrieving user before deleting", sl.Err(err), slog.String("user_id", uid.String()))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1;", u.TableName)
	if _, err := u.DB.ExecContext(ctx, query, uid); err != nil {
		log.Error("Error deleting user", sl.Err(err), slog.String("user_id", uid.String()))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("User deleted successfully", slog.String("user_id", uid.String()))
	return userForReturn, nil
}
