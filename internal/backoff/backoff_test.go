package backoff

import (
	"testing"
	"time"
)

func identity(upper time.Duration) time.Duration {
	return upper
}

func TestDelayGrowsExponentiallyUpToTheCeiling(t *testing.T) {
	policy := Policy{BaseDelay: 100 * time.Millisecond, MaxDelay: time.Second, Jitter: identity}
	want := []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond, 800 * time.Millisecond, time.Second, time.Second}
	for i, expected := range want {
		if got := policy.Delay(i+1, 0); got != expected {
			t.Fatalf("tentativa %d: %v, esperado %v", i+1, got, expected)
		}
	}
}

func TestDelayHonorsRetryAfterCappedByMaxDelay(t *testing.T) {
	policy := Policy{BaseDelay: time.Millisecond, MaxDelay: 5 * time.Second, Jitter: func(time.Duration) time.Duration { return 0 }}
	if got := policy.Delay(1, 2*time.Second); got != 2*time.Second {
		t.Fatalf("Retry-After de 2s: %v", got)
	}
	if got := policy.Delay(1, time.Minute); got != 5*time.Second {
		t.Fatalf("Retry-After de 60s deveria ser limitado a 5s: %v", got)
	}
}
