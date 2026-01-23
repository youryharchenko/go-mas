package mas

import (
	"context"
	"fmt"
	"log"
)

// BaseAgent бере на себе всю рутину: канали, системні виклики, цикл.
type BaseAgent struct {
	// Експортовані поля для GOB
	IDVal string

	// Приватні (інфраструктура)
	sys   *System
	inbox <-chan Envelope

	me Agent
}

func (b *BaseAgent) ID() string { return b.IDVal }

func (b *BaseAgent) Bind(sys *System, inbox <-chan Envelope, me Agent) {
	b.sys = sys
	b.inbox = inbox
	b.me = me
}

func (b *BaseAgent) SetSystem(sys *System) {
	b.sys = sys
}

// Run - тепер це стандартний цикл для всіх агентів
func (b *BaseAgent) Run(ctx context.Context) error {
	log.Println("BaseAgent running:", b.ID())
	// Якщо у агента є метод OnWakeUp, кличемо його
	if hook, ok := interface{}(b).(interface{ OnWakeUp() }); ok {
		hook.OnWakeUp()
	}

	for {
		select {

		case msg := <-b.inbox:
			b.processMessage(ctx, msg)
		case <-ctx.Done():
			log.Println("BaseAgent done:", b.ID())
			b.drainInbox(ctx)
			return nil
		}
	}
}

// drainInbox вичитує залишки повідомлень без блокування
func (b *BaseAgent) drainInbox(ctx context.Context) {
	for {
		select {
		case msg := <-b.inbox:
			// Важливо: тут ми все ще обробляємо повідомлення,
			// навіть якщо контекст вже Done.
			// Але передаємо Done-контекст, щоб Планувальник знав про це, якщо треба.
			b.processMessage(ctx, msg)
		default:
			// Канал порожній, тепер можна безпечно помирати
			return
		}
	}
}

// processMessage - винесена логіка (DRY), щоб не дублювати код
func (b *BaseAgent) processMessage(ctx context.Context, msg Envelope) {
	// Викликаємо планувальник
	actions, err := b.me.Plan(ctx, msg)
	if err != nil {
		fmt.Printf("Agent %s planning error: %v\n", b.IDVal, err)
		return
	}

	// Виконуємо дії
	for _, action := range actions {
		// УВАГА: Якщо дія - це Send, вона може впасти, бо система зупиняється.
		// Це нормально.
		if err := action(b.me, b.sys); err != nil {
			// Логуємо помилки, але не панікуємо
			// fmt.Printf("Action failed during shutdown: %v\n", err)
			// Тут можна вставити Middleware (Interceptors) для дій!
			fmt.Printf("Agent %s action failed: %v\n", b.IDVal, err)
		}
	}
}
