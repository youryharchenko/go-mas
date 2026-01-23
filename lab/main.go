package main

import (
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/youryharchenko/go-mas/mas"
	"github.com/youryharchenko/go-mas/simulations/maze"
	"github.com/youryharchenko/go-mas/ui"
)

func main() {
	a := app.New()
	w := a.NewWindow("MAS Laboratory")
	w.Resize(fyne.NewSize(1024, 768))

	// 1. Створюємо Систему
	sys := mas.NewSystem()

	// 2. Створюємо Лог (він спільний для всіх)
	logData := binding.NewString()
	logData.Set("System started...")

	outputEntry := ui.NewLogEntry()
	outputEntry.Bind(logData)
	outputEntry.TextStyle = fyne.TextStyle{Monospace: true}

	inputEntry := ui.NewHistoryEntry()
	inputEntry.PlaceHolder = "Enter command: TARGET PAYLOAD (use Up/Down for history)..."

	inputEntry.OnSubmitted = func(text string) {
		if text == "" {
			return
		}

		inputEntry.AddCommand(text)

		inputEntry.SetText("")

	}

	// Для краси загорнемо в скрол
	scroll := container.NewVScroll(outputEntry)
	//outputEntry.BindScroller(scroll)
	logData.AddListener(binding.NewDataListener(func() {
		// Ця функція спрацює автоматично при кожній зміні тексту
		outputEntry.CursorRow = len(outputEntry.Text) - 1
		outputEntry.Refresh()
	}))

	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(250, 0))
	rightStack := container.NewStack(spacer, scroll)

	// 3. Створюємо екран Лабіринту (ми його напишемо сьогодні/завтра)
	// Він повертає container, який містить графіку лабіринту

	mazeTabContent := maze.NewScreen(sys, logData)

	// 4. Створюємо вкладки
	tabs := container.NewAppTabs(
		container.NewTabItem("Maze Runner", mazeTabContent),
		container.NewTabItem("Future Sim", widget.NewLabel("Coming soon...")),
	)

	// Встановлюємо, де знаходяться вкладки (зверху, знизу...)
	tabs.SetTabLocation(container.TabLocationTop)

	// 3. КОМПОНУВАННЯ (Layout)
	// Використовуємо Border Layout:
	// - Знизу (Bottom): поле вводу
	// - Центр (Center): скрол з логами (займає весь доступний простір)
	content := container.NewBorder(
		nil,        // Top
		inputEntry, // Bottom
		nil,        // Left
		rightStack, // Right
		tabs,       // Center
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(1024, 800))

	err := sys.Startup()
	if err != nil {
		log.Println(err)
	}

	guiAgent := ui.NewLogWindowAgent("console", logData)
	sys.Spawn(guiAgent)

	adminAgent := ui.NewLogWindowAgent("admin", logData)
	sys.Spawn(adminAgent)

	w.ShowAndRun()

	sys.Kill("console") // Видаляємо агента, щоб не зберігати його у файл
	sys.Kill("admin")
	if err := sys.Shutdown(); err != nil {
		log.Println(err)
	}
}
