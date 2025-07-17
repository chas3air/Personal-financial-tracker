-- +goose Up
-- Описание: Эта миграция создает таблицу users
CREATE TABLE users (
    id UUID PRIMARY KEY,
    login VARCHAR(50) NOT NULL,
    password VARCHAR(50) NOT NULL,
    role VARCHAR(100) NOT NULL
);

-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
-- Описание: Эта миграция удаляет таблицу users
DROP TABLE users;

-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd