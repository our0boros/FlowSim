package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
)

var (
	totalWater            float64
	addedWaterThisFrame   float64
	decayedWaterThisFrame float64
	totalDecayedWater     float64
)

const (
	Width  = 79
	Height = 24
)

// Particle 表示单个流体单元的状态
type Particle struct {
	water    float64
	x        float64 // 位置 x
	y        float64 // 位置 y
	vx       float64 // 水平方向速度
	vy       float64 // 垂直方向速度
	obstacle bool    // 是否是障碍物
}

// loadMap 读取文件，解析流体与障碍物初始状态
func loadMap(filePath string) ([]*Particle, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open map file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var particles []*Particle

	y := 0
	for scanner.Scan() {
		line := scanner.Text()
		for x := 0; x < len(line); x++ {
			ch := line[x]
			p := &Particle{
				x: float64(x),
				y: float64(y),
			}

			switch ch {
			case '#':
				p.obstacle = true
			case ' ':
				continue // 空格表示没有粒子
			default:
				p.water = 1.0
			}

			particles = append(particles, p)
		}
		y++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading map: %v", err)
	}

	return particles, nil
}

// updateParticles 更新所有粒子的状态：位置和速度，处理边界反弹和速度损耗
func updateParticles(particles []*Particle) {
	grid := buildGrid(particles)
	for _, p := range particles {
		applyGravity(p, grid)
		if p.obstacle {
			// 障碍物不移动
			continue
		}

		// 更新位置
		p.x += p.vx
		p.y += p.vy

		// 边界检测及反弹
		if p.x < 0 {
			p.x = 0
			p.vx = -p.vx * 0.8 // 反弹且损耗速度
		}
		if p.x >= Width {
			p.x = Width - 1
			p.vx = -p.vx * 0.8
		}
		if p.y < 0 {
			p.y = 0
			p.vy = -p.vy * 0.8
		}
		if p.y >= Height {
			p.y = Height - 1
			p.vy = -p.vy * 0.8
		}

		// TODO: 碰撞检测与挤压产生额外速度（需要空间划分实现）
	}
}

// 构造二维网格索引粒子，方便邻居查询
func buildGrid(particles []*Particle) [][]*Particle {
	grid := make([][]*Particle, Height)
	for i := range grid {
		grid[i] = make([]*Particle, Width)
	}

	for _, p := range particles {
		x := int(p.x)
		y := int(p.y)
		if x >= 0 && x < Width && y >= 0 && y < Height {
			grid[y][x] = p
		}
	}

	return grid
}

// 获取某粒子周围邻居（上下左右及斜角）
func getNeighbors(grid [][]*Particle, x, y int) []*Particle {
	var neighbors []*Particle
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			if nx >= 0 && nx < Width && ny >= 0 && ny < Height {
				if neighbor := grid[ny][nx]; neighbor != nil {
					neighbors = append(neighbors, neighbor)
				}
			}
		}
	}
	return neighbors
}

// 反弹逻辑：根据碰撞方向反转速度，并按损耗比例缩减
func reflectVelocity(p, obstacle *Particle) {
	// 简单反弹示例：反转vx和vy，损耗20%
	p.vx = -p.vx * 0.8
	p.vy = -p.vy * 0.8
}

// 挤压力：邻居对当前粒子速度的影响
func applyCompressionForce(p, neighbor *Particle) {
	dx := neighbor.x - p.x
	dy := neighbor.y - p.y

	distSq := dx*dx + dy*dy
	if distSq == 0 {
		return // 避免除零
	}

	// 计算归一化方向（单位向量）
	dist := math.Sqrt(distSq)
	nx := dx / dist
	ny := dy / dist

	// 挤压力随距离衰减
	force := 0.05 / distSq

	// 对向检测：
	// 如果邻居和当前粒子运动方向相反，则挤压力更显著表现为垂直推动
	dot := p.vx*nx + p.vy*ny // 点积
	if dot < 0 {
		// 转为垂直方向分量（此处简单用 y 增强）
		p.vy -= force * math.Abs(dot) * 0.5
	} else {
		// 正常挤压力分解到 x/y
		p.vx -= force * nx
		p.vy -= force * ny
	}
}

