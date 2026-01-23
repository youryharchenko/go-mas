package main

import (
	"context"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"github.com/youryharchenko/go-mas/mas"
)

func main() {
	// Запускаємо MAS
	sys := mas.NewSystem()

	// 1. Створюємо Fyne App
	myApp := app.New()
	w := myApp.NewWindow("Maze 3")

	// 2. Створюємо Data Binding (спільна пам'ять для UI і Агента)
	logData := binding.NewString()
	logData.Set("System started...")

	// 3. Створюємо UI елемент, прив'язаний до даних
	// Label автоматично перемалюється, коли зміниться logData
	//label := widget.NewLabelWithData(logData)
	outputEntry := NewLogEntry()
	outputEntry.Bind(logData)
	outputEntry.TextStyle = fyne.TextStyle{Monospace: true}

	inputEntry := NewHistoryEntry()
	inputEntry.PlaceHolder = "Enter command: TARGET PAYLOAD (use Up/Down for history)..."

	inputEntry.OnSubmitted = func(text string) {
		if text == "" {
			return
		}

	}

	// Для краси загорнемо в скрол
	scroll := container.NewVScroll(outputEntry)
	//outputEntry.BindScroller(scroll)
	logData.AddListener(binding.NewDataListener(func() {
		// Ця функція спрацює автоматично при кожній зміні тексту
		outputEntry.CursorRow = len(outputEntry.Text) - 1
		outputEntry.Refresh()
	}))

	// 3. КОМПОНУВАННЯ (Layout)
	// Використовуємо Border Layout:
	// - Знизу (Bottom): поле вводу
	// - Центр (Center): скрол з логами (займає весь доступний простір)
	content := container.NewBorder(
		nil,        // Top
		inputEntry, // Bottom
		nil,        // Left
		nil,        // Right
		scroll,     // Center
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(1024, 800))

	err := sys.Startup() // Відновили старих (Воркера з Count=5)
	if err != nil {
		log.Println(err)
	}

	// --- МАГІЯ ТУТ ---
	// Створюємо агента і даємо йому в руки binding
	guiAgent := NewLogWindowAgent("console", logData)
	sys.Spawn(guiAgent)

	adminAgent := NewLogWindowAgent("admin", logData)
	sys.Spawn(adminAgent)

	wolkerID := "walker-1"

	if _, ok := sys.GetAgent("maze-1"); !ok {
		initialMap := GenerateMaze(15, 15)
		sys.Spawn(&MazeAgent{
			BaseAgent:  mas.BaseAgent{IDVal: "maze-1"},
			Grid:       initialMap,
			WalkerPosX: 1, // Координати S
			WalkerPosY: 1,
			WalkerID:   wolkerID,
		})
	}

	// 2. Створюємо Шукача
	if _, ok := sys.GetAgent(wolkerID); !ok {
		sys.Spawn(&PlannerWalker{
			BaseAgent: mas.BaseAgent{IDVal: wolkerID},
			MazeID:    "maze-1",
		})
	}

	// Глобальний таймер світу
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond) // 2 ходи на секунду
		for {
			select {
			case <-ticker.C:
				// Пінаємо волкера, щоб він думав
				sys.Send(context.Background(), "admin", wolkerID, "TICK")
			}
		}
	}()

	// Щоб це працювало, Воркер має вміти слати логи не в fmt.Println, а агенту
	// Це вимагає маленької зміни в CounterBot (див. нижче)

	w.ShowAndRun()

	sys.Kill("console") // Видаляємо агента, щоб не зберігати його у файл
	sys.Kill("admin")
	if err := sys.Shutdown(); err != nil {
		log.Println(err)
	}
}

// splitCommand ділить рядок на два: "WORD1 REST OF STRING"
func splitCommand(input string) []string {
	return strings.SplitN(input, " ", 2)
}

// appendLog безпечно додає рядок до binding
func appendLog(data binding.String, text string) {
	current, _ := data.Get()
	// Якщо лог дуже довгий, можна обрізати початок, щоб не їсти пам'ять
	if len(current) > 5000 {
		current = current[len(current)-4000:] // залишаємо останні 4000 символів
	}
	data.Set(current + "\n" + text)
}
