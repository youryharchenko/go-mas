package maze

import (
	"context"
	"encoding/gob"
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

	// 3. Ініціалізація агентів
	// Тут ми використовуємо ваші структури.
	// Щоб код скомпілювався, переконайтеся, що GenerateMaze доступна
	/* initialMap := []string{
		"#####",
		"#S..#",
		"#...#",
		"#..E#",
		"#####",
	} */
	// Якщо у вас є функція генерації: initialMap = models.GenerateMaze(15, 15)
	//initialMap := GenerateMaze(15, 15)

	// Створюємо та спавнимо агентів (псевдокод, підставте ваші реальні типи)
	/*
		mazeAgent := &models.MazeAgent{
			BaseAgent: mas.BaseAgent{IDVal: "maze-1"},
			Grid: initialMap,
			// ... інші поля
		}
		mazeSys.Spawn(mazeAgent)

		walkerAgent := &models.PlannerWalker{
			BaseAgent: mas.BaseAgent{IDVal: "walker-1"},
			MazeID: "maze-1",
			// ... інші поля
		}
		mazeSys.Spawn(walkerAgent)
	*/

	wolkerID := "walker-1"
	if _, ok := mazeSys.GetAgent("maze-1"); !ok {
		initialMap := GenerateMaze(15, 15)
		mazeSys.Spawn(&MazeAgent{
			BaseAgent:  mas.BaseAgent{IDVal: "maze-1"},
			Grid:       initialMap,
			WalkerPosX: 1, // Координати S
			WalkerPosY: 1,
			WalkerID:   wolkerID,
		})
	}

	// 2. Створюємо Шукача
	if _, ok := mazeSys.GetAgent(wolkerID); !ok {
		mazeSys.Spawn(&PlannerWalker{
			BaseAgent: mas.BaseAgent{IDVal: wolkerID},
			CurrentX:  1,
			CurrentY:  1,
			MazeID:    "maze-1",
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
				board.UpdateState(realMaze.Grid, realMaze.WalkerPosX, realMaze.WalkerPosY, &realWalker.Visited)
			}

			// ТИМЧАСОВА ЗАГЛУШКА ДЛЯ ТЕСТУ (Видаліть це, коли розкоментуєте код вище)
			// Просто рухаємо точку
			//board.UpdateState(initialMap, 1, 1)
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

	// 4. Кнопки керування
	btnReset := widget.NewButton("New Maze", func() {
		mazeSys.Send(context.Background(), "gui", "maze-1", "NEW")
	})

	// 5. Компонування
	// Board розтягується на весь вільний простір
	content := container.NewBorder(
		nil,      // Top
		btnReset, // Bottom
		nil,      // Left
		nil,      // Right
		board,    // Center
	)

	gob.Register(&MoveRequest{})
	gob.Register(&MoveResult{})
	gob.Register(&MazeAgent{})
	gob.Register(&PlannerWalker{})

	return content
}
