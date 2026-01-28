package maze

import (
	"context"
	"fmt"

	"github.com/youryharchenko/go-mas/mas"
	"github.com/youryharchenko/go-mas/planning"
)

// --- 1. STATE (Стан) ---

// MazeState представляє положення агента в лабіринті.
type MazeState struct {
	X, Y int
}

// String реалізує інтерфейс fmt.Stringer (вимога planning.State).
func (s MazeState) String() string {
	return fmt.Sprintf("%d,%d", s.X, s.Y)
}

func (s MazeState) Equals(other planning.State) bool {
	if o, ok := other.(MazeState); ok {
		return s.X == o.X && s.Y == o.Y
	}
	return false
}

// --- 2. AGENT + DOMAIN (Агент і Фізика) ---

// MazeAgent виступає як Середовище (Environment) і як Домен планування (Domain).
type MazeAgent struct {
	mas.BaseAgent

	// Стан світу
	Grid     []string
	Width    int
	Height   int
	WalkerID string // ID агента, якого ми "спінимо" при ресеті
	//
	WalkerPos MazeState
}

// --- Реалізація інтерфейсу planning.Domain[MazeState] ---

// Actions повертає список доступних дій для даної клітинки.
func (m *MazeAgent) Actions(s MazeState) []planning.Action {
	var actions []planning.Action

	// Напрямки (dx, dy, назва)
	dirs := []struct {
		dx, dy int
		name   string
	}{
		{0, -1, "UP"},
		{0, 1, "DOWN"},
		{-1, 0, "LEFT"},
		{1, 0, "RIGHT"},
	}

	for _, d := range dirs {
		nx, ny := s.X+d.dx, s.Y+d.dy

		// 1. Перевірка меж світу
		if nx >= 0 && nx < m.Width && ny >= 0 && ny < m.Height {
			// 2. Перевірка стін (фізика)
			cell := m.Grid[ny][nx]
			if cell != '#' {
				actions = append(actions, planning.Action(d.name))
			}
		}
	}

	return actions
}

// Result повертає стан, у який потрапить агент, виконавши дію.
func (m *MazeAgent) Result(s MazeState, a planning.Action) MazeState {
	next := s
	switch a {
	case "UP":
		next.Y--
	case "DOWN":
		next.Y++
	case "LEFT":
		next.X--
	case "RIGHT":
		next.X++
	}
	// Тут можна було б додати перевірку на стіни знову,
	// але за контрактом Planner викликає Result тільки для дій з Actions().
	return next
}

// IsGoal перевіряє, чи є цей стан перемогою.
func (m *MazeAgent) IsGoal(s MazeState) bool {
	// Перевірка меж (на всяк випадок)
	if s.Y < 0 || s.Y >= len(m.Grid) || s.X < 0 || s.X >= len(m.Grid[0]) {
		return false
	}
	return m.Grid[s.Y][s.X] == 'E'
}

// StepCost повертає вартість кроку (для A* буде 1.0, для болота може бути 5.0).
func (m *MazeAgent) StepCost(from MazeState, action planning.Action, to MazeState) float64 {
	return 1.0
}

// --- Реалізація інтерфейсу mas.Agent (Логіка Агента) ---

func (m *MazeAgent) Plan(ctx context.Context, msg mas.Envelope) ([]mas.Action, error) {

	// 1. GENERATION (Створення нового рівня)
	if msg.Payload == "NEW" {
		// Генеруємо лабіринт
		m.Grid = GenerateMaze(m.Width, m.Height)
		// Скидаємо позицію гравця на старт (зазвичай 1,1)
		m.WalkerPos.X, m.WalkerPos.Y = 1, 1

		return []mas.Action{
			mas.Send("console", "Maze: New map generated."),
			// Наказуємо воркеру забути минуле
			func(a mas.Agent, sys *mas.System) error {
				return sys.Send(ctx, m.IDVal, m.WalkerID, "RESET")
			},
		}, nil
	}

	// 2. MOVEMENT (Обробка реального руху)
	if req, ok := msg.Payload.(MoveRequest); ok {
		// Якщо карти немає - нічого не робимо
		if len(m.Grid) == 0 {
			return nil, nil
		}

		newX, newY := m.WalkerPos.X, m.WalkerPos.Y
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

		// Перевірка колізій (чи не врізався у стіну насправді)
		hitWall := false
		// Перевірка меж
		if newY < 0 || newY >= len(m.Grid) || newX < 0 || newX >= len(m.Grid[0]) {
			hitWall = true
		} else if m.Grid[newY][newX] == '#' {
			hitWall = true
		}

		if hitWall {
			// Врізався! Повертаємо помилку і СТАРІ координати
			return []mas.Action{
				mas.Send(msg.From, MoveResult{
					Success: false,
					Message: "BONK!",
					State:   m.WalkerPos,
				}),
			}, nil
		}

		// Успішний рух
		isWin := m.Grid[newY][newX] == 'E'

		return []mas.Action{
			// Оновлюємо стан середовища
			mas.MutateState(func(a any) {
				maze := a.(*MazeAgent)
				maze.WalkerPos = MazeState{X: newX, Y: newY}
				//maze.CurrentX = newX
				//maze.CurrentY = newY
			}),
			// Відповідаємо агенту з НОВИМИ координатами
			mas.Send(msg.From, MoveResult{
				Success:    true,
				IsFinished: isWin,
				State:      MazeState{X: newX, Y: newY},
				//NewX:       newX,
				//NewY:       newY,
			}),
		}, nil
	}

	return nil, nil
}
