package main

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	defaultCellSize        = 2
	initialCellProbability = 10
	logFrequency           = 100
	autoExitSeconds        = 2000
	minCellSize            = 1
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	game := &Game{
		fullscreen:   true,
		cellSize:     defaultCellSize,
		gameStart:    time.Now(),
		lastFPSCheck: time.Now(),
	}

	ebiten.SetWindowTitle("Game of Life")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowDecorated(true)
	ebiten.SetFullscreen(game.fullscreen)

	width, height := ebiten.ScreenSizeInFullscreen()
	game.windowHeight = height
	game.windowWidth = width
	log.Printf("Window size: %dx%d", width, height)

	rand.Seed(time.Now().UnixNano())
	game.initBoard()
	game.initFont()

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}

type Game struct {
	windowWidth  int
	windowHeight int
	gridWidth    int
	gridHeight   int
	board        []bool
	nextBoard    []bool
	cellAge      []uint8
	nextCellAge  []uint8
	gameStart    time.Time
	cycleCount   int64
	fullscreen   bool
	cellSize     int
	workerPool   *WorkerPool
	
	// Statistics
	population   int
	births       int
	deaths       int
	lastFPSCheck time.Time
	frameCount   int
	currentFPS   float64
	fontFace     font.Face
	
	// Evolution state
	isEvolving      bool
	evolvingMessage string
	evolutionMode   EvolutionMode
	showHelp        bool
}

type EvolutionMode int

const (
	ModeMovers EvolutionMode = iota
	ModeOscillators
	ModeChaos
	ModeGrowth
)

func (m EvolutionMode) String() string {
	switch m {
	case ModeMovers:
		return "Movers (Gliders/Spaceships)"
	case ModeOscillators:
		return "Oscillators (Blinkers/Pulsars)"
	case ModeChaos:
		return "Chaos (Maximum Activity)"
	case ModeGrowth:
		return "Growth (Expanding Patterns)"
	default:
		return "Unknown"
	}
}

type WorkerPool struct {
	jobs    chan job
	workers int
	wg      sync.WaitGroup
}

type job struct {
	startX, endX int
	game         *Game
}

func newWorkerPool(numWorkers int) *WorkerPool {
	pool := &WorkerPool{
		jobs:    make(chan job, numWorkers),
		workers: numWorkers,
	}
	
	for i := 0; i < numWorkers; i++ {
		go pool.worker()
	}
	
	return pool
}

func (pool *WorkerPool) worker() {
	for j := range pool.jobs {
		pool.processJob(j)
		pool.wg.Done()
	}
}

func (pool *WorkerPool) processJob(j job) {
	width := j.game.gridWidth
	height := j.game.gridHeight
	
	for x := j.startX; x < j.endX; x++ {
		// Process edges with bounds checking
		if x == 0 || x == width-1 {
			for y := 0; y < height; y++ {
				pool.processCell(j.game, x, y, true)
			}
			continue
		}
		
		// Top edge
		pool.processCell(j.game, x, 0, true)
		
		// Inner cells - no bounds checking needed
		for y := 1; y < height-1; y++ {
			pool.processCell(j.game, x, y, false)
		}
		
		// Bottom edge
		pool.processCell(j.game, x, height-1, true)
	}
}

func (pool *WorkerPool) processCell(game *Game, x, y int, checkBounds bool) {
	idx := y*game.gridWidth + x
	alive := game.board[idx]
	
	var count int
	if checkBounds {
		count = game.countNeighborsWithBounds(x, y)
	} else {
		count = game.countNeighborsNoBounds(x, y)
	}
	
	if alive && (count == 2 || count == 3) {
		game.nextBoard[idx] = true
		// Increment age, cap at 255
		if game.cellAge[idx] < 255 {
			game.nextCellAge[idx] = game.cellAge[idx] + 1
		} else {
			game.nextCellAge[idx] = 255
		}
	} else if !alive && count == 3 {
		game.nextBoard[idx] = true
		game.nextCellAge[idx] = 1 // Newborn cell
	} else {
		game.nextBoard[idx] = false
		game.nextCellAge[idx] = 0 // Dead cell
	}
}

func (pool *WorkerPool) execute(game *Game) {
	rowsPerWorker := game.gridWidth / pool.workers
	
	pool.wg.Add(pool.workers)
	
	for w := 0; w < pool.workers; w++ {
		startX := w * rowsPerWorker
		endX := startX + rowsPerWorker
		if w == pool.workers-1 {
			endX = game.gridWidth
		}
		
		pool.jobs <- job{
			startX: startX,
			endX:   endX,
			game:   game,
		}
	}
	
	pool.wg.Wait()
}

