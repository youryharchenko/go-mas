package main

import (
	"context"
	"encoding/gob"
	"log"
	"time"

	"github.com/youryharchenko/go-mas/mas"
)

type CounterBot struct {
	mas.BaseAgent // Вбудовуємо базу
	Count         int
}

// Реалізуємо Planner прямо для бота, або окремо
func (c *CounterBot) Plan(ctx context.Context, msg mas.Envelope) ([]mas.Action, error) {
	//me := state.(*CounterBot) // Type assertion

	if msg.Payload == "INC" {
		return []mas.Action{
			// Дія 1: Змінити стан
			mas.MutateState(func(a any) {
				a.(*CounterBot).Count++
			}),
			// Дія 2: Логувати
			mas.SayLog("Count increased to %d", c.Count+1),
		}, nil
	}

	return nil, nil
}

func init() {
	gob.Register(&CounterBot{})
}

func main() {
	sys := mas.NewSystem()

	// 1. Спроба відновити попередній стан
	if err := sys.Startup(); err != nil {
		panic(err)
	}

	// 2. Якщо це перший запуск (агентів немає), створюємо початкового
	// У реальній системі тут може бути перевірка if sys.CountAgents() == 0
	if agent, exists := sys.GetAgent("worker-1"); !exists {
		newAgent := &CounterBot{BaseAgent: mas.BaseAgent{IDVal: "worker-1"}, Count: 0}
		sys.Spawn(newAgent)
		log.Println("New CounterBot:", newAgent.ID(), newAgent.Count)
	} else {
		if oldAgent, ok := agent.(*CounterBot); ok {
			log.Println("Old CounterBot:", oldAgent.ID(), oldAgent.Count)
		} else {
			log.Println("Agent is Not CounterBot:", agent.ID())
		}
	}

	// ... робота системи ...
	// Можна відправити повідомлення
	sendCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	log.Println("Before Send")
	err := sys.Send(sendCtx, "main", "worker-1", "INC")
	if err != nil {
		log.Println(err)
	}
	log.Println("After Send")

	// 3. Коректне завершення збереже "worker-1" з Count=1
	// При наступному запуску він прокинеться вже з Count=1
	sys.Shutdown()
}
