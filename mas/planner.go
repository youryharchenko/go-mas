package mas

import "context"

// Planner - це мозок. Він чистий (pure function), не має side-effects.
// Він просто каже, ЩО треба зробити.
type Planner interface {
	Plan(ctx context.Context, state any, msg Envelope) ([]Action, error)
}
