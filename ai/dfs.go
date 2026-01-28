package ai

import (
	"context"
	"errors"
	"fmt"

	"github.com/youryharchenko/go-mas/planning"
)

var ErrNoSolution = errors.New("no solution found (stack empty)")
var ErrGoalReached = errors.New("goal reached")

// DFSPolicy реалізує алгоритм Depth-First Search в режимі онлайн.
// Він пам'ятає шлях (Stack) і вміє повертатися назад (Backtracking).
type DFSPolicy[S planning.State] struct {
	Stack []S // Історія шляху (для бектрекінгу)
}

func NewDFS[S planning.State]() *DFSPolicy[S] {
	return &DFSPolicy[S]{
		Stack: make([]S, 0),
	}
}

// Decide приймає поточний стан і вирішує, куди йти.
func (p *DFSPolicy[S]) Decide(ctx context.Context, current S, domain planning.Domain[S], mem planning.Memory[S]) (planning.Action, error) {

	// 0. Ініціалізація (якщо це перший хід)
	if len(p.Stack) == 0 {
		p.Stack = append(p.Stack, current)
		mem.Remember(current)
	}

	// 1. Перевірка перемоги
	if domain.IsGoal(current) {
		return "", ErrGoalReached
	}

	// 2. Отримуємо можливі дії
	actions := domain.Actions(current)

	// 3. Шукаємо невідвідані шляхи
	for _, action := range actions {
		nextState := domain.Result(current, action)
		if !mem.HasVisited(nextState) {
			// Знайшли новий шлях!
			// 1. Запам'ятовуємо його в стек (насправді ми кладемо туди майбутній стан)
			// Але правильніше класти поточний, щоб знати куди повертатись.
			// У цьому варіанті ми пушимо current в стек, коли йдемо глибше.
			p.Stack = append(p.Stack, current)
			return action, nil
		}
	}

	// 4. Якщо нових шляхів немає -> Backtracking (Відступ)
	// Нам треба повернутися до попереднього стану в стеку.

	// Якщо стек порожній (крім поточного елемента), то ми перевірили все.
	if len(p.Stack) == 0 {
		return "", ErrNoSolution
	}

	// Беремо попередній вузол (куди треба повернутися)
	// Поточний вузол ми вже дослідили повністю, тому ми його покидаємо.
	// Але оскільки ми не зберігаємо його в стеку при вході (в цій простій реалізації),
	// логіка трохи інша: Stack зберігає ШЛЯХ від старту до current.

	// Витягуємо батька
	previous := p.Stack[len(p.Stack)-1]

	// Якщо ми вже в ньому (старт), і йти нікуди
	if previous.Equals(current) {
		// (Це крайній випадок для старту)
		if len(p.Stack) > 1 {
			previous = p.Stack[len(p.Stack)-2]
		} else {
			return "", ErrNoSolution
		}
	}

	// Зменшуємо стек (ми покидаємо current і повертаємось в previous)
	p.Stack = p.Stack[:len(p.Stack)-1]

	// 5. Знаходимо дію, яка веде назад (від Current до Previous)
	// Оскільки граф ненаправлений, ми шукаємо сусіда Current, який дорівнює Previous.
	backActions := domain.Actions(current)
	for _, act := range backActions {
		if domain.Result(current, act).Equals(previous) {
			return act, nil
		}
	}

	return "", fmt.Errorf("critical: cannot find path back from %v to %v", current, previous)
}

func (p *DFSPolicy[S]) Reset() {
	p.Stack = make([]S, 0)
}
