package maze

import (
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

	"github.com/youryharchenko/go-mas/mas"
	// Імпортуйте ваші моделі агентів
	// "github.com/youryharchenko/go-mas/models"
)

// NewScreen створює вміст вкладки та запускає підсистему
func NewScreen(parentSys *mas.System, logData binding.String) fyne.CanvasObject {

	// 1. Створюємо підсистему
	mazeSys := parentSys.CreateSubsystem()

	// 2. Створюємо графічний віджет
	board := NewMazeBoard()

	wolkerID := "walker-1"
	if _, ok := mazeSys.GetAgent("maze-1"); !ok {
		initialMap := GenerateMaze(15, 15)
		mazeSys.Spawn(&MazeAgent{
			BaseAgent: mas.BaseAgent{IDVal: "maze-1"},
			Grid:      initialMap,
			Width:     15,
			Height:    15,
			WalkerPos: MazeState{X: 1, Y: 1},
			//CurrentX:  1, // Координати S
			//CurrentY:  1,
			WalkerID: wolkerID,
		})
	}

	// 2. Створюємо Шукача
	if _, ok := mazeSys.GetAgent(wolkerID); !ok {
		mazeSys.Spawn(&PlannerWalker{
			BaseAgent:    mas.BaseAgent{IDVal: wolkerID},
			CurrentState: MazeState{X: 1, Y: 1},
			//CurrentX:  1,
			//CurrentY:  1,
			MazeID: "maze-1",
		})
	}

	// --- UI LOOP (Оновлення графіки) ---
	// Запускаємо горутину, яка синхронізує стан агентів з UI
	// Це найпростіший спосіб без зміни коду агентів
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond) // 10 FPS
		for range ticker.C {
			// Отримуємо доступ до агентів (через GetAgent, який ми додали вчора)
			// (Тут треба привести типи до конкретних структур)

			// ПРИКЛАД:
			aMaze, ok1 := mazeSys.GetAgent("maze-1")
			aWalker, ok2 := mazeSys.GetAgent("walker-1")

			if ok1 && ok2 {
				realMaze := aMaze.(*MazeAgent)
				realWalker := aWalker.(*PlannerWalker)

				// Оновлюємо віджет
				// Важливо робити це в потоці UI, але Refresh() thread-safe
				//board.UpdateState(realMaze.Grid, realWalker.CurrentX, realWalker.CurrentY)
				//log.Println("realWalker:", realWalker.CurrentX, realWalker.CurrentY)
				//log.Println("realMaze:", realMaze.WalkerPosX, realMaze.WalkerPosY)

				// Підготовка даних пам'яті для UI
				uiVisited := make(map[string]bool)

				// Читаємо з sync.Map через наш helper Range
				if realWalker.Memory != nil {
					realWalker.Memory.Range(func(key, value any) bool {
						// Ключ у мапі тепер MazeState {X,Y}
						if s, ok := key.(MazeState); ok {
							k := fmt.Sprintf("%d,%d", s.X, s.Y)
							uiVisited[k] = true
						}
						return true
					})
				}

				board.UpdateState(realMaze.Grid, realWalker.CurrentState.X, realWalker.CurrentState.Y, uiVisited)
			}

		}
	}()

	// Глобальний таймер світу
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond) // 2 ходи на секунду
		for {
			select {
			case <-ticker.C:
				// Пінаємо волкера, щоб він думав
				mazeSys.Send(context.Background(), "admin", wolkerID, "TICK")
			}
		}
	}()

	// Кнопка "Новий лабіринт"
	btnNew := widget.NewButton("New Maze", func() {
		mazeSys.Send(context.Background(), "gui", "maze-1", "NEW")
	})

	// Кнопка "Скинути пам'ять"
	btnResetWalker := widget.NewButton("Reset Memory", func() {
		mazeSys.Send(context.Background(), "gui", "walker-1", "RESET")
	})

	// --- НОВЕ: Вибір Стратегії ---
	policySelect := widget.NewSelect([]string{"DFS", "AStar", "BFS", "Random"}, func(selected string) {
		// Відправляємо команду агенту змінити мозок
		// Формат payload: "POLICY:DFS"
		cmd := "POLICY:" + selected
		mazeSys.Send(context.Background(), "gui", "walker-1", cmd)
	})
	policySelect.SetSelected("DFS") // Значення за замовчуванням

	// Панель інструментів (додаємо селект)
	toolbar := container.NewHBox(
		btnNew,
		btnResetWalker,
		widget.NewLabel("Strategy:"), // Підпис
		policySelect,
	)

	// 5. Компонування
	// Board розтягується на весь вільний простір
	content := container.NewBorder(
		toolbar, // Top
		nil,     // Bottom
		nil,     // Left
		nil,     // Right
		board,   // Center
	)

	gob.Register(&MoveRequest{})
	gob.Register(&MoveResult{})
	gob.Register(&MazeAgent{})
	gob.Register(&PlannerWalker{})

	return content
}
