package userspsqlstorage_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"usersmanager/internal/domain/models"
	userspsqlstorage "usersmanager/internal/storage/users/psql"
	"usersmanager/pkg/lib/logger/handler/slogdiscard"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func newTestStorage(t *testing.T) (*userspsqlstorage.UsersPsqlStorage, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %s", err)
	}
	storage := &userspsqlstorage.UsersPsqlStorage{
		Log:       slogdiscard.NewDiscardLogger(),
		DB:        db,
		TableName: "users",
	}
	cleanup := func() { db.Close() }
	return storage, mock, cleanup
}

func TestGetUsers_ContextCanceled(t *testing.T) {
	storage, mock, cleanup := newTestStorage(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := storage.GetUsers(ctx)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetUsers_QueryError(t *testing.T) {
	storage, mock, cleanup := newTestStorage(t)
	defer cleanup()

	mock.ExpectQuery("SELECT \\* FROM users;").WillReturnError(sql.ErrConnDone)
	_, err := storage.GetUsers(context.Background())
	if err == nil || !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("expected sql.ErrConnDone, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetUsers_ScanError(t *testing.T) {
	storage, mock, cleanup := newTestStorage(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "login", "password", "role"}).
		AddRow("bad-uuid", "login", "pass", "role")
	mock.ExpectQuery("SELECT \\* FROM users;").WillReturnRows(rows)
	_, err := storage.GetUsers(context.Background())
	if err == nil {
		t.Fatal("expected error from Scan")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetUsers_Empty(t *testing.T) {
	storage, mock, cleanup := newTestStorage(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "login", "password", "role"})
	mock.ExpectQuery("SELECT \\* FROM users;").WillReturnRows(rows)
	users, err := storage.GetUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

func TestGetUserById_ScanError(t *testing.T) {
	storage, mock, cleanup := newTestStorage(t)
	defer cleanup()
	id := uuid.New()
	mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\$1;").
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password", "role"}).
			AddRow("bad-uuid", "login", "pass", "role"))
	_, err := storage.GetUserById(context.Background(), id)
	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestInsert_OtherDBError(t *testing.T) {
	storage, mock, cleanup := newTestStorage(t)
	defer cleanup()

	user := models.User{Id: uuid.New(), Login: "user", Password: "pass", Role: "role"}
	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.Id, user.Login, user.Password, user.Role).
		WillReturnError(sql.ErrConnDone)
	_, err := storage.Insert(context.Background(), user)
	if err == nil || !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("expected sql.ErrConnDone, got %v", err)
	}
}

func TestUpdate_DBError(t *testing.T) {
	storage, mock, cleanup := newTestStorage(t)
	defer cleanup()
	user := models.User{Id: uuid.New(), Login: "user", Password: "pass", Role: "role"}
	mock.ExpectExec("UPDATE users").
		WithArgs(user.Login, user.Password, user.Role, user.Id).
		WillReturnError(sql.ErrConnDone)
	_, err := storage.Update(context.Background(), user.Id, user)
	if err == nil || !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("expected sql.ErrConnDone, got %v", err)
	}
}

func TestDelete_GetByIdError(t *testing.T) {
	storage, mock, cleanup := newTestStorage(t)
	defer cleanup()
	id := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM users WHERE id = $1;")).
		WithArgs(id).WillReturnError(sql.ErrConnDone)
	_, err := storage.Delete(context.Background(), id)
	if err == nil || !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("expected get error, got %v", err)
	}
}

func TestDelete_ExecError(t *testing.T) {
	storage, mock, cleanup := newTestStorage(t)
	defer cleanup()
	id := uuid.New()

	row := sqlmock.NewRows([]string{"id", "login", "password", "role"}).
		AddRow(id, "user1", "pass1", "admin")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM users WHERE id = $1;")).
		WithArgs(id).WillReturnRows(row)
	mock.ExpectExec("DELETE FROM users").
		WithArgs(id).WillReturnError(sql.ErrConnDone)
	_, err := storage.Delete(context.Background(), id)
	if err == nil || !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("expected delete error, got %v", err)
	}
}
