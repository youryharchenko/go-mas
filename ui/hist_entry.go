package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// HistoryEntry - поле вводу, що пам'ятає історію команд
type HistoryEntry struct {
	widget.Entry
	history []string // Список попередніх команд
	pointer int      // Вказівник на поточну позицію в історії
}

func NewHistoryEntry() *HistoryEntry {
	e := &HistoryEntry{}
	e.ExtendBaseWidget(e)
	e.PlaceHolder = "Enter command..."
	e.history = []string{}
	e.pointer = 0
	return e
}

// AddCommand додає команду в історію і скидає вказівник
func (e *HistoryEntry) AddCommand(cmd string) {
	if cmd == "" {
		return
	}
	// Якщо остання команда така сама, не додаємо дублікат
	if len(e.history) > 0 && e.history[len(e.history)-1] == cmd {
		e.pointer = len(e.history)
		return
	}

	e.history = append(e.history, cmd)
	e.pointer = len(e.history) // Вказівник завжди за кінець списку (на пустий рядок)
}

// TypedKey перехоплює стрілки
func (e *HistoryEntry) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyUp:
		if len(e.history) == 0 {
			return
		}
		// Рухаємось назад
		if e.pointer > 0 {
			e.pointer--
			e.setTextAndMoveCursor(e.history[e.pointer])
		}

	case fyne.KeyDown:
		if len(e.history) == 0 {
			return
		}
		// Рухаємось вперед
		if e.pointer < len(e.history)-1 {
			e.pointer++
			e.setTextAndMoveCursor(e.history[e.pointer])
		} else {
			// Якщо дійшли до кінця - очищаємо рядок (ніби новий ввід)
			e.pointer = len(e.history)
			e.setTextAndMoveCursor("")
		}

	default:
		// Для всіх інших клавіш працюємо як звичайний Entry
		e.Entry.TypedKey(key)
	}
}

// setTextAndMoveCursor ставить текст і переміщує курсор в кінець
func (e *HistoryEntry) setTextAndMoveCursor(text string) {
	e.SetText(text)
	// Ставимо курсор в кінець рядка, щоб зручно було дописувати
	// Це трохи хак для Fyne, бо прямий доступ до CursorColumn може змінюватись у версіях
	// Але зазвичай це працює через Focus/Refresh, або просто SetText ставить курсор в кінець.
	// У поточній версії Fyne SetText може лишати курсор на початку, тому:
	e.CursorColumn = len([]rune(text))
	e.Refresh()
}