func (game *Game) initBoard() {
	game.gridWidth = game.windowWidth / game.cellSize
	game.gridHeight = game.windowHeight / game.cellSize
	size := game.gridWidth * game.gridHeight

	game.board = make([]bool, size)
	game.nextBoard = make([]bool, size)
	game.cellAge = make([]uint8, size)
	game.nextCellAge = make([]uint8, size)

	for i := 0; i < size; i++ {
		game.board[i] = rand.Intn(initialCellProbability) == 1
		if game.board[i] {
			game.cellAge[i] = 1
		}
	}
	
	if game.workerPool == nil {
		game.workerPool = newWorkerPool(runtime.NumCPU())
	}
	
	game.updatePopulation()
}

func (game *Game) initFont() {
	tt, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal(err)
	}
	game.fontFace = truetype.NewFace(tt, &truetype.Options{
		Size:    16,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}

func (game *Game) updatePopulation() {
	count := 0
	for i := 0; i < len(game.board); i++ {
		if game.board[i] {
			count++
		}
	}
	game.population = count
}

func (game *Game) Update() error {
	game.cycleCount++
	game.logProgress()
	game.checkAutoExit()
	game.handleInput()
	game.evolveBoard()
	game.updateFPS()
	return nil
}

func (game *Game) updateFPS() {
	game.frameCount++
	now := time.Now()
	elapsed := now.Sub(game.lastFPSCheck).Seconds()
	
	if elapsed >= 1.0 {
		game.currentFPS = float64(game.frameCount) / elapsed
		game.frameCount = 0
		game.lastFPSCheck = now
	}
}

func (game *Game) logProgress() {
	if game.cycleCount%logFrequency == 0 {
		elapsed := time.Since(game.gameStart).Seconds()
		rate := game.cycleCount / int64(elapsed)
		log.Printf("Total cycles: %d, rate: %d cycles/s", game.cycleCount, rate)
	}
}

func (game *Game) checkAutoExit() {
	elapsed := time.Since(game.gameStart).Seconds()
	if elapsed > autoExitSeconds {
		os.Exit(0)
	}
}

func (game *Game) handleInput() {
	game.windowWidth, game.windowHeight = ebiten.WindowSize()

	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		os.Exit(0)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		game.fullscreen = !game.fullscreen
		ebiten.SetFullscreen(game.fullscreen)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyNumpadAdd) {
		game.cellSize++
		game.initBoard()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyNumpadSubtract) {
		if game.cellSize > minCellSize {
			game.cellSize--
			game.initBoard()
		}
	}
	
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		if !game.isEvolving {
			game.isEvolving = true
			game.evolvingMessage = "🧬 Evolving pattern..."
			go game.evolveInterestingPatternAsync()
		}
	}
	
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		// Cycle through evolution modes
		game.evolutionMode = (game.evolutionMode + 1) % 4
		log.Printf("Evolution mode: %s", game.evolutionMode.String())
	}
	
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		game.showHelp = !game.showHelp
	}
}

func (game *Game) evolveBoard() {
	oldPopulation := game.population
	
	game.workerPool.execute(game)
	game.board, game.nextBoard = game.nextBoard, game.board
	game.cellAge, game.nextCellAge = game.nextCellAge, game.cellAge
	
	// Update statistics
	game.updatePopulation()
	newPopulation := game.population
	
	if newPopulation > oldPopulation {
		game.births = newPopulation - oldPopulation
		game.deaths = 0
	} else {
		game.deaths = oldPopulation - newPopulation
		game.births = 0
	}
}

// countNeighborsNoBounds - optimized for inner cells (no bounds checking)
func (game *Game) countNeighborsNoBounds(x, y int) int {
	width := game.gridWidth
	idx := y*width + x
	
	count := 0
	// Unrolled loop for maximum speed
	if game.board[idx-width-1] { count++ } // top-left
	if game.board[idx-width] { count++ }   // top
	if game.board[idx-width+1] { count++ } // top-right
	if game.board[idx-1] { count++ }       // left
	if game.board[idx+1] { count++ }       // right
	if game.board[idx+width-1] { count++ } // bottom-left
	if game.board[idx+width] { count++ }   // bottom
	if game.board[idx+width+1] { count++ } // bottom-right
	
	return count
}

