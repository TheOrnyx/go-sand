package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"math/rand"
	"github.com/veandco/go-sdl2/sdl"
)

const WIN_WIDTH, WIN_HEIGHT = 900, 900
const GAME_WIDTH, GAME_HEIGHT = 100, 100
const WIDTH_SCALE, HEIGHT_SCALE = WIN_WIDTH/GAME_WIDTH, WIN_HEIGHT/GAME_HEIGHT

const FRAME_RATE = 400

const ( //draw methods
	DRAW_DIRT = iota 
	DRAW_AIR
	DRAW_WALL
	END_DRAW //used for like switching
)

const (
	AIR = iota
	DIRT
	WALL
	WATER
)

func setupWorld(w, h int32) [][]uint8 {
	world := make([][]uint8, h)
	for i := range world {
		world[i] = make([]uint8, w)
	}

	return world
}

func drawWorld(world *[][]uint8, renderer *sdl.Renderer) {
	for i_row := 0; i_row < len((*world)); i_row++ {
		for i_col := 0; i_col < len((*world)[0]); i_col++ {
			switch (*world)[i_row][i_col] {
			case DIRT:
				renderer.SetDrawColor(255, 255, 0, 255)
				renderer.DrawPoint(int32(i_row), int32(i_col))
			case WALL:
				renderer.SetDrawColor(130, 130, 130, 255)
				renderer.DrawPoint(int32(i_row), int32(i_col))
			}
		}
	}
}

func updateWorld(world *[][]uint8) {
	for i_row := GAME_HEIGHT-1; i_row >= 0; i_row-- {
		for i_col := GAME_WIDTH-2; i_col >= 0; i_col-- {
			switch (*world)[i_row][i_col] {
			case DIRT:
				rightX, rightY := i_row+1, i_col+1
				rightInBounds := rightX >= 0 && rightX <= GAME_WIDTH-1 && rightY > 0 && rightY <= GAME_HEIGHT-1
				leftX, leftY := i_row-1, i_col+1
				leftInBounds := leftX >= 0 && leftX <= GAME_WIDTH-1 && leftY > 0 && leftY <= GAME_HEIGHT-1
				canMoveRight := rightInBounds && (*world)[rightX][rightY] == AIR
				canMoveLeft := leftInBounds && (*world)[leftX][leftY] == AIR
				
				if i_col < GAME_HEIGHT-1 && (*world)[i_row][i_col+1] == AIR {
					(*world)[i_row][i_col+1] = DIRT
					(*world)[i_row][i_col] = AIR
				} else if canMoveLeft && canMoveRight {
					direction := rand.Intn(21)
					if direction <= 10{
						(*world)[rightX][rightY] = DIRT
						(*world)[i_row][i_col] = AIR
					} else {
						(*world)[leftX][leftY] = DIRT
						(*world)[i_row][i_col] = AIR
					}
				} else if canMoveRight {
					(*world)[rightX][rightY] = DIRT
					(*world)[i_row][i_col] = AIR
				}  else if canMoveLeft {
					(*world)[leftX][leftY] = DIRT
					(*world)[i_row][i_col] = AIR
				}
			}
		}
	}
}

func clampMouse (mouseX *int32, mouseY *int32) {
	*mouseX /= WIDTH_SCALE
	*mouseY /= HEIGHT_SCALE

	*mouseX = max(min(*mouseX, GAME_WIDTH-1), 0)
	*mouseY = max(min(*mouseY, GAME_HEIGHT-1), 0)
}

func run() (err error) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Fatal(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Sandy", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, WIN_WIDTH, WIN_HEIGHT, sdl.WINDOW_SHOWN | sdl.WINDOW_OPENGL)
	if err != nil {
		log.Fatal(err)
	}
	defer window.Destroy()
	window.SetWindowOpacity(1.0)

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		println("Creating renderer:", err)
		log.Fatal(err)
	}
	defer renderer.Destroy()
	renderer.SetScale(WIDTH_SCALE, HEIGHT_SCALE)
	
	running := true
	drawType := DIRT
	world := setupWorld(GAME_WIDTH, GAME_HEIGHT)
	// var brushSize int32 = 0
	var frameStart time.Time
	var elapsedTime float32
	dragging := false
	
	for running {
		frameStart = time.Now()
		renderer.SetDrawColor(0, 0, 0, 0)
		renderer.Clear()

		mouseX, mouseY, _ := sdl.GetMouseState()
		clampMouse(&mouseX, &mouseY)
		
		drawWorld(&world, renderer)
		updateWorld(&world)
		renderer.Present()
		
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
				
			case *sdl.MouseButtonEvent:
				if t.State == sdl.PRESSED {
					dragging = true
				} else if t.State == sdl.RELEASED {
					dragging = false
				}

			case *sdl.KeyboardEvent:
				if t.Type == sdl.KEYUP {
					key := t.Keysym
					//fmt.Printf("key: %v\n", key.Sym)
					if key.Sym == 114 {
						world = setupWorld(GAME_WIDTH, GAME_HEIGHT)
					} else {
						drawType = (drawType+1) % END_DRAW
						fmt.Println("Draw Type: ", drawType)
					}
				}
			}
		}
		
		if dragging {
			switch drawType {
			case DRAW_DIRT:
				world[mouseX][mouseY] = DIRT
			case DRAW_AIR:
				world[mouseX][mouseY] = AIR
			case DRAW_WALL:
				world[mouseX][mouseY] = WALL
			}
		}

		elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		delayTime := float32(1000/FRAME_RATE) - elapsedTime
		if delayTime > 0 {
			sdl.Delay(uint32(delayTime))
		}
	}
	return
}

func main() {
	fmt.Println("begin")
	if err := run(); err != nil {
		os.Exit(1)
	}
}
