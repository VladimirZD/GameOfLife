package main

import (
	"image/color"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	_ "net/http/pprof"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

var (
	fulscreen bool = true
	cellSize  int  = 2
)

// func getWindowSize() (width int, height int) {
// 	if ebiten.IsFullscreen() {
// 		return ebiten.ScreenSizeInFullscreen()
// 	}
// 	return ebiten.WindowSize()
// }

func main() {

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	game := &Game{}
	ebiten.SetWindowTitle("Game of Life")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowDecorated(true)
	ebiten.SetFullscreen(fulscreen)

	width, height := ebiten.ScreenSizeInFullscreen()
	game.windowHeight = height
	game.windowWidth = width
	log.Printf("Window size: %dx%d", width, height)

	rand.Seed(time.Now().UnixNano())
	// board := make([][]bool, width/cellSize)
	// for i := range board {
	// 	board[i] = make([]bool, height/cellSize)
	// 	for j := range board[i] {
	// 		board[i][j] = rand.Intn(10) == 1
	// 	}
	// }
	// game.board = board
	initboard(game)

	game.gameStart = time.Now()
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}

func initboard(game *Game) {
	board := make([][]bool, game.windowWidth/cellSize)
	for i := range board {
		board[i] = make([]bool, game.windowHeight/cellSize)
		for j := range board[i] {
			board[i][j] = rand.Intn(10) == 1
		}
	}
	game.board = board
}

type Game struct {
	windowWidth  int
	windowHeight int
	board        [][]bool
	gameStart    time.Time
	cycleCount   int64
}

func (game *Game) Update() error {

	game.cycleCount++
	elapsed := time.Since(game.gameStart).Seconds()

	//log.Println(game.cycleCount, game.cycleCount/int64(elapsed))
	if game.cycleCount%100 == 0 {
		log.Println("Total cycles done:", game.cycleCount, " current rate", game.cycleCount/int64(elapsed), "cycles/s")
	}
	// Exit the game after 20 seconds
	if elapsed > 2000 {
		//log.Println("Time elapsed", game.gameStart, dt)
		os.Exit(0)
	}
	//game.elapsedTime += dt

	game.windowWidth, game.windowHeight = ebiten.WindowSize()
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		os.Exit(0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		fulscreen = !fulscreen
		ebiten.SetFullscreen(fulscreen)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyNumpadAdd) {
		cellSize++

	}
	if inpututil.IsKeyJustPressed(ebiten.KeyNumpadSubtract) {
		if cellSize > 1 {
			cellSize--
		}
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
