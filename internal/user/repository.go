package user

import (
	"context"
	"database/sql"
)

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash, name, surname string) (*User, error)
	ByEmail(ctx context.Context, email string) (*User, error)
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

func (r *PostgresUserRepository) ByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}

	err := r.db.QueryRowContext(
		ctx,
		`
		SELECT id, email, password_hash, name, surname, role, created_at
		FROM users
		WHERE email = $1
		`,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.Surname,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, email, passwordHash, name, surname string) (*User, error) {
	user := &User{}

	err := r.db.QueryRowContext(
		ctx,
		`
		INSERT INTO users (email, password_hash, name, surname)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, password_hash, name, surname, role, created_at
		`,
		email, passwordHash, name, surname,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.Surname,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}
