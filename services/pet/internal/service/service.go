package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/tamagotchi/pet/internal/game"
	"github.com/tamagotchi/pet/internal/repo"
)

var (
	ErrCooldown    = errors.New("action on cooldown")
	ErrLowEnergy   = errors.New("not enough energy")
	ErrDead        = errors.New("pet is dead")
	ErrNotOwner    = errors.New("not owner")
)

type Service struct {
	pets         *repo.PetRepo
	rdb          *redis.Client
	tickInterval time.Duration
}

func New(pets *repo.PetRepo, rdb *redis.Client, tickSec int) *Service {
	return &Service{pets: pets, rdb: rdb, tickInterval: time.Duration(tickSec) * time.Second}
}

func (s *Service) Create(ctx context.Context, owner uuid.UUID, name, species string) (*repo.Pet, error) {
	p, err := s.pets.Create(ctx, owner, name, species)
	if err != nil {
		return nil, err
	}
	_ = s.publish(ctx, "pet:created", map[string]any{"pet_id": p.ID, "owner_id": owner})
	return p, nil
}

func (s *Service) Mine(ctx context.Context, owner uuid.UUID) ([]repo.Pet, error) {
	return s.pets.ByOwner(ctx, owner)
}

func (s *Service) Get(ctx context.Context, owner, id uuid.UUID) (*repo.Pet, error) {
	p, err := s.pets.ByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.OwnerID != owner {
		return nil, ErrNotOwner
	}
	return s.applyPendingTicks(ctx, p)
}

func (s *Service) applyPendingTicks(ctx context.Context, p *repo.Pet) (*repo.Pet, error) {
	if p.DiedAt != nil {
		return p, nil
	}
	ticks := game.TicksSince(p.LastTickAt, s.tickInterval)
	if ticks == 0 {
		return p, nil
	}
	stats, dead := game.ApplyTicks(game.Stats{
		Hunger: p.Hunger, Happy: p.Happy, Energy: p.Energy, Health: p.Health,
	}, ticks)
	p.Hunger, p.Happy, p.Energy, p.Health = stats.Hunger, stats.Happy, stats.Energy, stats.Health
	p.LastTickAt = p.LastTickAt.Add(time.Duration(ticks) * s.tickInterval)
	if dead {
		now := time.Now()
		p.DiedAt = &now
	}
	newStage := string(game.EvolveStage(game.Stage(p.Stage), p.BornAt, p.XP))
	if newStage != p.Stage {
		p.Stage = newStage
		_ = s.publish(ctx, "pet:evolved", map[string]any{"pet_id": p.ID, "stage": newStage})
	}
	if err := s.pets.UpdateStats(ctx, p); err != nil {
		return nil, err
	}
	_ = s.publish(ctx, "pet:tick", map[string]any{
		"pet_id": p.ID, "stats": stats, "stage": p.Stage, "dead": dead,
	})
	return p, nil
}

func (s *Service) Feed(ctx context.Context, owner, id uuid.UUID) (*repo.Pet, error) {
	return s.act(ctx, owner, id, "feed", game.FeedCooldown, func(p *repo.Pet) error {
		p.Hunger = clamp(p.Hunger + game.FeedHungerGain)
		p.Happy = clamp(p.Happy + game.FeedHappyGain)
		p.XP += 5
		return nil
	})
}

func (s *Service) Play(ctx context.Context, owner, id uuid.UUID) (*repo.Pet, error) {
	return s.act(ctx, owner, id, "play", game.PlayCooldown, func(p *repo.Pet) error {
		if p.Energy < game.PlayEnergyCost {
			return ErrLowEnergy
		}
		p.Happy = clamp(p.Happy + game.PlayHappyGain)
		p.Energy = clamp(p.Energy - game.PlayEnergyCost)
		p.XP += 10
		return nil
	})
}

func (s *Service) Sleep(ctx context.Context, owner, id uuid.UUID) (*repo.Pet, error) {
	return s.act(ctx, owner, id, "sleep", 0, func(p *repo.Pet) error {
		p.Energy = clamp(p.Energy + 50)
		return nil
	})
}

func (s *Service) Heal(ctx context.Context, owner, id uuid.UUID) (*repo.Pet, error) {
	return s.act(ctx, owner, id, "heal", 0, func(p *repo.Pet) error {
		p.Health = clamp(p.Health + game.HealHealthGain)
		return nil
	})
}

func (s *Service) act(ctx context.Context, owner, id uuid.UUID, kind string, cd time.Duration, mut func(*repo.Pet) error) (*repo.Pet, error) {
	p, err := s.Get(ctx, owner, id)
	if err != nil {
		return nil, err
	}
	if p.DiedAt != nil {
		return nil, ErrDead
	}
	if cd > 0 {
		key := fmt.Sprintf("cd:%s:%s", id, kind)
		ok, err := s.rdb.SetNX(ctx, key, "1", cd).Result()
		if err == nil && !ok {
			return nil, ErrCooldown
		}
	}
	if err := mut(p); err != nil {
		return nil, err
	}
	if err := s.pets.UpdateStats(ctx, p); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]any{"action": kind})
	_ = s.pets.AddEvent(ctx, p.ID, kind, payload)
	_ = s.publish(ctx, "pet:event", map[string]any{
		"pet_id": p.ID, "type": kind, "stats": map[string]int{
			"hunger": p.Hunger, "happy": p.Happy, "energy": p.Energy, "health": p.Health,
		},
	})
	return p, nil
}

func (s *Service) Events(ctx context.Context, owner, id uuid.UUID, limit int) ([]repo.Event, error) {
	if _, err := s.Get(ctx, owner, id); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.pets.Events(ctx, id, limit)
}

func (s *Service) publish(ctx context.Context, channel string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.rdb.Publish(ctx, channel, b).Err()
}

func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > game.StatMax {
		return game.StatMax
	}
	return v
}
