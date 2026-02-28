package main

import (
	"math/rand"
	"sort"
)

// PatternEvolver uses genetic algorithms to evolve interesting patterns
type PatternEvolver struct {
	width      int
	height     int
	population []Pattern
	generation int
}

type Pattern struct {
	cells   []bool
	fitness float64
}

type FitnessMetrics struct {
	survival     int     // How many generations it survives
	stability    float64 // How stable is the population
	activity     float64 // How much activity (births/deaths)
	uniqueness   float64 // How unique compared to other patterns
	finalPop     int     // Final population count
	maxPop       int     // Peak population
	movement     float64 // How much the pattern moves (for spaceships/gliders)
}

func NewPatternEvolver(width, height int) *PatternEvolver {
	return &PatternEvolver{
		width:      width,
		height:     height,
		population: make([]Pattern, 0),
	}
}

// Known interesting patterns
func getKnownPatterns() map[string][][]bool {
	return map[string][][]bool{
		"glider": {
			{false, true, false},
			{false, false, true},
			{true, true, true},
		},
		"lightweight_spaceship": {
			{false, true, false, false, true},
			{true, false, false, false, false},
			{true, false, false, false, true},
			{true, true, true, true, false},
		},
		"middleweight_spaceship": {
			{false, false, true, false, false, false},
			{false, true, false, false, false, true},
			{true, false, false, false, false, false},
			{true, false, false, false, false, true},
			{true, true, true, true, true, false},
		},
		"pulsar": {
			{false, false, true, true, true, false, false, false, true, true, true, false, false},
			{false, false, false, false, false, false, false, false, false, false, false, false, false},
			{true, false, false, false, false, true, false, true, false, false, false, false, true},
			{true, false, false, false, false, true, false, true, false, false, false, false, true},
			{true, false, false, false, false, true, false, true, false, false, false, false, true},
			{false, false, true, true, true, false, false, false, true, true, true, false, false},
		},
		"blinker": {
			{true, true, true},
		},
		"toad": {
			{false, true, true, true},
			{true, true, true, false},
		},
	}
}

// SeedWithKnownPatterns adds known interesting patterns at random positions
func (pe *PatternEvolver) SeedWithKnownPatterns() {
	known := getKnownPatterns()
	
	for name, template := range known {
		pattern := Pattern{
			cells: make([]bool, pe.width*pe.height),
		}
		
		templateHeight := len(template)
		templateWidth := len(template[0])
		
		// Place pattern at random position
		startX := rand.Intn(pe.width - templateWidth)
		startY := rand.Intn(pe.height - templateHeight)
		
		for y := 0; y < templateHeight; y++ {
			for x := 0; x < templateWidth; x++ {
				if template[y][x] {
					boardX := startX + x
					boardY := startY + y
					idx := boardY*pe.width + boardX
					pattern.cells[idx] = true
				}
			}
		}
		
		pe.population = append(pe.population, pattern)
		_ = name // avoid unused warning
	}
}

// GenerateRandomPatterns creates initial random population
func (pe *PatternEvolver) GenerateRandomPatterns(count int, density float64) {
	size := pe.width * pe.height
	
	for i := 0; i < count; i++ {
		pattern := Pattern{
			cells: make([]bool, size),
		}
		
		// Create random pattern with given density
		for j := 0; j < size; j++ {
			if rand.Float64() < density {
				pattern.cells[j] = true
			}
		}
		
		pe.population = append(pe.population, pattern)
	}
}

// EvaluateFitness runs each pattern and scores it
func (pe *PatternEvolver) EvaluateFitness(maxGenerations int, mode EvolutionMode) {
	for i := range pe.population {
		pe.population[i].fitness = pe.calculateFitness(&pe.population[i], maxGenerations, mode)
	}
	
	// Sort by fitness (highest first)
	sort.Slice(pe.population, func(i, j int) bool {
		return pe.population[i].fitness > pe.population[j].fitness
	})
}

