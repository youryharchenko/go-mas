package ai

import (
	"context"
	"fmt"

	"github.com/youryharchenko/go-mas/planning"
)

// BFSPolicy реалізує алгоритм Breadth-First Search.
// Агент досліджує світ шарами.
type BFSPolicy[S interface {
	planning.State
	comparable
}] struct {
	Queue      []S               // Черга вузлів, які треба відвідати (Frontier)
	CameFrom   map[S]S           // Дерево шляхів (хто привів нас у цю точку)
	PathBuffer []planning.Action // Список дій для переміщення до наступної цілі
}

func NewBFS[S interface {
	planning.State
	comparable
}]() *BFSPolicy[S] {
	return &BFSPolicy[S]{
		Queue:    make([]S, 0),
		CameFrom: make(map[S]S),
	}
}

func (p *BFSPolicy[S]) Decide(ctx context.Context, current S, domain planning.Domain[S], mem planning.Memory[S]) (planning.Action, error) {

	// 0. Ініціалізація
	if len(p.Queue) == 0 && len(p.CameFrom) == 0 {
		p.Queue = append(p.Queue, current)
		// Start не має батька, або вказує сам на себе (як маркер кореня)
		p.CameFrom[current] = current
		mem.Remember(current)
	}

	// 1. Якщо у нас є запланований маршрут (перехід між гілками) - виконуємо його
	if len(p.PathBuffer) > 0 {
		action := p.PathBuffer[0]
		p.PathBuffer = p.PathBuffer[1:]
		return action, nil
	}

	// 2. Перевірка цілі
	if domain.IsGoal(current) {
		return "", ErrGoalReached
	}

	// 3. Основний цикл BFS
	// Ми беремо мету з голови черги.
	// Якщо ми вже там (current == target) -> розширюємо сусідів.
	// Якщо ми не там -> будуємо маршрут до target.

	if len(p.Queue) == 0 {
		return "", ErrNoSolution
	}

	target := p.Queue[0]

	// ВАРІАНТ А: Ми знаходимось у цільовій точці. Час досліджувати сусідів.
	if current.Equals(target) {
		// Видаляємо поточний вузол з черги (ми його обробили)
		p.Queue = p.Queue[1:]

		// Отримуємо сусідів
		actions := domain.Actions(current)
		for _, act := range actions {
			neighbor := domain.Result(current, act)

			// Якщо ми тут ще не були (і не планували бути)
			// Важливо: перевіряємо і Memory, і CameFrom (щоб не додавати дублікати в чергу)
			_, known := p.CameFrom[neighbor]
			if !known && !mem.HasVisited(neighbor) {
				p.CameFrom[neighbor] = current
				p.Queue = append(p.Queue, neighbor)
				// В BFS ми "запам'ятовуємо" клітинку, як тільки побачили її,
				// щоб інші гілки не намагалися її додати.
				mem.Remember(neighbor)
			}
		}

		// Після розширення, наша наступна мета - нова голова черги.
		// Ми повертаємось на початок Decide, щоб побудувати шлях до неї.
		// (Рекурсивний виклик або просто continue через структуру функції -
		// тут ми просто повернемо порожню дію, щоб викликатись на наступному тіку,
		// АБО, краще, перейдемо до логіки руху прямо зараз).
		if len(p.Queue) == 0 {
			return "", ErrNoSolution
		}
		target = p.Queue[0] // Нова ціль
	}

	// ВАРІАНТ Б: Нам треба дістатися від Current до Target.
	// Оскільки ми не вміємо телепортуватися, треба знайти шлях по дереву CameFrom.
	path, err := p.buildPathTree(current, target, domain)
	if err != nil {
		return "", fmt.Errorf("bfs pathfinding error: %w", err)
	}

	p.PathBuffer = path

	// Виконуємо перший крок
	if len(p.PathBuffer) > 0 {
		action := p.PathBuffer[0]
		p.PathBuffer = p.PathBuffer[1:]
		return action, nil
	}

	return "", nil
}

// buildPathTree знаходить шлях між двома вузлами у відомому дереві (LCA - Lowest Common Ancestor)
func (p *BFSPolicy[S]) buildPathTree(start, end S, domain planning.Domain[S]) ([]planning.Action, error) {
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

func (p *BFSPolicy[S]) Reset() {
	p.Queue = make([]S, 0)
	p.CameFrom = make(map[S]S)
	p.PathBuffer = nil
}
