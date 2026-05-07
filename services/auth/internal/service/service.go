package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/tamagotchi/auth/internal/repo"
)

var (
	ErrInvalidCreds = errors.New("invalid credentials")
	ErrConflict     = errors.New("email or username taken")
)

type Service struct {
	users      *repo.UserRepo
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(users *repo.UserRepo, secret string, accessMin, refreshDays int) *Service {
	return &Service{
		users:      users,
		secret:     []byte(secret),
		accessTTL:  time.Duration(accessMin) * time.Minute,
		refreshTTL: time.Duration(refreshDays) * 24 * time.Hour,
	}
}

type Claims struct {
	UserID uuid.UUID `json:"uid"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
}

func (s *Service) Register(ctx context.Context, email, username, password string) (*repo.User, *TokenPair, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, nil, err
	}
	u, err := s.users.Create(ctx, email, username, string(hash))
	if err != nil {
		return nil, nil, ErrConflict
	}
	tp, err := s.issue(u.ID)
	return u, tp, err
}

func (s *Service) Login(ctx context.Context, email, password string) (*repo.User, *TokenPair, error) {
	u, err := s.users.ByEmail(ctx, email)
	if err != nil {
		return nil, nil, ErrInvalidCreds
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCreds
	}
	tp, err := s.issue(u.ID)
	return u, tp, err
}

func (s *Service) Refresh(refresh string) (*TokenPair, error) {
	c, err := s.parse(refresh)
	if err != nil {
		return nil, err
	}
	return s.issue(c.UserID)
}

func (s *Service) Verify(token string) (*Claims, error) { return s.parse(token) }

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*repo.User, error) {
	return s.users.ByID(ctx, id)
}

func (s *Service) issue(uid uuid.UUID) (*TokenPair, error) {
	access, err := s.sign(uid, s.accessTTL)
	if err != nil {
		return nil, err
	}
	refresh, err := s.sign(uid, s.refreshTTL)
	if err != nil {
		return nil, err
	}
	return &TokenPair{Access: access, Refresh: refresh}, nil
}

func (s *Service) sign(uid uuid.UUID, ttl time.Duration) (string, error) {
	c := Claims{
		UserID: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(s.secret)
}

func (s *Service) parse(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("bad alg")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return c, nil
}
