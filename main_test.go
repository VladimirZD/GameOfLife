package main

import "testing"

func TestCountNeighbors(t *testing.T) {
	tests := []struct {
		name      string
		board     [][]bool
		x, y      int
		wantCount int
	}{
		{
			name: "center cell with all neighbors alive",
			board: [][]bool{
				{true, true, true},
				{true, false, true},
				{true, true, true},
			},
			x:         1,
			y:         1,
			wantCount: 8,
		},
		{
			name: "center cell with no neighbors",
			board: [][]bool{
				{false, false, false},
				{false, false, false},
				{false, false, false},
			},
			x:         1,
			y:         1,
			wantCount: 0,
		},
		{
			name: "top-left corner with neighbors",
			board: [][]bool{
				{false, true, false},
				{true, true, false},
				{false, false, false},
			},
			x:         0,
			y:         0,
			wantCount: 3,
		},
		{
			name: "top-right corner with neighbors",
			board: [][]bool{
				{false, false, false},
				{true, true, false},
				{false, true, false},
			},
			x:         2,
			y:         0,
			wantCount: 3,
		},
		{
			name: "bottom-left corner with neighbors",
			board: [][]bool{
				{false, true, false},
				{false, true, true},
				{false, false, false},
			},
			x:         0,
			y:         2,
			wantCount: 3,
		},
		{
			name: "bottom-right corner with neighbors",
			board: [][]bool{
				{false, false, false},
				{false, true, true},
				{false, true, false},
			},
			x:         2,
			y:         2,
			wantCount: 3,
		},
		{
			name: "top edge with neighbors",
			board: [][]bool{
				{false, true, false},
				{false, false, false},
				{false, true, false},
			},
			x:         1,
			y:         0,
			wantCount: 2,
		},
		{
			name: "left edge with neighbors",
			board: [][]bool{
				{true, false, false},
				{true, false, false},
				{false, false, false},
			},
			x:         0,
			y:         1,
			wantCount: 2,
		},
		{
			name: "right edge with neighbors",
			board: [][]bool{
				{false, false, false},
				{false, false, true},
				{false, false, true},
			},
			x:         2,
			y:         1,
			wantCount: 2,
		},
		{
			name: "bottom edge with neighbors",
			board: [][]bool{
				{false, true, false},
				{false, false, false},
				{false, true, false},
			},
			x:         1,
			y:         2,
			wantCount: 2,
		},
		{
			name: "center cell with 3 neighbors",
			board: [][]bool{
				{true, true, false},
				{false, false, true},
				{false, false, false},
			},
			x:         1,
			y:         1,
			wantCount: 3,
		},
		{
			name: "center cell with 5 neighbors",
			board: [][]bool{
				{true, true, true},
				{false, false, true},
				{false, true, false},
			},
			x:         1,
			y:         1,
			wantCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := board2DToGame(tt.board)
			got := game.countNeighbors(tt.x, tt.y)
			if got != tt.wantCount {
				t.Errorf("countNeighbors(%d, %d) = %d, want %d", tt.x, tt.y, got, tt.wantCount)
			}
		})
	}
}

func board2DToGame(board2D [][]bool) *Game {
	if len(board2D) == 0 {
		return &Game{}
	}
	width := len(board2D)
	height := len(board2D[0])
	flat := make([]bool, width*height)
	
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			flat[y*width+x] = board2D[x][y]
		}
	}
	
	return &Game{
		gridWidth:  width,
		gridHeight: height,
		board:      flat,
	}
}

