package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("not found")

type User struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

type UserRepo struct{ db *sqlx.DB }

func NewUserRepo(db *sqlx.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(ctx context.Context, email, username, hash string) (*User, error) {
	u := &User{}
	err := r.db.GetContext(ctx, u,
		`INSERT INTO users(email, username, password_hash)
		 VALUES($1,$2,$3) RETURNING *`,
		email, username, hash,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) ByEmail(ctx context.Context, email string) (*User, error) {
	u := &User{}
	err := r.db.GetContext(ctx, u, `SELECT * FROM users WHERE email=$1`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (r *UserRepo) ByID(ctx context.Context, id uuid.UUID) (*User, error) {
	u := &User{}
	err := r.db.GetContext(ctx, u, `SELECT * FROM users WHERE id=$1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}