func (pe *PatternEvolver) calculateFitness(pattern *Pattern, maxGen int, mode EvolutionMode) float64 {
	// Simulate the pattern
	game := &Game{
		gridWidth:  pe.width,
		gridHeight: pe.height,
		board:      make([]bool, len(pattern.cells)),
		nextBoard:  make([]bool, len(pattern.cells)),
	}
	
	copy(game.board, pattern.cells)
	
	metrics := FitnessMetrics{}
	prevPop := 0
	totalActivity := 0.0
	
	// Track center of mass to detect movement
	var prevCenterX, prevCenterY float64
	totalMovement := 0.0
	
	for gen := 0; gen < maxGen; gen++ {
		// Count current population and calculate center of mass
		pop := 0
		centerX := 0.0
		centerY := 0.0
		
		for i, cell := range game.board {
			if cell {
				pop++
				x := i % pe.width
				y := i / pe.width
				centerX += float64(x)
				centerY += float64(y)
			}
		}
		
		if pop == 0 {
			// Pattern died
			metrics.survival = gen
			break
		}
		
		// Calculate average center of mass
		centerX /= float64(pop)
		centerY /= float64(pop)
		
		// Track movement (distance center of mass traveled)
		if gen > 0 {
			dx := centerX - prevCenterX
			dy := centerY - prevCenterY
			distance := (dx*dx + dy*dy) // squared distance (faster, no sqrt needed)
			totalMovement += distance
		}
		prevCenterX = centerX
		prevCenterY = centerY
		
		metrics.survival = gen + 1
		metrics.finalPop = pop
		if pop > metrics.maxPop {
			metrics.maxPop = pop
		}
		
		// Track activity
		if gen > 0 {
			activity := float64(abs(pop - prevPop))
			totalActivity += activity
		}
		prevPop = pop
		
		// Evolve one step
		for x := 0; x < pe.width; x++ {
			for y := 0; y < pe.height; y++ {
				idx := y*pe.width + x
				alive := game.board[idx]
				neighbors := game.countNeighbors(x, y)
				
				if alive && (neighbors == 2 || neighbors == 3) {
					game.nextBoard[idx] = true
				} else if !alive && neighbors == 3 {
					game.nextBoard[idx] = true
				} else {
					game.nextBoard[idx] = false
				}
			}
		}
		game.board, game.nextBoard = game.nextBoard, game.board
	}
	
	if metrics.survival > 0 {
		metrics.activity = totalActivity / float64(metrics.survival)
		metrics.movement = totalMovement / float64(metrics.survival)
	}
	
	// Calculate stability (lower variation = more stable)
	if metrics.maxPop > 0 {
		metrics.stability = 1.0 - (float64(metrics.maxPop-metrics.finalPop) / float64(metrics.maxPop))
	}
	
	// Different fitness functions based on evolution mode
	fitness := 0.0
	
	switch mode {
	case ModeMovers:
		// Optimized for gliders, spaceships
		fitness += float64(metrics.survival) * 10.0
		fitness += metrics.movement * 50.0 // HEAVILY reward movement
		fitness += metrics.activity * 1.0
		if metrics.finalPop > 0 && metrics.finalPop < 50 {
			fitness += 100.0 / float64(metrics.finalPop) // smaller = better
		}
		if metrics.survival == maxGen && metrics.movement > 10.0 {
			fitness += 1000.0
		}
		
	case ModeOscillators:
		// Optimized for blinkers, pulsars - stable population, low movement
		fitness += float64(metrics.survival) * 15.0
		fitness += metrics.stability * 500.0 // High stability
		fitness += metrics.activity * 10.0   // Some activity (changing cells)
		fitness -= metrics.movement * 20.0   // PENALIZE movement
		if metrics.finalPop > 10 && metrics.finalPop < 100 {
			fitness += float64(metrics.finalPop) * 2.0
		}
		if metrics.survival == maxGen {
			fitness += 800.0
		}
		
	case ModeChaos:
		// Optimized for maximum activity and chaos
		fitness += float64(metrics.survival) * 5.0
		fitness += metrics.activity * 50.0  // HEAVILY reward activity
		fitness += float64(metrics.maxPop) * 3.0
		fitness -= metrics.stability * 200.0 // PENALIZE stability
		if metrics.finalPop > 100 {
			fitness += float64(metrics.finalPop) * 5.0
		}
		
	case ModeGrowth:
		// Optimized for expanding patterns
		fitness += float64(metrics.survival) * 10.0
		fitness += float64(metrics.maxPop-metrics.finalPop) * 2.0 // Reward growth
		fitness += float64(metrics.finalPop) * 10.0 // Bigger = better
		if metrics.finalPop > metrics.maxPop/2 {
			fitness += 500.0 // Still growing at the end
		}
		if metrics.survival == maxGen {
			fitness += 1000.0
		}
	}
	
	return fitness
}

// Evolve creates next generation using selection and mutation
func (pe *PatternEvolver) Evolve(keepBest int, mutationRate float64) {
	pe.generation++
	
	// Keep the best patterns
	newPop := make([]Pattern, len(pe.population))
	for i := 0; i < keepBest && i < len(pe.population); i++ {
		newPop[i] = pe.population[i]
	}
	
	// Fill rest with mutations of best patterns
	for i := keepBest; i < len(newPop); i++ {
		// Select parent (weighted toward better fitness)
		parentIdx := rand.Intn(keepBest)
		parent := pe.population[parentIdx]
		
		// Create mutated child
		child := Pattern{
			cells: make([]bool, len(parent.cells)),
		}
		copy(child.cells, parent.cells)
		
		// Mutate
		for j := range child.cells {
			if rand.Float64() < mutationRate {
				child.cells[j] = !child.cells[j]
			}
		}
		
		newPop[i] = child
	}
	
	pe.population = newPop
}

// GetBestPattern returns the highest fitness pattern
func (pe *PatternEvolver) GetBestPattern() *Pattern {
	if len(pe.population) == 0 {
		return nil
	}
	return &pe.population[0]
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
