package mas

import "context"

type Performative string

const (
	Request Performative = "REQUEST"
	Propose Performative = "PROPOSE"
	Inform  Performative = "INFORM"
)

type Envelope struct {
	From    string
	To      string
	Type    Performative
	Payload any
	// Metadata дозволяє middleware додавати контекст (наприклад, TraceID)
	Metadata map[string]string
}

// Handler - функція, яка обробляє повідомлення
type Handler func(ctx context.Context, env Envelope) error

// Middleware - функція-обгортка (Higher-Order Function)
type Middleware func(next Handler) Handler

// Chain - допоміжна функція для об'єднання middleware
func Chain(middlewares ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}
