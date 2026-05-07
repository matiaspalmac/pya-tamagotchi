package game

import (
	"testing"
	"time"
)

func TestApplyTicks_Decay(t *testing.T) {
	s := Stats{Hunger: 100, Happy: 100, Energy: 100, Health: 100}
	out, dead := ApplyTicks(s, 5)
	if dead {
		t.Fatal("should not die")
	}
	if out.Hunger != 90 {
		t.Errorf("hunger: got %d want 90", out.Hunger)
	}
	if out.Happy != 95 {
		t.Errorf("happy: got %d want 95", out.Happy)
	}
}

func TestApplyTicks_HungerCausesHealthDrain(t *testing.T) {
	s := Stats{Hunger: 22, Happy: 50, Energy: 50, Health: 100}
	out, _ := ApplyTicks(s, 5)
	// after tick 1: hunger=20 (no drain). tick2: 18 (<20, drain). tick3-5: drain
	if out.Health == 100 {
		t.Errorf("health should drop, got %d", out.Health)
	}
}

func TestApplyTicks_Death(t *testing.T) {
	s := Stats{Hunger: 0, Happy: 0, Energy: 0, Health: 3}
	_, dead := ApplyTicks(s, 10)
	if !dead {
		t.Fatal("expected death")
	}
}

func TestEvolveStage(t *testing.T) {
	now := time.Now()
	cases := []struct {
		born time.Time
		xp   int
		want Stage
	}{
		{now.Add(-30 * time.Minute), 0, StageEgg},
		{now.Add(-2 * time.Hour), 0, StageBaby},
		{now.Add(-25 * time.Hour), 100, StageTeen},
		{now.Add(-73 * time.Hour), 500, StageAdult},
		{now.Add(-8 * 24 * time.Hour), 0, StageElder},
	}
	for _, c := range cases {
		got := EvolveStage(StageEgg, c.born, c.xp)
		if got != c.want {
			t.Errorf("born=%v xp=%d: got %s want %s", c.born, c.xp, got, c.want)
		}
	}
}
