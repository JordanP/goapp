package store

const createTableUsers = `
CREATE EXTENSION IF not EXISTS pgcrypto;
CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	login VARCHAR(64) NOT NULL,
    password VARCHAR(255) NOT NULL,
	email VARCHAR(64) NOT NULL,
	role  text NOT NULL,
	created_at timestamp WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT unq_login UNIQUE(login),
    CONSTRAINT unq_email UNIQUE(email)
)`

const insertUser = `
INSERT INTO users (login, password, email, role)
VALUES ($1, $2, $3, $4)
RETURNING id, login, email, role, created_at
`

const deleteUser = `
DELETE FROM users
`

const selectUser = `
SELECT id, login, password, email, role, created_at FROM users
`

const selectAllUsers = `
SELECT id, login, email, role, created_at FROM users ORDER BY created_at asc;
`

const deleteAllUsers = `
TRUNCATE TABLE users CASCADE
`
