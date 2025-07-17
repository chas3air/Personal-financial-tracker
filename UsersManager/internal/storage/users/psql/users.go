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
	log       *slog.Logger
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
		log:       log,
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
	const op = "storage.user.psql.GetUsers"
	log := u.log.With(
		"op", op,
	)

	select {
	case <-ctx.Done():
		log.Info("context is over")
		return nil, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	rows, err := u.DB.QueryContext(ctx, `
		SELECT * FROM $1;
	`, u.TableName)
	if err != nil {
		log.Error("Error getting rows", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var bufUser models.User
	var users = make([]models.User, 0, 10)
	for rows.Next() {
		if err := rows.Scan(&bufUser.Id, &bufUser.Login, &bufUser.Password, &bufUser.Role); err != nil {
			log.Warn("Error scanning row", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		users = append(users, bufUser)
	}

	return users, nil
}

// GetUserById implements app.IUsersStorage.
func (u *UsersPsqlStorage) GetUserById(ctx context.Context, uid uuid.UUID) (models.User, error) {
	const op = "storage.user.psql.GetUserById"
	log := u.log.With(
		"op", op,
	)

	select {
	case <-ctx.Done():
		log.Info("context is over")
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	var user models.User

	err := u.DB.QueryRowContext(ctx, `
		SELECT * FROM $1 WHERE id = $2;
	`, u.TableName, uid).Scan(&user.Id, &user.Login, &user.Password, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("User doesn't exists", sl.Err(storageerrors.ErrNotFound))
			return models.User{}, fmt.Errorf("%s: %w", op, storageerrors.ErrNotFound)
		}

		log.Error("Error scaning row", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

// Insert implements app.IUsersStorage.
func (u *UsersPsqlStorage) Insert(ctx context.Context, user models.User) (models.User, error) {
	const op = "storage.user.psql.Insert"
	log := u.log.With(
		"op", op,
	)

	select {
	case <-ctx.Done():
		log.Info("context is over")
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	_, err := u.DB.ExecContext(ctx, `
		INSERT INTO $1
		VALUES ($2, $3, $4, $5);
	`, u.TableName, user.Id, user.Login, user.Password, user.Role)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			log.Warn("User already exists", sl.Err(storageerrors.ErrAlreadyExists))
			return models.User{}, fmt.Errorf("%s: %w", op, storageerrors.ErrAlreadyExists)
		}

		log.Error("Error inserting user", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

// Update implements app.IUsersStorage.
func (u *UsersPsqlStorage) Update(ctx context.Context, uid uuid.UUID, user models.User) (models.User, error) {
	const op = "storage.user.psql.Update"
	log := u.log.With(
		"op", op,
	)

	select {
	case <-ctx.Done():
		log.Info("context is over")
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	result, err := u.DB.ExecContext(ctx, `
		UPDATE $1
		SET login = $2, password = $3, role = $4
		WHERE id = $5;
	`, u.TableName, user.Login, user.Password, user.Role, uid)
	if err != nil {
		log.Error("Error updating user", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Error("Zero users affected")
		return models.User{}, fmt.Errorf("%s: %w", op, storageerrors.ErrNotFound)
	}

	return user, nil
}

// Delete implements app.IUsersStorage.
func (u *UsersPsqlStorage) Delete(ctx context.Context, uid uuid.UUID) (models.User, error) {
	const op = "storage.user.psql.Delete"
	log := u.log.With(
		"op", op,
	)

	select {
	case <-ctx.Done():
		log.Info("context is over")
		return models.User{}, fmt.Errorf("%s: %w", op, ctx.Err())
	default:
	}

	userForReturn, err := u.GetUserById(ctx, uid)
	if err != nil {
		if errors.Is(err, storageerrors.ErrNotFound) {
			log.Error("User doesn't exists", sl.Err(storageerrors.ErrNotFound))
			return models.User{}, fmt.Errorf("%s: %w", op, storageerrors.ErrNotFound)
		}

		log.Error("Error retrieving user before deleting", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	if _, err := u.DB.ExecContext(ctx, `
		DELETE FROM $1
		WHERE id = $2;
	`, u.TableName, uid); err != nil {
		log.Error("Error deleting user", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return userForReturn, nil
}
