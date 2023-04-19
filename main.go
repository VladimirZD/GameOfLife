package main

import (
	"image/color"
	"log"
	"math/rand"
	"os"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

var (
	fulscreen bool = true
	cellSize  int  = 3
)

func draw(screen *ebiten.Image) {
	// vector.DrawFilledRect(screen, 200, 200, float32(cellSize), float32(cellSize), color.White, false)
	// for i := range g.board {
	// 	for j := range g.board[i] {
	// 		if Game.board[i][j] {
	// 			vector.DrawFilledRect(screen, float64(i*cellSize), float64(j*cellSize), cellSize, cellSize, color.White)
	// 		}
	// 	}
	// }

	// vector.DrawFilledRect(screen, 200, 250, 10, 10, color.White, false)
	// vector.DrawFilledRect(screen, 200, 300, 30, 30, color.White, false)
	//vector.DrawFilledCircle(screen, 400, 400, 50, color.White, false)
}

func getWindowSize() (width int, height int) {
	if ebiten.IsFullscreen() {
		return ebiten.ScreenSizeInFullscreen()
	}
	return ebiten.WindowSize()
}

func main() {
	game := &Game{}
	ebiten.SetWindowTitle("Game of Life")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowDecorated(true)
	ebiten.SetFullscreen(fulscreen)
	width, height := ebiten.ScreenSizeInFullscreen()
	log.Printf("Window size: %dx%d", width, height)

	board := make([][]bool, width/cellSize)
	for i := range board {
		board[i] = make([]bool, height/cellSize)
		for j := range board[i] {
			board[i][j] = rand.Float32() < 0.5
		}
	}
	game.board = board

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}

type Game struct {
	windowWidth  int
	windowHeight int
	board        [][]bool
}

func (game *Game) Update() error {
	game.windowWidth, game.windowHeight = ebiten.WindowSize()
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		os.Exit(0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		fulscreen = !fulscreen
		ebiten.SetFullscreen(fulscreen)
	}
	newBoard := make([][]bool, len(game.board))
	for i := range game.board {
		newBoard[i] = make([]bool, len(game.board[i]))
		for j := range game.board[i] {
			alive := game.board[i][j]
			neighbors := game.countNeighbors(i, j)
			if alive && (neighbors == 2 || neighbors == 3) {
				newBoard[i][j] = true
			} else if !alive && neighbors == 3 {
				newBoard[i][j] = true
			} else {
				newBoard[i][j] = false
			}
		}
	}
	game.board = newBoard
	return nil
}

// func (game *Game) Update() error {
// 	game.windowWidth, game.windowHeight = ebiten.WindowSize()
// 	if ebiten.IsKeyPressed(ebiten.KeyQ) {
// 		os.Exit(0)
// 	}
// 	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
// 		fulscreen = !fulscreen
// 		ebiten.SetFullscreen(fulscreen)
// 	}

// 	rows, cols := len(game.board), len(game.board[0])
// 	for i := 0; i < rows; i++ {
// 		for j := 0; j < cols; j++ {
// 			alive := game.board[i][j]
// 			neighbors := game.countNeighbors(i, j)

// 			// Compute the new cell state using bitwise operations
// 			newState := (alive && (neighbors == 2 || neighbors == 3)) || (!alive && neighbors == 3)

// 			game.board[i][j] = newState
// 		}
// 	}

// 	return nil
// }

func (game *Game) countNeighbors(x, y int) int {
	count := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i == 0 && j == 0 {
				continue
			}
			ni := x + i
			nj := y + j
			if ni < 0 || ni >= len(game.board) || nj < 0 || nj >= len(game.board[0]) {
				continue
			}
			if game.board[ni][nj] {
				count++
			}
		}
	}
	return count
}

func (game *Game) Draw(screen *ebiten.Image) {
	//regularFont := GetFont()
	//windowWidth, windowHeight := getWindowSize()
	//text.Draw(screen, fmt.Sprintf("Current resolution %dx%d", windowWidth, windowHeight), regularFont, 0, 100, color.White)
	//draw(screen)
	//vector.DrawFilledRect(screen, 200, 200, float32(cellSize), float32(cellSize), color.White, false)
	for i := range game.board {
		for j := range game.board[i] {
			if game.board[i][j] {
				vector.DrawFilledRect(screen, float32(i*cellSize), float32(j*cellSize), float32(cellSize), float32(cellSize), color.White, false)
			}
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	// log.Println(outsideWidth, "x", outsideHeight)
	return ebiten.ScreenSizeInFullscreen()
}

func GetFont() font.Face {
	tt, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal(err)
	}
	font := truetype.NewFace(tt, &truetype.Options{
		Size:    32,
		DPI:     143,
		Hinting: font.HintingFull,
	})
	return font
}