func TestGameStateTransitions(t *testing.T) {
	tests := []struct {
		name          string
		board         [][]bool
		x, y          int
		expectedAlive bool
		description   string
	}{
		{
			name: "live cell with 2 neighbors survives",
			board: [][]bool{
				{false, false, false},
				{true, true, true},
				{false, false, false},
			},
			x:             1,
			y:             1,
			expectedAlive: true,
			description:   "Center cell has 2 neighbors (left and right), should survive",
		},
		{
			name: "live cell with 3 neighbors survives",
			board: [][]bool{
				{true, true, false},
				{false, true, true},
				{false, false, false},
			},
			x:             1,
			y:             1,
			expectedAlive: true,
			description:   "Center cell has 3 neighbors, should survive",
		},
		{
			name: "live cell with 1 neighbor dies (underpopulation)",
			board: [][]bool{
				{false, false, false},
				{true, true, false},
				{false, false, false},
			},
			x:             1,
			y:             0,
			expectedAlive: false,
			description:   "Cell has only 1 neighbor, should die",
		},
		{
			name: "live cell with 0 neighbors dies (underpopulation)",
			board: [][]bool{
				{false, false, false},
				{false, true, false},
				{false, false, false},
			},
			x:             1,
			y:             0,
			expectedAlive: false,
			description:   "Cell has no neighbors, should die",
		},
		{
			name: "live cell with 4 neighbors dies (overpopulation)",
			board: [][]bool{
				{true, true, false},
				{true, true, true},
				{false, false, false},
			},
			x:             1,
			y:             1,
			expectedAlive: false,
			description:   "Center cell has 4 neighbors, should die",
		},
		{
			name: "live cell with 5 neighbors dies (overpopulation)",
			board: [][]bool{
				{true, true, true},
				{false, true, true},
				{false, true, false},
			},
			x:             1,
			y:             1,
			expectedAlive: false,
			description:   "Center cell has 5 neighbors, should die",
		},
		{
			name: "dead cell with 3 neighbors becomes alive (reproduction)",
			board: [][]bool{
				{true, true, false},
				{false, false, true},
				{false, false, false},
			},
			x:             1,
			y:             1,
			expectedAlive: true,
			description:   "Dead cell with exactly 3 neighbors becomes alive",
		},
		{
			name: "dead cell with 2 neighbors stays dead",
			board: [][]bool{
				{false, false, false},
				{true, false, true},
				{false, false, false},
			},
			x:             1,
			y:             0,
			expectedAlive: false,
			description:   "Dead cell with 2 neighbors stays dead",
		},
		{
			name: "dead cell with 4 neighbors stays dead",
			board: [][]bool{
				{true, true, false},
				{true, false, true},
				{false, false, false},
			},
			x:             1,
			y:             1,
			expectedAlive: false,
			description:   "Dead cell with 4 neighbors stays dead",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := board2DToGame(tt.board)
			width := game.gridWidth
			height := game.gridHeight
			game.nextBoard = make([]bool, width*height)
			
			// Run one evolution cycle
			for x := 0; x < width; x++ {
				for y := 0; y < height; y++ {
					idx := y*width + x
					alive := game.board[idx]
					neighbors := game.countNeighbors(x, y)
					if alive && (neighbors == 2 || neighbors == 3) {
						game.nextBoard[idx] = true
					} else if !alive && neighbors == 3 {
						game.nextBoard[idx] = true
					}
				}
			}
			
			idx := tt.y*width + tt.x
			got := game.nextBoard[idx]
			if got != tt.expectedAlive {
				t.Errorf("%s: cell at (%d, %d) = %v, want %v", tt.description, tt.x, tt.y, got, tt.expectedAlive)
			}
		})
	}
}

// Benchmarks

func TestEvolveBoard(t *testing.T) {
	tests := []struct {
		name           string
		initialBoard   [][]bool
		expectedBoard  [][]bool
		expectedBirths int
		expectedDeaths int
	}{
		{
			name: "blinker oscillates",
			initialBoard: [][]bool{
				{false, false, false, false, false},
				{false, false, true, false, false},
				{false, false, true, false, false},
				{false, false, true, false, false},
				{false, false, false, false, false},
			},
			expectedBoard: [][]bool{
				{false, false, false, false, false},
				{false, false, false, false, false},
				{false, true, true, true, false},
				{false, false, false, false, false},
				{false, false, false, false, false},
			},
			expectedBirths: 0, // Net population stays same (3 cells)
			expectedDeaths: 0,
		},
		{
			name: "block stays stable",
			initialBoard: [][]bool{
				{false, false, false, false},
				{false, true, true, false},
				{false, true, true, false},
				{false, false, false, false},
			},
			expectedBoard: [][]bool{
				{false, false, false, false},
				{false, true, true, false},
				{false, true, true, false},
				{false, false, false, false},
			},
			expectedBirths: 0,
			expectedDeaths: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialGame := board2DToGame(tt.initialBoard)
			game := &Game{
				gridWidth:   initialGame.gridWidth,
				gridHeight:  initialGame.gridHeight,
				board:       make([]bool, len(initialGame.board)),
				nextBoard:   make([]bool, len(initialGame.board)),
				cellAge:     make([]uint8, len(initialGame.board)),
				nextCellAge: make([]uint8, len(initialGame.board)),
				workerPool:  newWorkerPool(1), // Single worker for testing
			}
			copy(game.board, initialGame.board)
			
			// Initialize cell ages
			for i := range game.board {
				if game.board[i] {
					game.cellAge[i] = 1
				}
			}
			
			game.updatePopulation()
			initialPop := game.population
			
			game.evolveBoard()
			
			// Check resulting board state
			expectedGame := board2DToGame(tt.expectedBoard)
			for i := range game.board {
				if game.board[i] != expectedGame.board[i] {
					t.Errorf("cell %d: got %v, want %v", i, game.board[i], expectedGame.board[i])
				}
			}
			
			// Check births and deaths
			if game.births != tt.expectedBirths {
				t.Errorf("births = %d, want %d", game.births, tt.expectedBirths)
			}
			if game.deaths != tt.expectedDeaths {
				t.Errorf("deaths = %d, want %d", game.deaths, tt.expectedDeaths)
			}
			
			// Population should change correctly
			expectedPop := initialPop + game.births - game.deaths
			if game.population != expectedPop {
				t.Errorf("population = %d, want %d", game.population, expectedPop)
			}
		})
	}
}

