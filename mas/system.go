package mas

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"sync"
)

type System struct {
	mu       sync.RWMutex
	agents   map[string]Agent         // Тут живуть типи
	registry map[string]chan Envelope // Тут живуть канали (runtime)

	filename string // Куди зберігати dump

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Option - функціональна опція для налаштування системи.
type Option func(*System)

// WithPersistence налаштовує шлях до файлу збереження стану.
func WithPersistence(filename string) Option {
	return func(s *System) {
		s.filename = filename
	}
}

// WithContext дозволяє передати батьківський контекст (наприклад, для тестів або signal.Notify).
func WithContext(ctx context.Context) Option {
	return func(s *System) {
		// Перестворюємо контекст з cancel, базуючись на батьківському
		s.ctx, s.cancel = context.WithCancel(ctx)
	}
}

// NewSystem створює новий екземпляр системи.
// Приймає список опцій для конфігурації.
func NewSystem(opts ...Option) *System {
	// 1. Значення за замовчуванням
	defaultCtx, defaultCancel := context.WithCancel(context.Background())

	s := &System{
		agents:   make(map[string]Agent),
		registry: make(map[string]chan Envelope),
		//filename: "mas_state.gob", // Дефолтне ім'я файлу
		ctx:    defaultCtx,
		cancel: defaultCancel,
	}

	// 2. Застосування опцій користувача
	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *System) Context() context.Context {
	return s.ctx
}

// Startup - завантаження світу
func (s *System) Startup() error {

	if len(s.filename) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.filename)
	if os.IsNotExist(err) {
		return nil // Файлу немає, починаємо з чистого аркуша
	} else if err != nil {
		return err
	}
	defer file.Close()

	// 1. Магія GOB: відновлюємо карту інтерфейсів
	// Важливо: типи агентів мають бути зареєстровані через gob.Register() у init()
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&s.agents); err != nil {
		return err
	}
	log.Println("Resurrection")
	// 2. Оживлення (Resurrection)
	for id, agent := range s.agents {
		log.Println(agent.ID())
		// Створюємо інфраструктуру, яку GOB не зберіг
		inbox := make(chan Envelope, 100)
		s.registry[id] = inbox

		// Прив'язуємо агента до системи
		agent.Bind(s, inbox, agent)

		// Запускаємо
		s.wg.Add(1)
		go func(a Agent) {
			defer s.wg.Done()
			a.Run(s.ctx)
		}(agent)
	}

	return nil
}

// Shutdown - збереження світу
func (s *System) Shutdown() error {

	log.Println("System begin Shutdown")
	// 1. Зупинка всіх процесів
	s.cancel()
	s.wg.Wait()

	if len(s.filename) == 0 {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 2. Запис у файл
	file, err := os.Create(s.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	// Ми просто пишемо всю мапу агентів.
	// GOB сам збереже конкретні структури, сховані за інтерфейсом Agent.
	return encoder.Encode(s.agents)
}

func (s *System) GetAgent(id string) (Agent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agent, exists := s.agents[id]
	return agent, exists
}

// Spawn реєструє нового агента в системі та запускає його цикл обробки.
func (s *System) Spawn(agent Agent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := agent.ID()

	// 1. Валідація: Перевірка на унікальність ID
	// Це запобігає конфліктам адресації та перезапису стану
	if _, exists := s.agents[id]; exists {
		return fmt.Errorf("spawn failed: agent with ID '%s' already exists", id)
	}

	// 2. Ініціалізація інфраструктури (Транспорт)
	// Створюємо буферизований канал. Розмір буфера (100) можна винести в конфіг,
	// але для MVP це нормальне значення, щоб згладжувати пікові навантаження.
	inbox := make(chan Envelope, 100)

	// 3. Реєстрація
	// s.registry потрібен для маршрутизації (Send)
	s.registry[id] = inbox
	// s.agents потрібен для GOB-серіалізації (Shutdown) та GetAgent
	s.agents[id] = agent

	// 4. Прив'язка (Binding)
	// Впроваджуємо залежності (Dependency Injection) в структуру агента.
	// Це наповнює приватні поля (sys, inbox), які GOB ігнорує.
	agent.Bind(s, inbox, agent)

	// 5. Запуск (Execution)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done() // Сигналізуємо про завершення при виході

		// Запускаємо Run із системним контекстом.
		// Якщо s.Shutdown() скасує контекст, агент отримає сигнал ctx.Done()
		if err := agent.Run(s.ctx); err != nil {
			// Тут можна підключити системний логер помилок
			// fmt.Printf("Agent %s crashed: %v\n", id, err)
		}
	}()

	return nil
}

// Send відправляє повідомлення від одного агента іншому.
// Ця операція є потокобезпечною.
//
// Аргументи:
//
//	ctx     - Контекст виконання (можна використати для тайм-ауту: context.WithTimeout).
//	fromID  - ID відправника.
//	toID    - ID отримувача.
//	payload - Корисне навантаження (суть задачі).
func (s *System) Send(ctx context.Context, fromID, toID string, payload any) error {
	// 1. Пошук адресата
	// Використовуємо RLock, бо це операція читання, яка відбувається дуже часто.
	s.mu.RLock()
	ch, exists := s.registry[toID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("send failed: agent '%s' not found", toID)
	}

	// 2. Формування конверта
	env := Envelope{
		From:    fromID,
		To:      toID,
		Payload: payload,
		// Metadata можна додати тут, якщо потрібно (наприклад, timestamp)
	}

	// 3. Доставка з урахуванням Backpressure (зворотного тиску)
	select {
	case ch <- env:
		// Успішно поклали в канал
		//log.Println("Send Success:", env)
		return nil

	case <-ctx.Done():
		// Відправник (caller) скасував операцію або вийшов час (timeout)
		return fmt.Errorf("send canceled by caller: %w", ctx.Err())

	case <-s.ctx.Done():
		// Сама система вимикається, канали можуть бути вже закриті або неактивні
		return fmt.Errorf("system is shutting down")
	}
}

// Kill примусово видаляє агента з системи (пам'яті та реєстру).
// Корисно для тимчасових агентів (GUI, Debug), які не треба зберігати.
func (s *System) Kill(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.agents, id)
	delete(s.registry, id)
	// Якщо треба, тут можна закрити канал inbox, але обережно
}
