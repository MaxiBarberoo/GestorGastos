package models

import "database/sql"

// RunMigrations ensures every table required by the app exists.
func RunMigrations(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
		`CREATE TABLE IF NOT EXISTS expenses (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			tag TEXT NOT NULL,
			amount NUMERIC(12,2) NOT NULL,
			expense_date DATE NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
		`CREATE TABLE IF NOT EXISTS monthly_expenses (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			tag TEXT NOT NULL,
			amount NUMERIC(12,2) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			last_applied_at TIMESTAMPTZ,
			last_applied_expense_id BIGINT REFERENCES expenses(id)
		);`,
		`ALTER TABLE monthly_expenses
			ADD COLUMN IF NOT EXISTS last_applied_at TIMESTAMPTZ;`,
		`ALTER TABLE monthly_expenses
			ADD COLUMN IF NOT EXISTS last_applied_expense_id BIGINT REFERENCES expenses(id);`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}