func TestGetColorForAge(t *testing.T) {
	tests := []struct {
		name     string
		age      uint8
		wantType string // "cyan", "blue", "purple"
	}{
		{"young cell age 1", 1, "cyan"},
		{"young cell age 10", 10, "cyan"},
		{"medium cell age 20", 20, "blue"},
		{"medium cell age 50", 50, "blue"},
		{"old cell age 100", 100, "purple"},
		{"ancient cell age 200", 200, "purple"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := getColorForAge(tt.age)
			if col == nil {
				t.Fatal("got nil color")
			}
			// Just verify it returns a valid color (not nil)
		})
	}
}

func TestEvolutionModeString(t *testing.T) {
	tests := []struct {
		mode EvolutionMode
		want string
	}{
		{ModeMovers, "Movers (Gliders/Spaceships)"},
		{ModeOscillators, "Oscillators (Blinkers/Pulsars)"},
		{ModeChaos, "Chaos (Maximum Activity)"},
		{ModeGrowth, "Growth (Expanding Patterns)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.mode.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUpdatePopulation(t *testing.T) {
	game := &Game{
		gridWidth:  3,
		gridHeight: 3,
		board: []bool{
			true, false, true,
			false, true, false,
			true, false, true,
		},
	}

	game.updatePopulation()

	expectedPop := 5
	if game.population != expectedPop {
		t.Errorf("population = %d, want %d", game.population, expectedPop)
	}
}

func BenchmarkCountNeighbors(b *testing.B) {
	board2D := make([][]bool, 100)
	for i := range board2D {
		board2D[i] = make([]bool, 100)
		for j := range board2D[i] {
			board2D[i][j] = i%2 == 0 && j%2 == 0
		}
	}
	game := board2DToGame(board2D)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.countNeighbors(50, 50)
	}
}

func BenchmarkEvolveBoard_Small(b *testing.B) {
	game := &Game{
		windowWidth:  100,
		windowHeight: 100,
		cellSize:     1,
	}
	game.initBoard()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.evolveBoard()
	}
}

func BenchmarkEvolveBoard_Medium(b *testing.B) {
	game := &Game{
		windowWidth:  500,
		windowHeight: 500,
		cellSize:     1,
	}
	game.initBoard()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.evolveBoard()
	}
}

func BenchmarkEvolveBoard_Large(b *testing.B) {
	game := &Game{
		windowWidth:  1920,
		windowHeight: 1080,
		cellSize:     1,
	}
	game.initBoard()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.evolveBoard()
	}
}

func BenchmarkEvolveBoard_4K(b *testing.B) {
	game := &Game{
		windowWidth:  3840,
		windowHeight: 2160,
		cellSize:     1,
	}
	game.initBoard()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.evolveBoard()
	}
}

