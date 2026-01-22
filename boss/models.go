package main

import (
	"context"
	"encoding/gob"
	"fmt"

	"github.com/youryharchenko/go-mas/mas"
)

type WorkOrder struct {
	TaskID string
	Amount int
}

type ManagerBot struct {
	mas.BaseAgent
	TargetAgentID string // Кого ми будемо пінати
	TasksSent     int
}

func (m *ManagerBot) Plan(ctx context.Context, msg mas.Envelope) ([]mas.Action, error) {
	// 1. Реакція на вхідні (наприклад, звіт від воркера)
	if msg.Payload == "DONE" {
		return []mas.Action{
			mas.SayLog("Worker finished a task. Good job."),
		}, nil
	}

	// 2. Реакція на "Тік" таймера (див. нижче про Loop)
	if msg.Payload == "TICK" {
		m.TasksSent++
		task := WorkOrder{TaskID: fmt.Sprintf("job-%d", m.TasksSent), Amount: 1}

		return []mas.Action{
			mas.SayLog("Assigning task %s to %s", task.TaskID, m.TargetAgentID),
			// Менеджер відправляє повідомлення Воркеру
			func(a mas.Agent, sys *mas.System) error {
				return sys.Send(ctx, m.IDVal, m.TargetAgentID, task)
			},
		}, nil
	}

	return nil, nil
}

type WorkerBot struct {
	mas.BaseAgent // Вбудовуємо базу
	Count         int
}

func (w *WorkerBot) Plan(ctx context.Context, msg mas.Envelope) ([]mas.Action, error) {
	// Перевіряємо тип повідомлення (Type Switch)
	switch payload := msg.Payload.(type) {

	case WorkOrder:
		return []mas.Action{
			mas.MutateState(func(a any) {
				a.(*WorkerBot).Count += payload.Amount
			}),
			mas.SayLog("Received order %s. Count is now %d", payload.TaskID, w.Count+payload.Amount),
			// Відповідаємо Менеджеру
			mas.Send(msg.From, "DONE"),
		}, nil

	default:
		// Ігноруємо невідоме
		return nil, nil
	}
}

// init обов'язково
func init() {
	gob.Register(&WorkOrder{})
	gob.Register(&ManagerBot{})
	gob.Register(&WorkerBot{})
}
