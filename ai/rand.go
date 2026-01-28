package ai

import (
	"context"
	"math/rand"
	"time"

	"github.com/youryharchenko/go-mas/planning"
)

type RandomPolicy[S planning.State] struct {
	rng *rand.Rand
}

func NewRandom[S planning.State]() *RandomPolicy[S] {
	return &RandomPolicy[S]{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (p *RandomPolicy[S]) Decide(ctx context.Context, current S, domain planning.Domain[S], mem planning.Memory[S]) (planning.Action, error) {
	if domain.IsGoal(current) {
		return "", ErrGoalReached
	}

	actions := domain.Actions(current)
	if len(actions) == 0 {
		return "", nil // Тупик
	}

	// Просто вибираємо випадкову дію
	idx := p.rng.Intn(len(actions))
	return actions[idx], nil
}

func (p *RandomPolicy[S]) Reset() {
	//p.Stack = make([]S, 0)
}
