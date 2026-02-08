package usecase

import (
	"math/rand"
	"time"
)

type Rand interface {
	Intn(n int) int
}

type Sleeper func(time.Duration)

func (s *MatchingService) randIntn(n int) int {
	if s != nil && s.Rand != nil {
		return s.Rand.Intn(n)
	}
	return rand.Intn(n)
}

func (s *MatchingService) sleep(d time.Duration) {
	if d <= 0 {
		return
	}
	if s != nil && s.Sleep != nil {
		s.Sleep(d)
		return
	}
	time.Sleep(d)
}
