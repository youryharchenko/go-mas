package maze

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/youryharchenko/go-mas/ai"
	"github.com/youryharchenko/go-mas/mas"
	"github.com/youryharchenko/go-mas/planning"
)

type PlannerWalker struct {
	mas.BaseAgent
	MazeID string

	// --- КОМПОНЕНТИ ІНТЕЛЕКТУ ---
	// Тепер ми використовуємо інтерфейси!
	Brain  planning.Policy[MazeState] // Стратегія (DFS)
	Memory planning.Memory[MazeState] // Пам'ять (Thread-safe)
	Domain planning.Domain[MazeState] // Фізика світу

	// Внутрішній стан
	CurrentState MazeState
	Steps        int
	Solved       bool
}

func (w *PlannerWalker) Bind(sys *mas.System, inbox <-chan mas.Envelope, me mas.Agent) {
	// Викликаємо базовий Bind
	w.BaseAgent.Bind(sys, inbox, me)

	// Ініціалізуємо компоненти, якщо їх немає
	if w.Memory == nil {
		w.Memory = NewMazeMemory()
	}
	if w.Brain == nil {
		w.Brain = ai.NewDFS[MazeState]()
	}

	// ВАЖЛИВО: Нам треба отримати доступ до Domain (MazeAgent).
	// Оскільки ми в одному процесі, ми можемо знайти його через System.
	// Це трохи порушує чисту акторну модель, але необхідно для Planner-абстракції.
	// Ми зробимо це ліниво (Lazy) у методі Plan, або тут, якщо Maze вже існує.
}

func (w *PlannerWalker) Plan(ctx context.Context, msg mas.Envelope) ([]mas.Action, error) {

	// Ініціалізація Домену (Лінива)
	if w.Domain == nil {
		agent, ok := w.Sys().GetAgent(w.MazeID)
		if ok {
			if domain, ok := agent.(planning.Domain[MazeState]); ok {
				w.Domain = domain
			} else {
				return nil, fmt.Errorf("agent %s does not implement Domain[MazeState]", w.MazeID)
			}
		}
	}

	// 1. ОБРОБКА TICK (Прийняття рішень)
	if msg.Payload == "TICK" {
		if w.Solved || w.Domain == nil {
			return nil, nil
		}

		// --- OODA LOOP: Decide ---
		// Агент запитує у Політики: "Що робити?"
		log.Println("CurrentState:", w.CurrentState)
		actionName, err := w.Brain.Decide(ctx, w.CurrentState, w.Domain, w.Memory)

		if err == ai.ErrGoalReached {
			// Якщо Політика каже, що ми прийшли - фіксуємо перемогу
			// (Хоча технічно це має статися після Action, але DFS знає це заздалегідь)
			return nil, nil // Чекаємо, поки MoveResult підтвердить це
		}

		if err != nil {
			// Глухий кут або помилка
			// mas.SayLog("Brain error: %v", err)
			return nil, nil
		}

		// --- Act ---
		// Конвертуємо planning.Action (string) у MoveRequest (struct)
		var dir Direction
		switch actionName {
		case "UP":
			dir = DirUp
		case "DOWN":
			dir = DirDown
		case "LEFT":
			dir = DirLeft
		case "RIGHT":
			dir = DirRight
		}

		return []mas.Action{
			func(a mas.Agent, sys *mas.System) error {
				req := MoveRequest{Dir: dir}
				log.Println("MoveRequest:", req)
				return sys.Send(ctx, w.IDVal, w.MazeID, req)
			},
		}, nil
	}

	// 2. ОБРОБКА РЕЗУЛЬТАТУ (Сприйняття)
	if res, ok := msg.Payload.(MoveResult); ok {
		log.Println("MoveResult:", res)
		if res.Success {
			// Оновлюємо стан тільки якщо хід успішний
			return []mas.Action{
				mas.MutateState(func(a any) {
					walker := a.(*PlannerWalker)
					walker.Steps++

					// ОНОВЛЮЄМО КООРДИНАТИ З ПОВІДОМЛЕННЯ
					walker.CurrentState.X = res.State.X
					walker.CurrentState.Y = res.State.Y

					// Запам'ятовуємо новий стан
					walker.Memory.Remember(walker.CurrentState)

				}),
				func(a mas.Agent, sys *mas.System) error {
					walker := a.(*PlannerWalker)
					if res.IsFinished {
						walker.Solved = true
						sys.Send(ctx, w.IDVal, "console", fmt.Sprintf("VICTORY in %d steps!", walker.Steps))

						// Авто-рестарт
						go func() {
							time.Sleep(3 * time.Second)
							sys.Send(context.Background(), w.IDVal, w.MazeID, "NEW")
						}()
					} else {
						// Оновлюємо пам'ять про новий стан
						// (Це важливо! Ми додаємо у пам'ять тільки ФАКТИЧНО відвідані стани)

						// !!! ТУТ ПРОБЛЕМА: Ми не знаємо нових координат без MoveResult !!!
						// Треба оновити MoveResult у models.go
					}
					return nil
				},
			}, nil
		}
	}

	// 3. ОБРОБКА RESET
	if msg.Payload == "RESET" {
		return []mas.Action{
			mas.MutateState(func(a any) {
				w := a.(*PlannerWalker)
				w.Memory.Clear()
				w.Steps = 0
				w.Solved = false
				w.CurrentState = MazeState{X: 1, Y: 1} // Start

				// Скидаємо мозок
				/* if dfs, ok := w.Brain.(*ai.DFSPolicy[MazeState]); ok {
					dfs.Reset()
				} */
				w.Brain.Reset()
				// Стартова точка в пам'ять
				w.Memory.Remember(w.CurrentState)
			}),
		}, nil
	}

	// --- НОВЕ: ЗМІНА ПОЛІТИКИ ---
	// Перевіряємо, чи це рядок і чи починається він з "POLICY:"
	if str, ok := msg.Payload.(string); ok && strings.HasPrefix(str, "POLICY:") {
		policyName := strings.TrimPrefix(str, "POLICY:")

		return []mas.Action{
			mas.MutateState(func(a any) {
				walker := a.(*PlannerWalker)

				// 1. Міняємо мозок
				switch policyName {
				case "DFS":
					walker.Brain = ai.NewDFS[MazeState]()
				case "BFS":
					walker.Brain = ai.NewBFS[MazeState]() // Коли напишете BFS
				case "AStar": // <--- НОВЕ
					walker.Brain = ai.NewAStar[MazeState]()
				case "Random":
					// Можна швидко зробити заглушку або реальний RandomPolicy
					// walker.Brain = policies.NewRandom[MazeState]()
					// Поки скинемо на DFS, якщо іншого нема
					walker.Brain = ai.NewRandom[MazeState]()
				default:
					// Якщо не знаємо, залишаємо як є або DFS
				}

				// 2. Скидаємо стан (Reset)
				// Дублюємо логіку RESET, щоб агент почав нове життя з новим мозком
				walker.Memory.Clear()
				walker.Steps = 0
				walker.Solved = false
				walker.CurrentState = MazeState{X: 1, Y: 1}
				walker.Memory.Remember(walker.CurrentState)
			}),
			mas.Send("console", fmt.Sprintf("Planner: Switched brain to %s", policyName)),
		}, nil
	}

	return nil, nil
}

// Sys повертає посилання на систему (helper)
/* func (w *PlannerWalker) Sys() *mas.System {
	// Це поле приватне в BaseAgent, треба або геттер, або змінити видимість.
	// Припустимо, ми додали метод System() в BaseAgent.
	// Якщо ні - треба додати.
	// Поки повернемо nil, це псевдокод.
	return nil
} */
