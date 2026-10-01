package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cargoprener/backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *Store) CreateUser(ctx context.Context, input domain.NewUser) (domain.User, error) {
	role := input.Role
	if role == "" {
		role = domain.RoleBusiness
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO users (business_name, contact_name, email, password_hash, region, role)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, business_name, contact_name, email, password_hash, region, role, created_at
	`, input.BusinessName, input.ContactName, input.Email, input.PasswordHash, input.Region, role)

	user, err := scanUser(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, domain.ErrEmailTaken
		}
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}
	return user, nil
}

func (s *Store) FindUserByEmail(ctx context.Context, email string) (domain.User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, business_name, contact_name, email, password_hash, region, role, created_at
		FROM users
		WHERE email = $1
	`, email)
	user, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("find user by email: %w", err)
	}
	return user, nil
}

func (s *Store) FindUserByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, business_name, contact_name, email, password_hash, region, role, created_at
		FROM users
		WHERE id = $1
	`, id)
	user, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("find user by id: %w", err)
	}
	return user, nil
}

func (s *Store) EnsureAdmin(ctx context.Context, email, password string, hash func(string) (string, error)) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return nil
	}
	_, err := s.FindUserByEmail(ctx, email)
	if err == nil {
		_, err = s.pool.Exec(ctx, `UPDATE users SET role = $1 WHERE email = $2 AND role <> $1`, domain.RoleAdmin, email)
		if err != nil {
			return fmt.Errorf("promote admin: %w", err)
		}
		return nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	passwordHash, err := hash(password)
	if err != nil {
		return err
	}
	_, err = s.CreateUser(ctx, domain.NewUser{
		BusinessName: "CargoPrener",
		ContactName:  "Admin",
		Email:        email,
		PasswordHash: passwordHash,
		Region:       "Platform",
		Role:         domain.RoleAdmin,
	})
	return err
}

func scanUser(row pgx.Row) (domain.User, error) {
	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.BusinessName,
		&user.ContactName,
		&user.Email,
		&user.PasswordHash,
		&user.Region,
		&user.Role,
		&user.CreatedAt,
	)
	return user, err
}
