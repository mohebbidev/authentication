package postgres

import (
	"authpractice/internal/domain"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, full_name, date_of_birth, email, hash_password, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.FullName,
		user.DateOfBirth,
		user.Email,
		user.HashPassword,
		user.Active,
		user.CreatedAt,
	)
	if err != nil {
		return domain.MapPostgresError(err)
	}
	return nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, full_name, date_of_birth, email, hash_password, is_active, created_at
		FROM users
		WHERE email = $1
	`
	user, err := scanUser(r.db.QueryRow(ctx, query, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrorsInstance.UserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	query := `
		SELECT id, full_name, date_of_birth, email, hash_password, is_active, created_at
		FROM users
		WHERE id = $1
	`
	user, err := scanUser(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrorsInstance.UserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id domain.UserID, newHash string) error {
	query := `UPDATE users SET hash_password = $1 WHERE id = $2`
	tag, err := r.db.Exec(ctx, query, newHash, id)
	if err != nil {
		return fmt.Errorf("update password: %w", domain.MapPostgresError(err))
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrorsInstance.UserNotFound
	}
	return nil
}

func (r *UserRepo) SetActive(ctx context.Context, id domain.UserID, active bool) error {
	query := `UPDATE users SET is_active = $1 WHERE id = $2`
	tag, err := r.db.Exec(ctx, query, active, id)
	if err != nil {
		return fmt.Errorf("set active: %w", domain.MapPostgresError(err))
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrorsInstance.UserNotFound
	}
	return nil
}

// scanUser is a shared helper so column order is defined exactly once.
func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(
		&u.ID,
		&u.FullName,
		&u.DateOfBirth,
		&u.Email,
		&u.HashPassword,
		&u.Active,
		&u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}