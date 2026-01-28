package maze

import (
	"sync"
)

// MazeMemory - потокобезпечна пам'ять для агента.
// Використовує sync.Map, щоб UI міг читати її під час роботи агента.
type MazeMemory struct {
	visited sync.Map
}

func NewMazeMemory() *MazeMemory {
	return &MazeMemory{}
}

// Remember додає стан у пам'ять.
func (m *MazeMemory) Remember(s MazeState) {
	// Ми зберігаємо саме MazeState, але sync.Map приймає any.
	// State вимагає comparable, тому MazeState (struct {X,Y int}) підходить як ключ.
	m.visited.Store(s, true)
}

// HasVisited перевіряє наявність стану.
func (m *MazeMemory) HasVisited(s MazeState) bool {
	_, exists := m.visited.Load(s)
	return exists
}

// Clear очищає пам'ять.
func (m *MazeMemory) Clear() {
	// sync.Map не має методу Clear, тому створюємо нову.
	// Або проходимо Range і видаляємо (повільно).
	// Простіше замінити об'єкт, але оскільки ми передаємо вказівник UI,
	// краще використати Range+Delete.
	m.visited.Range(func(key, value any) bool {
		m.visited.Delete(key)
		return true
	})
}

// ExportForUI - допоміжний метод для Screen.go
// Дозволяє легко отримати дані для малювання
func (m *MazeMemory) Range(f func(key, value any) bool) {
	m.visited.Range(f)
}
