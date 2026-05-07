package game

import "time"

const (
	StatMax = 100

	HungerDecayPerTick = 2
	HappyDecayPerTick  = 1
	EnergyDecayPerTick = 1
	HealthDecayWhenLow = 1
	LowHungerThreshold = 20

	FeedHungerGain = 30
	FeedHappyGain  = 5
	PlayHappyGain  = 25
	PlayEnergyCost = 15
	HealHealthGain = 50

	FeedCooldown = 60 * time.Second
	PlayCooldown = 30 * time.Second
)

type Stage string

const (
	StageEgg   Stage = "egg"
	StageBaby  Stage = "baby"
	StageTeen  Stage = "teen"
	StageAdult Stage = "adult"
	StageElder Stage = "elder"
)

type Stats struct {
	Hunger int
	Happy  int
	Energy int
	Health int
}

func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > StatMax {
		return StatMax
	}
	return v
}

// ApplyTicks aplica N ticks de decay. Devuelve stats nuevas + si murió.
func ApplyTicks(s Stats, ticks int) (Stats, bool) {
	for i := 0; i < ticks; i++ {
		s.Hunger = clamp(s.Hunger - HungerDecayPerTick)
		s.Happy = clamp(s.Happy - HappyDecayPerTick)
		s.Energy = clamp(s.Energy - EnergyDecayPerTick)
		if s.Hunger < LowHungerThreshold {
			s.Health = clamp(s.Health - HealthDecayWhenLow)
		}
		if s.Health <= 0 {
			return s, true
		}
	}
	return s, false
}

func TicksSince(last time.Time, interval time.Duration) int {
	if interval <= 0 {
		return 0
	}
	d := time.Since(last)
	if d <= 0 {
		return 0
	}
	return int(d / interval)
}

// EvolveStage decide stage según vida y xp.
func EvolveStage(current Stage, bornAt time.Time, xp int) Stage {
	age := time.Since(bornAt)
	switch {
	case age >= 7*24*time.Hour:
		return StageElder
	case age >= 72*time.Hour && xp >= 500:
		return StageAdult
	case age >= 24*time.Hour && xp >= 100:
		return StageTeen
	case age >= time.Hour:
		return StageBaby
	default:
		return StageEgg
	}
}
