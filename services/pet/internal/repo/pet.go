package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("pet not found")

type Pet struct {
	ID         uuid.UUID  `db:"id"          json:"id"`
	OwnerID    uuid.UUID  `db:"owner_id"    json:"owner_id"`
	Name       string     `db:"name"        json:"name"`
	Species    string     `db:"species"     json:"species"`
	Hunger     int        `db:"hunger"      json:"hunger"`
	Happy      int        `db:"happy"       json:"happy"`
	Energy     int        `db:"energy"      json:"energy"`
	Health     int        `db:"health"      json:"health"`
	XP         int        `db:"xp"          json:"xp"`
	Stage      string     `db:"stage"       json:"stage"`
	BornAt     time.Time  `db:"born_at"     json:"born_at"`
	LastTickAt time.Time  `db:"last_tick_at" json:"last_tick_at"`
	DiedAt     *time.Time `db:"died_at"     json:"died_at,omitempty"`
}

type PetRepo struct{ db *sqlx.DB }

func NewPetRepo(db *sqlx.DB) *PetRepo { return &PetRepo{db: db} }

func (r *PetRepo) Create(ctx context.Context, ownerID uuid.UUID, name, species string) (*Pet, error) {
	p := &Pet{}
	err := r.db.GetContext(ctx, p, `
		INSERT INTO pets(owner_id, name, species)
		VALUES($1,$2,$3) RETURNING *`, ownerID, name, species)
	return p, err
}

func (r *PetRepo) ByID(ctx context.Context, id uuid.UUID) (*Pet, error) {
	p := &Pet{}
	err := r.db.GetContext(ctx, p, `SELECT * FROM pets WHERE id=$1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func (r *PetRepo) ByOwner(ctx context.Context, ownerID uuid.UUID) ([]Pet, error) {
	var out []Pet
	err := r.db.SelectContext(ctx, &out, `SELECT * FROM pets WHERE owner_id=$1 ORDER BY born_at DESC`, ownerID)
	return out, err
}

func (r *PetRepo) UpdateStats(ctx context.Context, p *Pet) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE pets SET hunger=$1, happy=$2, energy=$3, health=$4,
		                xp=$5, stage=$6, last_tick_at=$7, died_at=$8
		WHERE id=$9`,
		p.Hunger, p.Happy, p.Energy, p.Health,
		p.XP, p.Stage, p.LastTickAt, p.DiedAt, p.ID)
	return err
}

func (r *PetRepo) AddEvent(ctx context.Context, petID uuid.UUID, eventType string, payload []byte) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO pet_events(pet_id, type, payload)
		VALUES($1,$2,$3::jsonb)`, petID, eventType, string(payload))
	return err
}

type Event struct {
	ID        uuid.UUID `db:"id"        json:"id"`
	PetID     uuid.UUID `db:"pet_id"    json:"pet_id"`
	Type      string    `db:"type"      json:"type"`
	Payload   []byte    `db:"payload"   json:"payload"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

func (r *PetRepo) Events(ctx context.Context, petID uuid.UUID, limit int) ([]Event, error) {
	var out []Event
	err := r.db.SelectContext(ctx, &out, `
		SELECT * FROM pet_events WHERE pet_id=$1
		ORDER BY created_at DESC LIMIT $2`, petID, limit)
	return out, err
}
