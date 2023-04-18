package main

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

func draw(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 50, 50, 250, 250, color.White, false)
	vector.DrawFilledCircle(screen, 400, 400, 50, color.White, false)
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
	ebiten.SetFullscreen(true)
	width, height := ebiten.ScreenSizeInFullscreen()
	log.Printf("Window size: %dx%d", width, height)
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}

type Game struct {
	windowWidth  int
	windowHeight int
}

func (g *Game) Update() error {
	g.windowWidth, g.windowHeight = ebiten.WindowSize()

	// Print the window size
	// log.Printf("Window size: %dx%d", windowWidth, windowHeight)

	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		os.Exit(0)
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	//ebitenutil.DebugPrint(screen, fmt.Sprintf("Current resolution %dx%d", g.windowWidth, g.windowHeight))

	// Load the regular font
	tt, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal(err)
	}
	regularFont := truetype.NewFace(tt, &truetype.Options{
		Size:    64,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	windowWidth, windowHeight := getWindowSize()
	text.Draw(screen, fmt.Sprintf("Current resolution %dx%d", windowWidth, windowHeight), regularFont, 0, 100, color.White)
	draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	// log.Println(outsideWidth, "x", outsideHeight)
	return ebiten.ScreenSizeInFullscreen()
}
