package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	width  int
	height int
)

type Cell struct {
	obstacle bool
	water    float64
	velX     float64
	velY     float64
}

var grid [][]Cell
var (
	totalWater            float64
	addedWaterThisFrame   float64
	decayedWaterThisFrame float64
	totalDecayedWater     float64
)

func clearScreen() {
	cmd := exec.Command("clear") // or "cls" on Windows
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

func getMapSize(path string) (int, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	maxWidth := 0
	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, err
	}

	return maxWidth, lineCount, nil
}

// 初始化动态 grid
func initGrid(w, h int) {
	grid = make([][]Cell, h)
	for y := 0; y < h; y++ {
		grid[y] = make([]Cell, w)
	}
}

func loadMap(path string) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open map file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	clearScreen()
	scanner := bufio.NewScanner(file)
	y := 0
	for scanner.Scan() && y < height {
		line := scanner.Text()
		for x := 0; x < len(line) && x < width; x++ {
			switch line[x] {
			case '#':
				grid[y][x].obstacle = true
			case ' ':
				// nothing to do
			default:
				grid[y][x].water = 1.0
			}
		}
		y++
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading map: %v\n", err)
	}
}

func addWater() {
	addedWaterThisFrame = 0
	y := 0
	x := rand.Intn(width)
	waterAmount := 0.1 + rand.Float64()*0.4

	if !grid[y][x].obstacle {
		oldWater := grid[y][x].water
		grid[y][x].water += waterAmount
		if grid[y][x].water > 1.0 {
			grid[y][x].water = 1.0
		}
		addedWaterThisFrame += grid[y][x].water - oldWater

		if grid[y][x].water >= 0.8 {
			radius := rand.Intn(4)
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					nx, ny := x+dx, y+dy
					if nx >= 0 && nx < width && ny >= 0 && ny < height {
						if math.Sqrt(float64(dx*dx+dy*dy)) <= float64(radius) {
							if !grid[ny][nx].obstacle {
								oldW := grid[ny][nx].water
								amount := 0.05 + rand.Float64()*0.1
								grid[ny][nx].water += amount
								if grid[ny][nx].water > 1.0 {
									grid[ny][nx].water = 1.0
								}
								addedWaterThisFrame += grid[ny][nx].water - oldW
							}
						}
					}
				}
			}
		}
	}
}

func simulate() {
	newGrid := make([][]Cell, height)
	for y := range newGrid {
		newGrid[y] = make([]Cell, width)
		copy(newGrid[y], grid[y])
	}
	decayedWaterThisFrame = 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := grid[y][x]
			if cell.obstacle || cell.water <= 0 {
				continue
			}

			newX := int(float64(x) + cell.velX)
			newY := int(float64(y) + cell.velY)

			// 判断是否撞墙
			isWall := false
			if newX < 0 || newX >= width || newY < 0 || newY >= height {
				isWall = true
			} else if grid[newY][newX].obstacle {
				isWall = true
			}

			if isWall {
				// 撞墙/障碍反弹，带损耗
				// 速度反向并乘以损耗系数
				var loss float64
				if newX < 0 || newX >= width || newY < 0 || newY >= height {
					// 碰边界墙，损耗更大
					loss = 0.6
				} else {
					// 碰障碍物，损耗稍小
					loss = 0.8
				}

				reflectX := -cell.velX * loss
				reflectY := -cell.velY * loss

				// 反弹力度也和水量有关，水越多能量越大，可调节如下：
				reboundFactor := 0.5 + 0.5*min(1.0, cell.water) // 0.5~1.0

				newGrid[y][x].velX = reflectX * reboundFactor
				newGrid[y][x].velY = reflectY * reboundFactor

				// 水量保持不变
				newGrid[y][x].water += cell.water
			} else {
				// 正常移动
				if newGrid[newY][newX].water < 1.0 {
					flow := min(cell.water, 1.0-newGrid[newY][newX].water)
					newGrid[newY][newX].water += flow
					newGrid[newY][newX].velX = cell.velX * 0.98          // 阻尼
					newGrid[newY][newX].velY = (cell.velY + 0.05) * 0.98 // 加重力模拟

					newGrid[y][x].water -= flow
				}
			}
		}
	}

	grid = newGrid
}
func velocityArrow(vx, vy float64) rune {
	const threshold = 0.1 // 低于这个速度视为静止

	if math.Hypot(vx, vy) < threshold {
		return '·'
	}

	angle := math.Atan2(vy, vx) * 180 / math.Pi // 角度制，-180 ~ +180

	switch {
	case angle >= -22.5 && angle < 22.5:
		return '→'
	case angle >= 22.5 && angle < 67.5:
		return '↗'
	case angle >= 67.5 && angle < 112.5:
		return '↑'
	case angle >= 112.5 && angle < 157.5:
		return '↖'
	case angle >= 157.5 || angle < -157.5:
		return '←'
	case angle >= -157.5 && angle < -112.5:
		return '↙'
	case angle >= -112.5 && angle < -67.5:
		return '↓'
	case angle >= -67.5 && angle < -22.5:
		return '↘'
	default:
		return '·'
	}
}

func draw() {
	ss := []rune(" `.^,:~\"<!ct+{i7?u30pw4A8DX%#HWM")
	maxIdx := len(ss) - 1

	var b strings.Builder
	b.WriteString("\x1b[H")

	for y := 0; y < height-1; y++ {
		for x := 0; x < width; x++ {
			cell := grid[y][x]
			debugMode := true // 你可以定义成全局变量或者函数参数控制是否显示方向

			if cell.obstacle {
				b.WriteByte('#')
			} else if cell.water > 0 {
				if debugMode {
					b.WriteRune(velocityArrow(cell.velX, cell.velY))
				} else {
					idx := int(cell.water * float64(maxIdx))
					if idx > maxIdx {
						idx = maxIdx
					} else if idx < 0 {
						idx = 0
					}
					b.WriteRune(ss[idx])
				}
			} else {
				b.WriteByte(' ')
			}

		}
		b.WriteByte('\n')
	}

	totalWater = 0
	for y := 0; y < height-1; y++ {
		for x := 0; x < width; x++ {
			totalWater += grid[y][x].water
		}
	}

	status1 := fmt.Sprintf("Total Water: %.2f | Added This Frame: %.2f", totalWater, addedWaterThisFrame)
	if len(status1) > width {
		status1 = status1[:width]
	}
	b.WriteString(status1)
	b.WriteByte('\n')

	status2 := fmt.Sprintf("Decayed This Frame: %.2f | Total Decayed: %.2f", decayedWaterThisFrame, totalDecayedWater)
	if len(status2) > width {
		status2 = status2[:width]
	}
	b.WriteString(status2)
	b.WriteByte('\n')

	fmt.Print(b.String())
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func hideCursor() {
	fmt.Print("\x1b[?25l")
}

func showCursor() {
	fmt.Print("\x1b[?25h")
}

var frameCount int

func main() {
	hideCursor()
	defer showCursor()

	mapPath := "map/endoh1.c"
	if len(os.Args) > 1 {
		mapPath = os.Args[1]
	}

	var err error
	width, height, err = getMapSize(mapPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get map size: %v\n", err)
		os.Exit(1)
	}

	initGrid(width, height)
	loadMap(mapPath)

	for {
		frameCount++
		if frameCount%5 == 0 {
			addWater()
		}
		simulate()
		draw()
		time.Sleep(80 * time.Millisecond)
	}
}