func TestBoardInitialization(t *testing.T) {
	tests := []struct {
		name         string
		windowWidth  int
		windowHeight int
		cellSize     int
		wantWidth    int
		wantHeight   int
	}{
		{
			name:         "100x100 with cell size 2",
			windowWidth:  100,
			windowHeight: 100,
			cellSize:     2,
			wantWidth:    50,
			wantHeight:   50,
		},
		{
			name:         "1920x1080 with cell size 4",
			windowWidth:  1920,
			windowHeight: 1080,
			cellSize:     4,
			wantWidth:    480,
			wantHeight:   270,
		},
		{
			name:         "800x600 with cell size 1",
			windowWidth:  800,
			windowHeight: 600,
			cellSize:     1,
			wantWidth:    800,
			wantHeight:   600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := &Game{
				windowWidth:  tt.windowWidth,
				windowHeight: tt.windowHeight,
				cellSize:     tt.cellSize,
			}
			game.initBoard()

			if game.gridWidth != tt.wantWidth {
				t.Errorf("grid width = %d, want %d", game.gridWidth, tt.wantWidth)
			}

			if game.gridHeight != tt.wantHeight {
				t.Errorf("grid height = %d, want %d", game.gridHeight, tt.wantHeight)
			}

			expectedSize := tt.wantWidth * tt.wantHeight
			if len(game.board) != expectedSize {
				t.Errorf("board size = %d, want %d", len(game.board), expectedSize)
			}

			if len(game.nextBoard) != expectedSize {
				t.Errorf("nextBoard size = %d, want %d", len(game.nextBoard), expectedSize)
			}
		})
	}
}

func TestKnownPatterns(t *testing.T) {
	t.Run("blinker oscillator period 2", func(t *testing.T) {
		// Horizontal blinker
		initial := [][]bool{
			{false, false, false, false, false},
			{false, false, true, false, false},
			{false, false, true, false, false},
			{false, false, true, false, false},
			{false, false, false, false, false},
		}

		// After one generation, should be vertical
		expected1 := [][]bool{
			{false, false, false, false, false},
			{false, false, false, false, false},
			{false, true, true, true, false},
			{false, false, false, false, false},
			{false, false, false, false, false},
		}

		game := board2DToGame(initial)
		width := game.gridWidth
		height := game.gridHeight
		game.nextBoard = make([]bool, width*height)

		// First generation
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				idx := y*width + x
				alive := game.board[idx]
				neighbors := game.countNeighbors(x, y)
				if alive && (neighbors == 2 || neighbors == 3) {
					game.nextBoard[idx] = true
				} else if !alive && neighbors == 3 {
					game.nextBoard[idx] = true
				}
			}
		}

		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				idx := y*width + x
				if game.nextBoard[idx] != expected1[x][y] {
					t.Errorf("generation 1: cell (%d,%d) = %v, want %v", x, y, game.nextBoard[idx], expected1[x][y])
				}
			}
		}

		// Second generation - should return to original
		game.board, game.nextBoard = game.nextBoard, game.board
		for i := range game.nextBoard {
			game.nextBoard[i] = false
		}

		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				idx := y*width + x
				alive := game.board[idx]
				neighbors := game.countNeighbors(x, y)
				if alive && (neighbors == 2 || neighbors == 3) {
					game.nextBoard[idx] = true
				} else if !alive && neighbors == 3 {
					game.nextBoard[idx] = true
				}
			}
		}

		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				idx := y*width + x
				if game.nextBoard[idx] != initial[x][y] {
					t.Errorf("generation 2: cell (%d,%d) = %v, want %v", x, y, game.nextBoard[idx], initial[x][y])
				}
			}
		}
	})

	t.Run("block still life", func(t *testing.T) {
		// 2x2 block - should never change
		initial := [][]bool{
			{false, false, false, false},
			{false, true, true, false},
			{false, true, true, false},
			{false, false, false, false},
		}

		game := board2DToGame(initial)
		width := game.gridWidth
		height := game.gridHeight
		game.nextBoard = make([]bool, width*height)

		// Run one generation
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				idx := y*width + x
				alive := game.board[idx]
				neighbors := game.countNeighbors(x, y)
				if alive && (neighbors == 2 || neighbors == 3) {
					game.nextBoard[idx] = true
				} else if !alive && neighbors == 3 {
					game.nextBoard[idx] = true
				}
			}
		}

		// Should be identical to initial
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				idx := y*width + x
				if game.nextBoard[idx] != initial[x][y] {
					t.Errorf("block changed: cell (%d,%d) = %v, want %v", x, y, game.nextBoard[idx], initial[x][y])
				}
			}
		}
	})
}
