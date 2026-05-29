package defense

import (
	"testing"
	"time"
)

func TestEffectiveFallSpeed(t *testing.T) {
	base := 1.0
	if got := EffectiveFallSpeed(base, 0); got != 1.0 {
		t.Fatalf("t=0: got %g want 1.0", got)
	}
	if got := EffectiveFallSpeed(base, 120*time.Second); got != 1.5 {
		t.Fatalf("t=120s: got %g want 1.5", got)
	}
	if got := EffectiveFallSpeed(base, 300*time.Second); got != 2.0 {
		t.Fatalf("t=300s: got %g want 2.0 (capped)", got)
	}
}

func TestEffectiveSpawnInterval(t *testing.T) {
	base := 1.2
	if got := EffectiveSpawnInterval(base, 0); got != 1.2 {
		t.Fatalf("t=0: got %g want 1.2", got)
	}
	got60 := EffectiveSpawnInterval(base, 60*time.Second)
	want60 := 1.2 / 1.05
	if got60 < want60-0.001 || got60 > want60+0.001 {
		t.Fatalf("t=60s: got %g want ~%g", got60, want60)
	}
}
