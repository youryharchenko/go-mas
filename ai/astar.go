package ai

import (
	"container/heap"
	"context"
	"fmt"

	"github.com/youryharchenko/go-mas/planning"
)

// AStarPolicy реалізує пошук A*.
// F = G + H (Вартість шляху + Евристика)
type AStarPolicy[S interface {
	planning.State
	comparable
}] struct {
	OpenSet    *PriorityQueue[S]
	CameFrom   map[S]S
	GScore     map[S]float64
	PathBuffer []planning.Action

	initialized bool
}

func NewAStar[S interface {
	planning.State
	comparable
}]() *AStarPolicy[S] {
	pq := make(PriorityQueue[S], 0)
	return &AStarPolicy[S]{
		OpenSet:  &pq,
		CameFrom: make(map[S]S),
		GScore:   make(map[S]float64),
	}
}

func (p *AStarPolicy[S]) Decide(ctx context.Context, current S, domain planning.Domain[S], mem planning.Memory[S]) (planning.Action, error) {

	// 0. Ініціалізація
	if !p.initialized {
		heap.Init(p.OpenSet)
		p.GScore[current] = 0
		fScore := domain.Heuristic(current)

		heap.Push(p.OpenSet, &Item[S]{
			Value:    current,
			Priority: fScore, // Чим менше F, тим вищий пріоритет (min-heap)
		})
		p.CameFrom[current] = current // Корінь
		mem.Remember(current)
		p.initialized = true
	}

	// 1. Виконання буфера (рух до цілі)
	if len(p.PathBuffer) > 0 {
		action := p.PathBuffer[0]
		p.PathBuffer = p.PathBuffer[1:]
		return action, nil
	}

	if domain.IsGoal(current) {
		return "", ErrGoalReached
	}

	// 2. Алгоритм A* (розширення вузлів)
	if p.OpenSet.Len() == 0 {
		return "", ErrNoSolution
	}

	// Беремо найкращий вузол з черги
	// (Але не видаляємо остаточно, поки не перейдемо до нього?
	// Ні, в A* ми беремо, розкриваємо і кладемо сусідів).

	// Проблема фізичного агента: Ми "думаємо" про вузол currentItem.Value,
	// але фізично ми знаходимося в 'current'.
	// Нам треба фізично дійти до currentItem.Value, щоб "розкрити" його сусідів
	// (у реальному світі), або ми можемо "думати" віддалено, якщо Domain дозволяє (наш дозволяє).

	// Оскільки наш Domain.Actions(s) працює для будь-якого s (ми бачимо всю карту),
	// ми можемо прорахувати весь шлях A* в голові за один раз!
	// Але щоб це було "чесно" (і красиво візуально), будемо робити це покроково.

	targetItem := (*p.OpenSet)[0] // Peek (найкращий кандидат)
	target := targetItem.Value

	// ВАРІАНТ А: Ми вже у найкращій точці. Розширюємо її.
	if current.Equals(target) {
		heap.Pop(p.OpenSet) // Видаляємо з черги

		actions := domain.Actions(current)
		for _, act := range actions {
			neighbor := domain.Result(current, act)
			cost := domain.StepCost(current, act, neighbor)
			tentativeG := p.GScore[current] + cost

			if oldG, exists := p.GScore[neighbor]; !exists || tentativeG < oldG {
				p.CameFrom[neighbor] = current
				p.GScore[neighbor] = tentativeG
				fScore := tentativeG + domain.Heuristic(neighbor)

				heap.Push(p.OpenSet, &Item[S]{
					Value:    neighbor,
					Priority: fScore,
				})
				mem.Remember(neighbor) // Для візуалізації "відкритого списку"
			}
		}

		// Після розширення треба вибрати нову ціль
		if p.OpenSet.Len() == 0 {
			return "", ErrNoSolution
		}
		target = (*p.OpenSet)[0].Value
	}

	// ВАРІАНТ Б: Треба дійти до target (найкращого вузла у фронті)
	path, err := p.buildPathTree(current, target, domain)
	if err != nil {
		return "", fmt.Errorf("astar move error: %w", err)
	}
	p.PathBuffer = path

	if len(p.PathBuffer) > 0 {
		action := p.PathBuffer[0]
		p.PathBuffer = p.PathBuffer[1:]
		return action, nil
	}

	return "", nil
}

// buildPathTree знаходить шлях між двома вузлами у відомому дереві (LCA - Lowest Common Ancestor)
func (p *AStarPolicy[S]) buildPathTree(start, end S, domain planning.Domain[S]) ([]planning.Action, error) {
	// 1. Знаходимо шлях від кореня до start
	pathStart := []S{start}
	curr := start
	for !curr.Equals(p.CameFrom[curr]) { // Поки не корінь
		curr = p.CameFrom[curr]
		pathStart = append([]S{curr}, pathStart...) // Prepend
	}

	// 2. Знаходимо шлях від кореня до end
	pathEnd := []S{end}
	curr = end
	for !curr.Equals(p.CameFrom[curr]) {
		curr = p.CameFrom[curr]
		pathEnd = append([]S{curr}, pathEnd...)
	}

	// 3. Знаходимо спільного предка (LCA)
	i := 0
	for i < len(pathStart) && i < len(pathEnd) && pathStart[i].Equals(pathEnd[i]) {
		i++
	}
	// Останній спільний індекс був i-1.
	// LCA = pathStart[i-1]

	// 4. Формуємо шлях: Start -> ... -> LCA -> ... -> End
	var route []S

	// Вгору до LCA (зворотній порядок від start)
	// pathStart[i-1] це LCA. Ми йдемо від pathStart[len-1] до pathStart[i]
	for k := len(pathStart) - 1; k >= i; k-- {
		// Тут ми йдемо "вниз" по масиву, але "вгору" по дереву,
		// тому нам треба брати батька (k-1)
		// Але ми просто запишемо вузли.
	}

	// Простіше: шлях вузлів
	// Від Start вгору до LCA
	for k := len(pathStart) - 1; k >= i; k-- {
		route = append(route, pathStart[k-1]) // Батько
	}

	// Від LCA вниз до End
	for k := i; k < len(pathEnd); k++ {
		route = append(route, pathEnd[k])
	}

	// 5. Конвертуємо послідовність станів у дії
	var actions []planning.Action
	currentPos := start
	for _, nextPos := range route {
		// Знаходимо дію, яка веде з currentPos в nextPos
		availActions := domain.Actions(currentPos)
		found := false
		for _, act := range availActions {
			if domain.Result(currentPos, act).Equals(nextPos) {
				actions = append(actions, act)
				currentPos = nextPos
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("cannot move from %v to %v", currentPos, nextPos)
		}
	}

	return actions, nil
}

func (p *AStarPolicy[S]) Reset() {
	pq := make(PriorityQueue[S], 0)
	p.OpenSet = &pq
	p.CameFrom = make(map[S]S)
	p.GScore = make(map[S]float64)
	p.PathBuffer = nil
	p.initialized = false
}

// --- Priority Queue Implementation ---

type Item[S any] struct {
	Value    S
	Priority float64 // f-score
	Index    int
}

type PriorityQueue[S any] []*Item[S]

func (pq PriorityQueue[S]) Len() int           { return len(pq) }
func (pq PriorityQueue[S]) Less(i, j int) bool { return pq[i].Priority < pq[j].Priority }
func (pq PriorityQueue[S]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}
func (pq *PriorityQueue[S]) Push(x any) {
	n := len(*pq)
	item := x.(*Item[S])
	item.Index = n
	*pq = append(*pq, item)
}
func (pq *PriorityQueue[S]) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}
