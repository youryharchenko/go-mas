package maze

import (
	"context"
	"fmt"
	"sync"

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
	WalkerID   string
}

func (m *MazeAgent) Plan(ctx context.Context, msg mas.Envelope) ([]mas.Action, error) {
	if msg.Payload == "NEW" {
		// 1. Генеруємо нову карту (непарні розміри, наприклад 15x15)
		// Це дасть ~7x7 ігрових клітинок + стіни
		newMap := GenerateMaze(15, 15)

		return []mas.Action{
			// Змінюємо стан лабіринту
			mas.MutateState(func(a any) {
				maze := a.(*MazeAgent)
				maze.Grid = newMap
				maze.WalkerPosX = 1 // Start завжди в (1,1) за нашим алгоритмом
				maze.WalkerPosY = 1
				maze.Finished = false
			}),
			// Скидаємо мізки Волкеру!
			func(a mas.Agent, sys *mas.System) error {
				return sys.Send(ctx, m.IDVal, m.WalkerID, "RESET")
			},
			mas.Send("console", "Maze: Generated new random level! Resetting walker..."),
			// Показуємо нову карту
			//mas.Send("console", renderMap(newMap, 1, 1)), // func renderMap - це ваш код малювання
		}, nil
	}

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
		/* visualMap := "\n"
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
		} */

		//visualMap := renderMap(m.Grid, newX, newY)

		//actions = append(actions, mas.Send("console", visualMap))

		return actions, nil
	}

	return nil, nil
}

func renderMap(grid []string, wx, wy int) string {
	visualMap := "\n"
	for y, row := range grid {
		line := ""
		for x, char := range row {
			if x == wx && y == wy {
				line += "W"
			} else {
				line += string(char)
			}
		}
		visualMap += line + "\n"
	}
	return visualMap
}

// Helper: Отримання зворотнього напрямку
func Reverse(d Direction) Direction {
	switch d {
	case DirUp:
		return DirDown
	case DirDown:
		return DirUp
	case DirLeft:
		return DirRight
	case DirRight:
		return DirLeft
	}
	return DirUp
}

// Структура розвилки (вузла прийняття рішень)
type Fork struct {
	X, Y         int         // Де ми знаходимось (відносно старту)
	PendingMoves []Direction // Які ходи ми ще НЕ перевірили тут
	EntryMove    Direction   // Яким ходом ми сюди прийшли (щоб знати, як повернутись)
}

// Розумний агент
type PlannerWalker struct {
	mas.BaseAgent
	MazeID string

	// Пам'ять
	Stack []*Fork // Наша "нитка Аріадни"
	//Visited map[string]bool // Де ми вже були ("x,y")
	Visited sync.Map // map[string]bool

	// Стан
	CurrentX, CurrentY int
	LastMove           Direction // Останній зроблений хід
	IsBacktracking     bool      // Чи ми зараз відступаємо назад

	Solved bool
}