// countNeighborsWithBounds - for edge cells
func (game *Game) countNeighborsWithBounds(x, y int) int {
	count := 0
	width := game.gridWidth
	height := game.gridHeight
	
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx := x + dx
			ny := y + dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				if game.board[ny*width+nx] {
					count++
				}
			}
		}
	}
	return count
}

func (game *Game) countNeighbors(x, y int) int {
	width := game.gridWidth
	height := game.gridHeight
	
	// Use optimized version for inner cells
	if x > 0 && x < width-1 && y > 0 && y < height-1 {
		return game.countNeighborsNoBounds(x, y)
	}
	
	return game.countNeighborsWithBounds(x, y)
}

func (game *Game) Draw(screen *ebiten.Image) {
	width := game.gridWidth
	for i := 0; i < len(game.board); i++ {
		if game.board[i] {
			x := float32((i % width) * game.cellSize)
			y := float32((i / width) * game.cellSize)
			size := float32(game.cellSize)
			
			// Color based on age: young cells are bright, old cells fade to deeper colors
			age := game.cellAge[i]
			col := getColorForAge(age)
			
			vector.DrawFilledRect(screen, x, y, size, size, col, false)
		}
	}
	
	game.drawStats(screen)
}

func (game *Game) drawStats(screen *ebiten.Image) {
	// Semi-transparent background for stats
	vector.DrawFilledRect(screen, 10, 10, 250, 110, color.RGBA{0, 0, 0, 180}, false)
	
	// Draw statistics
	stats := []string{
		fmt.Sprintf("Generation: %d", game.cycleCount),
		fmt.Sprintf("Population: %d", game.population),
		fmt.Sprintf("Births: %d", game.births),
		fmt.Sprintf("Deaths: %d", game.deaths),
		fmt.Sprintf("FPS: %.1f", game.currentFPS),
	}
	
	// Simple text rendering using embedded font
	faceSource, _ := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	face := &text.GoTextFace{
		Source: faceSource,
		Size:   16,
	}
	
	for i, stat := range stats {
		op := &text.DrawOptions{}
		op.GeoM.Translate(15, float64(15+i*20))
		op.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, stat, face, op)
	}
	
	// Show evolution status if evolving
	if game.isEvolving && game.evolvingMessage != "" {
		// Draw in upper right corner using actual screen bounds
		bounds := screen.Bounds()
		screenWidth := float32(bounds.Dx())
		
		boxWidth := float32(320)
		boxHeight := float32(50)
		boxX := screenWidth - boxWidth - 10
		boxY := float32(10)
		
		// Background box
		vector.DrawFilledRect(screen, boxX, boxY, boxWidth, boxHeight, 
			color.RGBA{0, 0, 0, 200}, false)
		
		// Border
		vector.StrokeRect(screen, boxX, boxY, boxWidth, boxHeight,
			2, color.RGBA{0, 255, 255, 255}, false)
		
		// Message text
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(boxX+10), float64(boxY+20))
		op.ColorScale.ScaleWithColor(color.RGBA{0, 255, 255, 255})
		text.Draw(screen, game.evolvingMessage, face, op)
	}
	
	// Show help screen if F1 pressed
	if game.showHelp {
		game.drawHelp(screen, face)
	}
}

func (game *Game) drawHelp(screen *ebiten.Image, face *text.GoTextFace) {
	bounds := screen.Bounds()
	screenWidth := float32(bounds.Dx())
	screenHeight := float32(bounds.Dy())
	
	// Center overlay
	boxWidth := float32(600)
	boxHeight := float32(400)
	boxX := (screenWidth - boxWidth) / 2
	boxY := (screenHeight - boxHeight) / 2
	
	// Semi-transparent background
	vector.DrawFilledRect(screen, boxX, boxY, boxWidth, boxHeight,
		color.RGBA{0, 0, 0, 230}, false)
	
	// Border
	vector.StrokeRect(screen, boxX, boxY, boxWidth, boxHeight,
		3, color.RGBA{0, 255, 255, 255}, false)
	
	// Title
	titleFace := &text.GoTextFace{
		Source: face.Source,
		Size:   24,
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(boxX+boxWidth/2-100), float64(boxY+30))
	op.ColorScale.ScaleWithColor(color.RGBA{0, 255, 255, 255})
	text.Draw(screen, "CONTROLS", titleFace, op)
	
	// Help text
	helpText := []string{
		"Q - Quit",
		"F - Toggle Fullscreen",
		"F1 - Toggle Help",
		"",
		"E - Evolve Pattern",
		"M - Change Evolution Mode",
		"    Current: " + game.evolutionMode.String(),
		"",
		"+/- (Numpad) - Change Cell Size",
		"",
		"Evolution Modes:",
		"  • Movers - Gliders & Spaceships",
		"  • Oscillators - Blinkers & Pulsars",
		"  • Chaos - Maximum Activity",
		"  • Growth - Expanding Patterns",
	}
	
	smallFace := &text.GoTextFace{
		Source: face.Source,
		Size:   16,
	}
	
	for i, line := range helpText {
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(boxX+30), float64(boxY)+70+float64(i*22))
		op.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, line, smallFace, op)
	}
}