// 更新所有粒子速度（包括碰撞和挤压）
func updateVelocities(particles []*Particle) {
	grid := buildGrid(particles)

	for _, p := range particles {
		if p.obstacle {
			continue
		}
		x := int(p.x)
		y := int(p.y)

		neighbors := getNeighbors(grid, x, y)

		for _, n := range neighbors {
			if n.obstacle {
				reflectVelocity(p, n)
			} else {
				applyCompressionForce(p, n)
			}
		}
	}
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

const gravity = 0.08 // 每帧重力加速度
const damping = 0.2  // 被顶住时速度衰减系数

func applyGravity(p *Particle, grid [][]*Particle) {
	// 添加重力
	p.vy += gravity

	x := int(p.x)
	y := int(p.y)

	// 检查上方是否存在另一个粒子
	if y > 0 {
		above := grid[y-1][x]
		if above != nil {
			// 如果上方粒子速度 <= 当前，说明自己快，应该受阻
			if above.vy <= p.vy {
				p.vy *= damping // 被顶住时衰减垂直速度
			}
		}
	}
}

// renderParticles 在终端打印所有粒子，显示障碍物和运动方向
func renderParticles(particles []*Particle, debugMode bool) {
	ss := []rune(" `.^,:~\"<!ct+{i7?u30pw4A8DX%#HWM")
	maxIdx := len(ss) - 1

	// 初始化屏幕字符网格
	screen := make([][]rune, Height)
	for i := range screen {
		screen[i] = make([]rune, Width)
		for j := range screen[i] {
			screen[i][j] = ' ' // 默认是空格
		}
	}

	// 渲染粒子
	for _, p := range particles {
		x := int(p.x)
		y := int(p.y)

		if x < 0 || x >= Width || y < 0 || y >= Height-1 {
			continue
		}

		if p.obstacle {
			screen[y][x] = '#'
		} else if p.water > 0 {
			if debugMode {
				screen[y][x] = velocityArrow(p.vx, p.vy)
			} else {
				idx := int(p.water * float64(maxIdx))
				if idx > maxIdx {
					idx = maxIdx
				} else if idx < 0 {
					idx = 0
				}
				screen[y][x] = ss[idx]
			}
		}
	}

	// 构造输出字符串
	var b strings.Builder
	b.WriteString("\x1b[H") // 光标回到屏幕左上角

	// 绘制网格内容
	for _, row := range screen[:Height-1] {
		b.WriteString(string(row))
		b.WriteByte('\n')
	}

	// 统计行
	totalWater = 0
	for _, p := range particles {
		totalWater += p.water
	}
	status1 := fmt.Sprintf("Total Water: %.2f | Added This Frame: %.2f", totalWater, addedWaterThisFrame)
	status2 := fmt.Sprintf("Decayed This Frame: %.2f | Total Decayed: %.2f", decayedWaterThisFrame, totalDecayedWater)

	if len(status1) > Width {
		status1 = status1[:Width]
	}
	if len(status2) > Width {
		status2 = status2[:Width]
	}
	b.WriteString(status1 + "\n" + status2 + "\n")

	fmt.Print(b.String())
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <mapfile>")
		return
	}
	filePath := os.Args[1]

	particles, err := loadMap(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load map: %v\n", err)
		return
	}

	//// 初始化粒子位置（示例，需根据实际需求设置）
	//for i, p := range particles {
	//	p.x = float64(i % Width)
	//	p.y = float64(i / Width)
	//	// 初始速度随机或为0
	//	p.vx = 0
	//	p.vy = 0
	//}
	renderParticles(particles, false)
	// 清屏
	fmt.Print("\x1b[2J")

	// 模拟循环
	for {
		updateVelocities(particles)
		updateParticles(particles)
		renderParticles(particles, true)
		time.Sleep(50 * time.Millisecond)
	}
	time.Sleep(50 * time.Millisecond)
}