func (w *PlannerWalker) Plan(ctx context.Context, msg mas.Envelope) ([]mas.Action, error) {

	// Ініціалізація при першому запуску
	if w.Stack == nil {
		w.Stack = []*Fork{}
		w.Visited.Clear() //make(map[string]bool)
		// Створюємо першу розвилку на старті (0,0)
		startFork := &Fork{
			X: 1, Y: 1,
			PendingMoves: []Direction{DirUp, DirDown, DirLeft, DirRight},
		}
		w.Stack = append(w.Stack, startFork)
		w.Visited.Store("1,1", true)
	}

	if msg.Payload == "RESET" {
		return []mas.Action{
			mas.MutateState(func(a any) {
				walker := a.(*PlannerWalker)
				// Повне очищення пам'яті
				walker.Stack = []*Fork{} // Новий стек
				walker.Visited.Clear()   //= make(map[string]bool)
				walker.Solved = false
				walker.IsBacktracking = false

				// Ініціалізуємо першу точку (як при старті)
				startFork := &Fork{
					X: 1, Y: 1,
					PendingMoves: []Direction{DirUp, DirDown, DirLeft, DirRight},
				}
				walker.Stack = append(walker.Stack, startFork)
				walker.Visited.Store("1,1", true)
			}),
			mas.Send("console", "Planner: Memory wiped. Ready for new maze."),
		}, nil
	}

	// 1. ЛОГІКА ПРИЙНЯТТЯ РІШЕНЬ (TICK)
	if msg.Payload == "TICK" {
		if w.Solved {
			return nil, nil
		}

		if len(w.Stack) == 0 {
			return []mas.Action{mas.Send("console", "Planner: Stack empty. No solution found!")}, nil
		}

		// Беремо поточну розвилку (верхній елемент стека)
		currentFork := w.Stack[len(w.Stack)-1]

		// --- ВАРІАНТ А: Є доступні ходи? ---
		if len(currentFork.PendingMoves) > 0 {
			// 1. Беремо наступний хід
			move := currentFork.PendingMoves[0]
			// 2. Видаляємо його зі списку (ми його зараз перевіримо)
			currentFork.PendingMoves = currentFork.PendingMoves[1:]

			// 3. Перевіряємо, чи не вели б ми в стіну/відвідане (оптимізація в умі)
			// Розрахуємо потенційні координати
			nextX, nextY := currentFork.X, currentFork.Y
			switch move {
			case DirUp:
				nextY--
			case DirDown:
				nextY++
			case DirLeft:
				nextX--
			case DirRight:
				nextX++
			}
			key := fmt.Sprintf("%d,%d", nextX, nextY)

			// Якщо ми там вже були - пропускаємо цей хід (цикл)
			if v, ok := w.Visited.Load(key); ok && v.(bool) {
				// Рекурсивно викликаємо самого себе (або чекаємо наступний тік)
				// Для простоти просто чекаємо наступний тік
				return nil, nil
			}

			// 4. Виконуємо хід
			w.LastMove = move
			w.IsBacktracking = false
			return []mas.Action{
				func(a mas.Agent, sys *mas.System) error {
					return sys.Send(ctx, w.IDVal, w.MazeID, MoveRequest{Dir: move})
				},
			}, nil
		}

		// --- ВАРІАНТ Б: Тупик (ходів немає) ---
		// Треба повертатися до попередньої розвилки
		w.IsBacktracking = true
		w.LastMove = Reverse(currentFork.EntryMove) // Рухаємось назад

		return []mas.Action{
			mas.Send("console", "Planner: Dead end. Backtracking..."),
			func(a mas.Agent, sys *mas.System) error {
				return sys.Send(ctx, w.IDVal, w.MazeID, MoveRequest{Dir: w.LastMove})
			},
		}, nil
	}

	// 2. ОБРОБКА РЕЗУЛЬТАТУ ХОДУ
	if res, ok := msg.Payload.(MoveResult); ok {
		if res.IsFinished {
			return []mas.Action{
				mas.MutateState(func(a any) {
					a.(*PlannerWalker).Solved = true
				}),
				mas.Send("console", "Planner: VICTORY! Path found."),
			}, nil
		}

		if w.IsBacktracking {
			// Ми успішно відступили назад
			if res.Success {
				// Видаляємо тупикову розвилку зі стека
				w.Stack = w.Stack[:len(w.Stack)-1]
				// Оновлюємо координати (повертаємось в старі)
				// (Тут спрощено, в ідеалі брати координати з попереднього Fork)
				curr := w.Stack[len(w.Stack)-1]
				w.CurrentX, w.CurrentY = curr.X, curr.Y
			} else {
				// Це критична помилка - ми не можемо повернутися назад!
				mas.Send("console", "Planner: CRITICAL ERROR. Cannot backtrack!")
			}
			return nil, nil
		}

		// Ми рухалися ВПЕРЕД
		if res.Success {
			// Ура, ми увійшли в нову клітинку
			// 1. Оновлюємо поточні координати
			switch w.LastMove {
			case DirUp:
				w.CurrentY--
			case DirDown:
				w.CurrentY++
			case DirLeft:
				w.CurrentX--
			case DirRight:
				w.CurrentX++
			}

			// 2. Відмічаємо як відвідане
			key := fmt.Sprintf("%d,%d", w.CurrentX, w.CurrentY)
			w.Visited.Store(key, true)

			// 3. СТВОРЮЄМО НОВУ РОЗВИЛКУ (Push to Stack)
			newFork := &Fork{
				X:            w.CurrentX,
				Y:            w.CurrentY,
				PendingMoves: []Direction{DirUp, DirDown, DirLeft, DirRight}, // Всі варіанти
				EntryMove:    w.LastMove,                                     // Запам'ятовуємо, звідки прийшли
			}
			w.Stack = append(w.Stack, newFork)

		} else {
			// Врізалися в стіну
			// Нічого робити не треба. Ми залишилися на місці.
			// Хід вже видалено з PendingMoves поточної розвилки.
			// На наступному TICK ми спробуємо інший варіант.
		}
	}

	return nil, nil
}
