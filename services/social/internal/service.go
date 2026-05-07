package internal

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

var ErrInvalidGift = errors.New("invalid gift type")

type Friendship struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserA     uuid.UUID `db:"user_a" json:"user_a"`
	UserB     uuid.UUID `db:"user_b" json:"user_b"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Gift struct {
	ID        uuid.UUID `db:"id" json:"id"`
	FromUser  uuid.UUID `db:"from_user" json:"from_user"`
	ToUser    uuid.UUID `db:"to_user" json:"to_user"`
	Type      string    `db:"type" json:"type"`
	Claimed   bool      `db:"claimed" json:"claimed"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Service struct {
	db  *sqlx.DB
	rdb *redis.Client
}

func New(db *sqlx.DB, rdb *redis.Client) *Service { return &Service{db: db, rdb: rdb} }

func (s *Service) Request(ctx context.Context, from, to uuid.UUID) (*Friendship, error) {
	if from == to {
		return nil, errors.New("self friendship")
	}
	a, b := from, to
	if a.String() > b.String() {
		a, b = b, a
	}
	f := &Friendship{}
	err := s.db.GetContext(ctx, f, `
		INSERT INTO friendships(user_a, user_b, status)
		VALUES($1,$2,'pending')
		ON CONFLICT (user_a, user_b) DO UPDATE SET status=friendships.status
		RETURNING *`, a, b)
	if err != nil {
		return nil, err
	}
	_ = s.publish(ctx, "friend:request", map[string]any{"owner_id": to, "from": from})
	return f, nil
}

func (s *Service) Accept(ctx context.Context, user, requestID uuid.UUID) (*Friendship, error) {
	f := &Friendship{}
	err := s.db.GetContext(ctx, f, `
		UPDATE friendships SET status='accepted'
		WHERE id=$1 AND (user_a=$2 OR user_b=$2)
		RETURNING *`, requestID, user)
	return f, err
}

func (s *Service) ListFriends(ctx context.Context, user uuid.UUID) ([]Friendship, error) {
	var out []Friendship
	err := s.db.SelectContext(ctx, &out, `
		SELECT * FROM friendships
		WHERE (user_a=$1 OR user_b=$1) AND status='accepted'
		ORDER BY created_at DESC`, user)
	return out, err
}

func (s *Service) SendGift(ctx context.Context, from, to uuid.UUID, kind string) (*Gift, error) {
	if kind != "food" && kind != "toy" && kind != "medicine" {
		return nil, ErrInvalidGift
	}
	g := &Gift{}
	err := s.db.GetContext(ctx, g, `
		INSERT INTO gifts(from_user, to_user, type)
		VALUES($1,$2,$3) RETURNING *`, from, to, kind)
	if err != nil {
		return nil, err
	}
	_ = s.publish(ctx, "gift:received", map[string]any{
		"owner_id": to, "from": from, "type": kind, "gift_id": g.ID,
	})
	return g, nil
}

func (s *Service) Inbox(ctx context.Context, user uuid.UUID) ([]Gift, error) {
	var out []Gift
	err := s.db.SelectContext(ctx, &out, `
		SELECT * FROM gifts WHERE to_user=$1 AND claimed=false
		ORDER BY created_at DESC LIMIT 100`, user)
	return out, err
}

func (s *Service) Claim(ctx context.Context, user, giftID uuid.UUID) (*Gift, error) {
	g := &Gift{}
	err := s.db.GetContext(ctx, g, `
		UPDATE gifts SET claimed=true
		WHERE id=$1 AND to_user=$2 AND claimed=false
		RETURNING *`, giftID, user)
	return g, err
}

func (s *Service) publish(ctx context.Context, channel string, payload any) error {
	b, _ := json.Marshal(payload)
	return s.rdb.Publish(ctx, channel, b).Err()
}
