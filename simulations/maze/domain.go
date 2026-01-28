package maze

import (
	"github.com/youryharchenko/go-mas/planning"
)

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
	//NewX, NewY  int
	State MazeState
}

// --- 2. DOMAIN (Правила гри) ---
type MazeDomain struct {
	Grid []string // Карта (ReadOnly для агента)
}

func (d *MazeDomain) Actions(s MazeState) []planning.Action {
	var actions []planning.Action
	// Перевіряємо 4 напрямки, чи немає там стіни '#'
	// (Логіка перевірки меж масиву та стін)
	// ...
	return actions
}

func (d *MazeDomain) Result(s MazeState, a planning.Action) MazeState {
	// Повертає нові координати залежно від дії "UP", "DOWN" і т.д.
	next := s
	switch a {
	case "UP":
		next.Y--
	case "DOWN":
		next.Y++
		// ...
	}
	return next
}

func (d *MazeDomain) IsGoal(s MazeState) bool {
	// Перевіряє, чи є в цій клітинці 'E'
	return d.Grid[s.Y][s.X] == 'E'
}

func (d *MazeDomain) StepCost(s1 MazeState, a planning.Action, s2 MazeState) float64 {
	return 1.0 // Кожен крок коштує 1
}

// Heuristic повертає Манхеттенську відстань до виходу 'E'
func (m *MazeAgent) Heuristic(s MazeState) float64 {
	// Знаходимо координати виходу (можна закешувати, але для простоти знайдемо перебором або знаємо структуру)
	// В нашому генераторі вихід завжди в (Width-2, Height-2), але краще знайти чесно.

	// Оптимізація: MazeAgent може зберігати GoalX, GoalY.
	// Для цього прикладу пройдемося по Grid (або використаємо хардкод з генератора).

	goalX, goalY := m.Width-2, m.Height-2

	// Більш чесний пошук (якщо карта не з генератора):
	/*
		found := false
		for y, row := range m.Grid {
			for x, char := range row {
				if char == 'E' {
					goalX, goalY = x, y
					found = true
					break
				}
			}
			if found { break }
		}
	*/

	// Манхеттенська відстань: |x1 - x2| + |y1 - y2|
	dx := float64(s.X - goalX)
	if dx < 0 {
		dx = -dx
	}

	dy := float64(s.Y - goalY)
	if dy < 0 {
		dy = -dy
	}

	return dx + dy
}
