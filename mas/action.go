package mas

import (
	"context"
	"fmt"
)

// Action - це команда, яку агент хоче виконати (наприклад: "Надіслати листа", "Змінити стан")
type Action func(agent Agent, sys *System) error

// Send створює дію відправки повідомлення
func Send(to string, payload any) Action {
	return func(a Agent, sys *System) error {
		// Тут ми використовуємо контекст Background, але в ідеалі треба прокидувати з Run
		return sys.Send(context.Background(), a.ID(), to, payload)
	}
}

// SayLog просто пише в консоль (для дебагу)
func SayLog(format string, args ...any) Action {
	return func(a Agent, sys *System) error {
		fmt.Printf("[LOG %s]: "+format+"\n", append([]any{a.ID()}, args...)...)
		return nil
	}
}

// MutateState дозволяє змінити стан (безпечно)
func MutateState(fn func(agent any)) Action {
	return func(a Agent, sys *System) error {
		fn(a)
		return nil
	}
}
