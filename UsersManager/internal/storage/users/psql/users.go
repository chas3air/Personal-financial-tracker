package userspsqlstorage

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"usersmanager/internal/domain/models"

	"github.com/google/uuid"
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
	panic("unimplemented")
}

// GetUserById implements app.IUsersStorage.
func (u *UsersPsqlStorage) GetUserById(ctx context.Context, uid uuid.UUID) (models.User, error) {
	panic("unimplemented")
}

// Insert implements app.IUsersStorage.
func (u *UsersPsqlStorage) Insert(ctx context.Context, user models.User) (models.User, error) {
	panic("unimplemented")
}

// Update implements app.IUsersStorage.
func (u *UsersPsqlStorage) Update(ctx context.Context, uid uuid.UUID, user models.User) (models.User, error) {
	panic("unimplemented")
}

// Delete implements app.IUsersStorage.
func (u *UsersPsqlStorage) Delete(ctx context.Context, uid uuid.UUID) (models.User, error) {
	panic("unimplemented")
}
