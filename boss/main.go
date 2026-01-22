package main

import (
	"context"
	"time"

	"github.com/youryharchenko/go-mas/mas"
)

func main() {
	sys := mas.NewSystem(mas.WithPersistence("world.gob"))
	sys.Startup() // Відновили старих (Воркера з Count=5)

	if _, exists := sys.GetAgent("worker-1"); !exists {
		worker := &WorkerBot{BaseAgent: mas.BaseAgent{IDVal: "worker-1"}, Count: 0}
		sys.Spawn(worker)
	}

	if _, ok := sys.GetAgent("boss-1"); !ok {
		boss := &ManagerBot{
			BaseAgent:     mas.BaseAgent{IDVal: "boss-1"}, // ID треба задавати явно
			TargetAgentID: "worker-1",
		}
		// Важливо: BaseAgent має поле IDVal, але ми ще не реалізували конструктор,
		// тому задаємо вручну або через NewManagerBot("boss-1")
		sys.Spawn(boss)
	}

	// 2. Запускаємо "Ігровий Цикл"
	// Замість ручного send("INC"), ми просто пінаємо менеджера, щоб він почав працювати
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// Зовнішній світ каже босу: "Час працювати"
				sys.Send(context.Background(), "main", "boss-1", "TICK")
			case <-sys.Context().Done(): // Якщо система зупиняється
				return
			}
		}
	}()

	// Працюємо 10 секунд і виходимо
	time.Sleep(10 * time.Second)
	sys.Shutdown()
}
