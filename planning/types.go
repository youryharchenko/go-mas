package planning

import (
	"context"
	"fmt"
)

// Action - це атомарна дія, яку агент може виконати.
// Ми використовуємо string для простоти серіалізації та логування,
// але це може бути будь-який тип, що реалізує Stringer.
type Action string

// State - це "зліпок" реальності в конкретний момент часу.
// Constraint:
// 1. comparable - щоб ми могли використовувати стан як ключ у map (для Memory/Visited).
// 2. fmt.Stringer - щоб ми могли логувати стан у консоль.
type State interface {
	fmt.Stringer
	Equals(State) bool
}

// Domain (або Environment Physics) - описує правила світу.
// Це чиста логіка: вона не зберігає стан, а лише відповідає на запитання про нього.
// Аналог в Haskell: набір чистих функцій, що приймають State і повертають State.
type Domain[S State] interface {
	// Actions повертає список доступних дій для даного стану.
	// Наприклад: для MazeState(0,0) -> ["RIGHT", "DOWN"]
	Actions(s S) []Action

	// Result (Transition Model) повертає новий стан після виконання дії.
	// Це детермінована зміна світу: S x A -> S'
	Result(s S, a Action) S

	// IsGoal перевіряє, чи досягнуто мети (Terminal State).
	IsGoal(s S) bool

	// StepCost повертає вартість переходу (потрібно для A*, Uniform Cost Search).
	// Зазвичай 1.0.
	StepCost(from S, action Action, to S) float64

	Heuristic(s S) float64
}

// Memory - абстракція пам'яті агента.
// Дозволяє агенту пам'ятати, де він був, щоб не ходити колами.
type Memory[S State] interface {
	// Remember додає стан у пам'ять.
	Remember(s S)

	// HasVisited перевіряє, чи був агент у цьому стані.
	HasVisited(s S) bool

	// Clear очищує пам'ять (для Reset).
	Clear()

	Range(f func(key, value any) bool)
}

// Policy (Стратегія/Brain) - вирішує, що робити.
// Це "чорна скринька", яка приймає поточну ситуацію і видає дію.
// Реалізації: RandomPolicy, DFSPolicy, AStarPolicy, HumanPolicy.
type Policy[S State] interface {
	// Decide повертає наступну дію.
	// Якщо дій немає (глухий кут) або рішення знайдено, повертає порожню Action ("") або error.
	Decide(ctx context.Context, current S, domain Domain[S], memory Memory[S]) (Action, error)

	Reset()
}

// Planner - (опціонально) об'єкт, який може побудувати повний план (послідовність дій)
// від start до goal, не виконуючи їх.
type Planner[S State] interface {
	MakePlan(ctx context.Context, start S, domain Domain[S]) ([]Action, error)
}
