package repository

import (
	"context"
	"database/sql"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

type CreateUserParams struct {
	Email       string
	Username    string
	DisplayName string
	StatusID    string
}

func (r *UserRepository) CreateUser(ctx context.Context, params CreateUserParams) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users(email, username, display_name, account_status_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, params.Email, params.Username, params.DisplayName, params.StatusID).Scan(&id)

	return id, err
}

func (r *UserRepository) CreateCredential(ctx context.Context, userID, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_credentials (user_id, password_hash)
		VALUES ($1, $2)
	`, userID, passwordHash)
	return err
}

func (r *UserRepository) GetActiveStatusID(ctx context.Context) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx,
		`SELECT id FROM account_status WHERE status = 'active'`,
	).Scan(&id)
	return id, err
}

func (r *UserRepository) ExistsEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`,
		email,
	).Scan(&exists)
	return exists, err
}

func (r *UserRepository) ExistsUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`,
		username,
	).Scan(&exists)
	return exists, err
}
