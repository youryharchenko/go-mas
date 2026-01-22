package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// LogEntry - це віджет, який виглядає як звичайний Entry,
// але дозволяє ТІЛЬКИ копіювати текст, а не писати.
type LogEntry struct {
	widget.Entry
	//scroller *container.Scroll
}

func NewLogEntry() *LogEntry {
	e := &LogEntry{}
	e.ExtendBaseWidget(e) // Магія Fyne для наслідування
	e.MultiLine = true
	e.TextStyle = fyne.TextStyle{Monospace: true} // Шрифт як у терміналі
	e.Wrapping = fyne.TextWrapWord                // Перенос слів
	return e
}

/* func (e *LogEntry) BindScroller(s *container.Scroll) {
	e.scroller = s
} */

// TypedRune викликається, коли ви вводите букви.
// Ми робимо його пустим -> введення ігнорується.
func (e *LogEntry) TypedRune(r rune) {
	// Нічого не робимо (блокуємо ввід тексту)
}

// TypedKey викликається для стрілок, Enter, Backspace тощо.
func (e *LogEntry) TypedKey(key *fyne.KeyEvent) {
	// Дозволяємо тільки навігацію (стрілки, PageUp/Down)
	// щоб можна було скролити клавіатурою
	switch key.Name {
	case fyne.KeyDown, fyne.KeyUp, fyne.KeyLeft, fyne.KeyRight,
		fyne.KeyPageDown, fyne.KeyPageUp, fyne.KeyHome, fyne.KeyEnd:
		e.Entry.TypedKey(key) // Викликаємо оригінальний метод
	}
	// Всі інші клавіші (Backspace, Delete, Enter) ігноруємо
}

// TypedShortcut обробляє Ctrl+C, Ctrl+V
func (e *LogEntry) TypedShortcut(shortcut fyne.Shortcut) {
	// Якщо це копіювання - дозволяємо
	if _, ok := shortcut.(*fyne.ShortcutCopy); ok {
		e.Entry.TypedShortcut(shortcut)
	}
	// Cut (Вирізати) і Paste (Вставити) - ігноруємо
}
