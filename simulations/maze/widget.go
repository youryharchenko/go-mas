package maze

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// MazeBoard - це наш кастомний віджет
type MazeBoard struct {
	widget.BaseWidget

	// Дані для відображення
	Grid    []string
	WalkerX int
	WalkerY int

	Visited map[string]bool
}

func NewMazeBoard() *MazeBoard {
	m := &MazeBoard{}
	m.ExtendBaseWidget(m)
	return m
}

// UpdateState оновлює дані і перемальовує віджет
func (m *MazeBoard) UpdateState(grid []string, wx, wy int, visited map[string]bool) {

	m.Grid = grid
	m.WalkerX = wx
	m.WalkerY = wy
	m.Visited = visited
	m.Refresh() // Викликає CreateRenderer -> Refresh
}

func (m *MazeBoard) CreateRenderer() fyne.WidgetRenderer {
	return &mazeRenderer{board: m}
}

// --- RENDERER (Внутрішня кухня малювання) ---

type mazeRenderer struct {
	board   *MazeBoard
	objects []fyne.CanvasObject

	// Кеш для стін, щоб не перестворювати їх щоразу
	walls  []*canvas.Rectangle
	walker *canvas.Circle
	exit   *canvas.Rectangle
}

func (r *mazeRenderer) MinSize() fyne.Size {
	return fyne.NewSize(200, 200) // Мінімальний розмір поля
}

func (r *mazeRenderer) Layout(size fyne.Size) {
	// Цей метод викликається при зміні розміру вікна
	// Але ми будемо рахувати позиції в Refresh, тому тут пусто
}

func (r *mazeRenderer) Refresh() {
	// 1. Очищаємо список об'єктів
	r.objects = nil
	grid := r.board.Grid
	if len(grid) == 0 {
		return
	}

	rows := len(grid)
	cols := len(grid[0])

	// Розраховуємо розмір клітинки
	// Беремо меншу сторону, щоб сітка влізла і була квадратною
	canvasSize := r.board.Size()
	cellW := canvasSize.Width / float32(cols)
	cellH := canvasSize.Height / float32(rows)
	cellSize := cellW
	if cellH < cellW {
		cellSize = cellH
	}

	// Зсув, щоб відцентрувати лабіринт
	offsetX := (canvasSize.Width - (float32(cols) * cellSize)) / 2
	offsetY := (canvasSize.Height - (float32(rows) * cellSize)) / 2

	// 2. Малюємо Стіни та Вихід
	// (У реальному проекті це треба кешувати, але поки малюємо з нуля для простоти)
	for y, row := range grid {
		for x, char := range row {
			posX := offsetX + float32(x)*cellSize
			posY := offsetY + float32(y)*cellSize

			// Ключ для перевірки відвідування
			key := fmt.Sprintf("%d,%d", x, y)
			isVisited := false
			if v, ok := r.board.Visited[key]; ok {
				isVisited = v
			}

			// КОЛЬОРИ
			wallColor := color.RGBA{60, 60, 60, 255}    // Темно-сірий (Стіна)
			pathColor := color.RGBA{200, 200, 200, 255} // Світло-сірий (Відвіданий шлях)

			if char == '#' {
				wall := canvas.NewRectangle(color.RGBA{60, 60, 60, 255}) // Темно-сірий
				wall.Resize(fyne.NewSize(cellSize, cellSize))
				wall.Move(fyne.NewPos(posX, posY))
				r.objects = append(r.objects, wall)
			} else if char == 'E' {
				exit := canvas.NewRectangle(color.RGBA{0, 200, 0, 255}) // Зелений вихід
				exit.Resize(fyne.NewSize(cellSize, cellSize))
				exit.Move(fyne.NewPos(posX, posY))
				r.objects = append(r.objects, exit)
			} else if !isVisited {
				// 2. НЕВІДОМА ТЕРИТОРІЯ (Шлях, де ми ще не були)
				// Малюємо його ТАКИМ САМИМ кольором, як стіну (ховаємо)
				fog := canvas.NewRectangle(wallColor)
				fog.Resize(fyne.NewSize(cellSize, cellSize))
				fog.Move(fyne.NewPos(posX, posY))
				r.objects = append(r.objects, fog)
			} else {
				// 3. ВІДВІДАНИЙ ШЛЯХ (Світло)
				// Якщо ми тут були - малюємо світлий квадрат
				path := canvas.NewRectangle(pathColor)
				path.Resize(fyne.NewSize(cellSize, cellSize))
				path.Move(fyne.NewPos(posX, posY))
				r.objects = append(r.objects, path)
			}
		}
	}

	// 3. Малюємо Волкера
	walker := canvas.NewCircle(color.RGBA{255, 50, 50, 255}) // Червоний
	// Робимо його трохи меншим за клітинку
	padding := cellSize * 0.1
	walker.Resize(fyne.NewSize(cellSize-padding*2, cellSize-padding*2))

	//log.Println("Refresh", r.board.WalkerX, r.board.WalkerY)

	wx := offsetX + float32(r.board.WalkerX)*cellSize + padding
	wy := offsetY + float32(r.board.WalkerY)*cellSize + padding
	walker.Move(fyne.NewPos(wx, wy))

	r.objects = append(r.objects, walker)

	// Оновлюємо полотно
	fyne.Do(func() { canvas.Refresh(r.board) })

}

func (r *mazeRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *mazeRenderer) Destroy() {}
