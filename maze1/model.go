package main

import (
	"context"
	"encoding/gob"
	"math/rand"

	"github.com/youryharchenko/go-mas/mas"
)

// Напрямки
type Direction int

const (
	DirUp Direction = iota
	DirDown
	DirLeft
	DirRight
)

// Запит на рух (від Волкера до Лабіринту)
type MoveRequest struct {
	Dir Direction
}

// Відповідь (від Лабіринту до Волкера)
type MoveResult struct {
	Success     bool   // Чи вдалося переміститись
	Message     string // Текстовий опис ("Bonk! Wall.", "Step taken.")
	IsFinished  bool   // Чи це вихід?
	CurrentView string // (Опціонально) Що бачить агент навколо
}

type MazeAgent struct {
	mas.BaseAgent
	Grid       []string // Карта рядками
	WalkerPosX int      // Де зараз гравець (X)
	WalkerPosY int      // Де зараз гравець (Y)
	Finished   bool
}

func (m *MazeAgent) Plan(ctx context.Context, msg mas.Envelope) ([]mas.Action, error) {

	// Обробка запиту на рух
	if req, ok := msg.Payload.(MoveRequest); ok {
		if m.Finished {
			return nil, nil // Гра закінчена
		}

		newX, newY := m.WalkerPosX, m.WalkerPosY

		// 1. Рахуємо нову координату
		switch req.Dir {
		case DirUp:
			newY--
		case DirDown:
			newY++
		case DirLeft:
			newX--
		case DirRight:
			newX++
		}

		// 2. Перевіряємо фізику (стіни)
		// Увага: спрощена перевірка меж масиву (в реальності треба check bounds)
		cell := m.Grid[newY][newX]

		if cell == '#' {
			// Врізався!
			return []mas.Action{
				mas.Send(msg.From, MoveResult{Success: false, Message: "BONK! Wall hit."}),
			}, nil
		}

		// 3. Якщо прохід вільний - оновлюємо стан
		actions := []mas.Action{
			mas.MutateState(func(a any) {
				maze := a.(*MazeAgent)
				maze.WalkerPosX = newX
				maze.WalkerPosY = newY
				if cell == 'E' {
					maze.Finished = true
				}
			}),
		}

		// 4. Формуємо відповідь
		resultMsg := "Moved."
		isWin := false
		if cell == 'E' {
			resultMsg = "VICTORY! Found Exit!"
			isWin = true
		}

		actions = append(actions, mas.Send(msg.From, MoveResult{
			Success:    true,
			Message:    resultMsg,
			IsFinished: isWin,
		}))

		// 5. (Бонус) Малюємо карту в консоль, щоб ми бачили прогрес
		// Ми сформуємо рядок, де 'W' - це позиція воркера
		visualMap := "\n"
		for y, row := range m.Grid {
			line := ""
			for x, char := range row {
				if x == newX && y == newY {
					line += "W" // Показуємо Волкера
				} else {
					line += string(char)
				}
			}
			visualMap += line + "\n"
		}
		actions = append(actions, mas.Send("console", visualMap))

		return actions, nil
	}

	return nil, nil
}

type WalkerBot struct {
	mas.BaseAgent
	MazeID string
}

func (w *WalkerBot) Plan(ctx context.Context, msg mas.Envelope) ([]mas.Action, error) {

	// 1. Таймер тікає - треба робити хід
	if msg.Payload == "TICK" {
		// Випадковий напрямок (0..3)
		randomDir := Direction(rand.Intn(4))

		return []mas.Action{
			// Шлемо запит Лабіринту
			func(a mas.Agent, sys *mas.System) error {
				return sys.Send(ctx, w.IDVal, w.MazeID, MoveRequest{Dir: randomDir})
			},
		}, nil
	}

	// 2. Реакція на відповідь від стіни/проходу
	if res, ok := msg.Payload.(MoveResult); ok {
		if res.IsFinished {
			return []mas.Action{mas.Send("console", "Walker: I WON! STOPPING.")}, nil
		}

		// Можна логувати удари
		// if !res.Success { mas.Send("console", "Walker: Ouch!") }

		// Важливо: Тут ми могли б запам'ятовувати карту, якби мали пам'ять
	}

	return nil, nil
}

func init() {
	gob.Register(&MoveRequest{})
	gob.Register(&MoveResult{})
	gob.Register(&MazeAgent{})
	gob.Register(&WalkerBot{})
}
