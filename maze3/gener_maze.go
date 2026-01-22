package main

import "math/rand/v2"

// GenerateMaze створює випадковий лабіринт
// width, height мають бути непарними (наприклад, 11, 11)
func GenerateMaze(width, height int) []string {
	// 1. Ініціалізуємо сітку стінами
	grid := make([][]rune, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]rune, width)
		for x := 0; x < width; x++ {
			grid[y][x] = '#'
		}
	}

	// 2. Алгоритм прокладання шляхів (DFS)
	var visit func(x, y int)
	visit = func(x, y int) {
		grid[y][x] = '.' // Робимо прохід

		// Напрямки (вгору, вниз, вліво, вправо) у випадковому порядку
		dirs := []struct{ dx, dy int }{{0, -2}, {0, 2}, {-2, 0}, {2, 0}}
		rand.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })

		for _, d := range dirs {
			nx, ny := x+d.dx, y+d.dy
			// Перевірка меж
			if nx > 0 && nx < width-1 && ny > 0 && ny < height-1 && grid[ny][nx] == '#' {
				// Пробиваємо стіну між поточною і новою клітинкою
				grid[y+d.dy/2][x+d.dx/2] = '.'
				// Йдемо далі рекурсивно
				visit(nx, ny)
			}
		}
	}

	// Починаємо з точки (1,1)
	visit(1, 1)

	// 3. Ставимо Старт і Фініш
	grid[1][1] = 'S'
	// Вихід робимо в протилежному кутку (або близько до нього)
	grid[height-2][width-2] = 'E'

	// 4. Конвертуємо в []string
	result := make([]string, height)
	for y := 0; y < height; y++ {
		result[y] = string(grid[y])
	}
	return result
}