func getColorForAge(age uint8) color.Color {
	// Young cells (age 1-10): bright cyan to white
	// Medium cells (age 11-50): cyan to blue
	// Old cells (age 51+): blue to purple
	
	if age <= 10 {
		// Bright cyan to white (young cells)
		intensity := uint8(200 + (age * 5))
		return color.RGBA{R: intensity, G: 255, B: 255, A: 255}
	} else if age <= 50 {
		// Cyan to blue (maturing cells)
		fade := uint8((age - 10) * 6)
		return color.RGBA{R: 100 - fade, G: 200, B: 255, A: 255}
	} else {
		// Blue to purple (ancient cells)
		age_capped := age
		if age_capped > 150 {
			age_capped = 150
		}
		fade := uint8((age_capped - 50))
		return color.RGBA{R: 100 + fade, G: 50, B: 255, A: 255}
	}
}

func (game *Game) evolveInterestingPatternAsync() {
	// Use smaller resolution for fast evolution, then scale up
	searchWidth := 200
	searchHeight := 200
	
	evolver := NewPatternEvolver(searchWidth, searchHeight)
	
	// Seed with known interesting patterns (gliders, spaceships, etc.)
	evolver.SeedWithKnownPatterns()
	
	// Adjust density based on mode
	density := 0.05 + rand.Float64()*0.1 // Default for movers
	if game.evolutionMode == ModeChaos || game.evolutionMode == ModeGrowth {
		density = 0.15 + rand.Float64()*0.2 // Higher density for chaos/growth
	}
	
	// Generate additional random population
	populationSize := 24 // 6 known + 24 random = 30 total
	for i := 0; i < populationSize; i++ {
		evolver.GenerateRandomPatterns(1, density)
	}
	
	// Evolve for 12 generations
	for gen := 0; gen < 12; gen++ {
		game.evolvingMessage = fmt.Sprintf("🧬 Evolving %s... %d/12", game.evolutionMode.String(), gen+1)
		evolver.EvaluateFitness(50, game.evolutionMode) // Pass evolution mode
		
		if gen < 11 {
			// Higher mutation rate for more diversity
			evolver.Evolve(8, 0.08 + rand.Float64()*0.04) // Keep top 8, 8-12% mutation
		}
	}
	
	best := evolver.GetBestPattern()
	if best != nil {
		log.Printf("✨ Found interesting pattern! Fitness: %.1f", best.fitness)
		
		// Fill entire board with sparse random cells (20-30%)
		randomDensity := 0.20 + rand.Float64()*0.10
		for i := range game.board {
			if rand.Float64() < randomDensity {
				game.board[i] = true
				game.cellAge[i] = 1
			} else {
				game.board[i] = false
				game.cellAge[i] = 0
			}
		}
		
		// Place evolved pattern in center (overwrites random cells)
		startX := (game.gridWidth - searchWidth) / 2
		startY := (game.gridHeight - searchHeight) / 2
		
		for y := 0; y < searchHeight; y++ {
			for x := 0; x < searchWidth; x++ {
				boardX := startX + x
				boardY := startY + y
				if boardX >= 0 && boardX < game.gridWidth && boardY >= 0 && boardY < game.gridHeight {
					boardIdx := boardY*game.gridWidth + boardX
					if best.cells[y*searchWidth+x] {
						game.board[boardIdx] = true
						game.cellAge[boardIdx] = 1
					} else {
						game.board[boardIdx] = false
						game.cellAge[boardIdx] = 0
					}
				}
			}
		}
		
		game.updatePopulation()
		game.evolvingMessage = "✨ Pattern evolved! Press E for new one"
		log.Println("   Pattern loaded! Press E again to evolve a new one.")
	}
	
	// Clear evolving status after 2 seconds
	time.Sleep(2 * time.Second)
	game.isEvolving = false
	game.evolvingMessage = ""
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ebiten.ScreenSizeInFullscreen()
}
