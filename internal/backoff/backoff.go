package backoff

import "time"

type Jitter func(upper time.Duration) time.Duration

type Policy struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
	Jitter    Jitter
}

func (p Policy) Delay(attempt int, retryAfter time.Duration) time.Duration {
	ceiling := p.ceiling(attempt)
	jittered := p.Jitter(ceiling)
	honored := min(retryAfter, p.MaxDelay)
	return max(jittered, honored)
}

func (p Policy) ceiling(attempt int) time.Duration {
	delay := p.BaseDelay
	for range attempt - 1 {
		delay *= 2
		if delay >= p.MaxDelay {
			return p.MaxDelay
		}
	}
	return min(delay, p.MaxDelay)
}
