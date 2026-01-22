package mas

import "context"

// Agent - базовий інтерфейс
type Agent interface {
	ID() string
	//Start(ctx context.Context) error
	// Receive - це точка входу для повідомлень
	//Receive(ctx context.Context, env Envelope) error
	//Process(ctx context.Context, env Envelope) error

	// Bind викликається Системою при старті (Spawn або Restore).
	// Тут агент отримує свій канал (Inbox) та посилання на Систему.
	Bind(sys *System, inbox <-chan Envelope, me Agent)

	// Run - запуск основної логіки (горутини)
	Run(ctx context.Context) error

	// Агент зобов'язаний мати "Мозок"
	Plan(ctx context.Context, msg Envelope) ([]Action, error)
}
