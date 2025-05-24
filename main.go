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

const (
	width  = 80
	height = 24
)

type Cell struct {
	obstacle bool
	water    float64
	velY     float64
}

var grid [height][width]Cell
var (
	totalWater            float64 // 当前总水量（每帧计算）
	addedWaterThisFrame   float64 // 本帧新加的水量
	decayedWaterThisFrame float64 // 本帧衰减掉的水量
	totalDecayedWater     float64 // 累计衰减水量
)

func clearScreen() {
	cmd := exec.Command("clear") // or "cls" on Windows
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
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
	addedWaterThisFrame = 0 // 清零统计

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

// 计算每一格水的运动
func simulate() {
	newGrid := grid
	decayRate := 0.8          // 每帧衰减20%
	decayedWaterThisFrame = 0 // 本帧衰减统计清零

	for y := height - 2; y >= 0; y-- {
		for x := 0; x < width; x++ {
			cell := grid[y][x]

			// 碰到底部水量慢慢减少
			if y == height-2 && cell.water > 0 {
				oldWater := cell.water
				newWater := cell.water * decayRate
				if newWater < 0.01 {
					newWater = 0
				}
				decayedWaterThisFrame += oldWater - newWater
				newGrid[y][x].water = newWater
				continue
			}

			if cell.obstacle || cell.water <= 0 {
				continue
			}

			amount := cell.water

			// Try to flow down
			if !grid[y+1][x].obstacle && newGrid[y+1][x].water < 1.0 {
				flow := min(amount, 1.0-newGrid[y+1][x].water)
				newGrid[y+1][x].water += flow
				newGrid[y][x].water -= flow
				continue
			}

			// Try to flow left
			if x > 0 && !grid[y][x-1].obstacle && newGrid[y][x-1].water < 1.0 {
				flow := min(amount/2, 1.0-newGrid[y][x-1].water)
				newGrid[y][x-1].water += flow
				newGrid[y][x].water -= flow
			}

			// Try to flow right
			if x < width-1 && !grid[y][x+1].obstacle && newGrid[y][x+1].water < 1.0 {
				flow := min(amount/2, 1.0-newGrid[y][x+1].water)
				newGrid[y][x+1].water += flow
				newGrid[y][x].water -= flow
			}
		}
	}

	grid = newGrid
	totalDecayedWater += decayedWaterThisFrame
}

func draw() {
	ss := []rune(" `.^,:~\"<!ct+{i7?u30pw4A8DX%#HWM")
	maxIdx := len(ss) - 1

	var b strings.Builder
	b.WriteString("\x1b[H") // 光标回顶部，不清屏

	// 主画面
	for y := 0; y < height-1; y++ { // 留最后一行给统计
		for x := 0; x < width; x++ {
			cell := grid[y][x]

			if cell.obstacle {
				b.WriteByte('#')
			} else if cell.water > 0 {
				idx := int(cell.water * float64(maxIdx))
				if idx > maxIdx {
					idx = maxIdx
				} else if idx < 0 {
					idx = 0
				}
				b.WriteRune(ss[idx])
			} else {
				b.WriteByte(' ')
			}
		}
		b.WriteByte('\n')
	}

	// 统计行
	totalWater = 0
	for y := 0; y < height-1; y++ {
		for x := 0; x < width; x++ {
			totalWater += grid[y][x].water
		}
	}
	// 统计行内容
	// 第一行水量总览
	status1 := fmt.Sprintf(
		"Total Water: %.2f | Added This Frame: %.2f",
		totalWater, addedWaterThisFrame,
	)
	if len(status1) > width {
		status1 = status1[:width]
	}
	b.WriteString(status1)
	b.WriteByte('\n')

	// 第二行衰减统计
	status2 := fmt.Sprintf(
		"Decayed This Frame: %.2f | Total Decayed: %.2f",
		decayedWaterThisFrame, totalDecayedWater,
	)
	if len(status2) > width {
		status2 = status2[:width]
	}
	b.WriteString(status2)
	b.WriteByte('\n')

	fmt.Print(b.String())
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
	mapPath := ""
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <map/endoh1.c>")
		mapPath = "map/endoh1.c"
		//return
	} else {
		mapPath = os.Args[1]
	}

	loadMap(mapPath)
	draw()

	// 模拟循环（伪代码，按你已有逻辑调用即可）
	for {
		frameCount++
		if frameCount%5 == 0 {
			// 每隔 5 帧才添加一次水，减少水量突变
			addWater()
		}
		simulate()
		draw()
		time.Sleep(80 * time.Millisecond)
	}
	showCursor()
}
