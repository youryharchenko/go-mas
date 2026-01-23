package ui

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2/data/binding"
	"github.com/youryharchenko/go-mas/mas"
)

// LogWindowAgent - це агент, який керує текстовим полем на екрані
type LogWindowAgent struct {
	mas.BaseAgent

	// Приватне поле, GOB його проігнорує (і це добре, бо GUI не можна зберігати)
	// Це "ниточка" до інтерфейсу
	output binding.String
}

// NewLogWindowAgent - конструктор
func NewLogWindowAgent(id string, data binding.String) *LogWindowAgent {
	return &LogWindowAgent{
		BaseAgent: mas.BaseAgent{IDVal: id},
		output:    data,
	}
}

// Plan - реагує на повідомлення
func (g *LogWindowAgent) Plan(ctx context.Context, msg mas.Envelope) ([]mas.Action, error) {
	// Всі повідомлення, що приходять, ми просто відображаємо
	text := fmt.Sprintf("[%s -> %s]: %+v", msg.From, g.IDVal, msg.Payload)

	// Оновлюємо GUI через binding (це Thread-Safe у Fyne)
	current, _ := g.output.Get()
	newText := current + "\n" + text
	g.output.Set(newText)

	return nil, nil
}

// GUI агенти не треба відновлювати з диска, вони створюються при запуску UI
// Тому OnWakeUp тут не потрібен, або порожній.
