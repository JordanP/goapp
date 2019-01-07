package store

const createTableCompanies = `
CREATE EXTENSION IF not EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS companies (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
	created_at timestamp WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT unq_name UNIQUE(name)
);

CREATE TABLE IF NOT EXISTS users_companies (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID references companies(id) ON DELETE CASCADE,
	user_id UUID references users(id) ON DELETE CASCADE,
    CONSTRAINT unq_set UNIQUE(company_id,user_id)
)`

const deleteAllCompanies = `
TRUNCATE TABLE companies CASCADE
`

const insertCompany = `
INSERT INTO companies (name)
VALUES ($1)
RETURNING id, name, created_at
`

const selectCompany = `
SELECT id, name, created_at FROM companies`

const deleteCompany = `
DELETE FROM companies
`

const insertUserInCompany = `
INSERT INTO users_companies (company_id, user_id)
VALUES ($1, $2)
`

const selectCompanyUsers = `
SELECT u.id, u.login, u.email, u.role, u.created_at FROM companies c
JOIN users_companies uc ON uc.company_id = c.id
JOIN users u ON uc.user_id = u.id
WHERE c.id = $1
`

// docker exec -i -t e32a07615cec psql -h localhost -U myuser --dbname=myuser
