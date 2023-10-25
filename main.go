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

const FRAME_RATE = 300

const ( //draw methods
	DRAW_DIRT = iota 
	DRAW_AIR
	DRAW_WALL
	DRAW_WATER
	END_DRAW //used for like switching
)

var drawNameList = [...]string{"Dirt", "Air", "Wall", "Water"}

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
	for i_row := 0; i_row < GAME_HEIGHT; i_row++ {
		for i_col := 0; i_col < GAME_WIDTH; i_col++ {
			switch (*world)[i_row][i_col] {
			case DIRT:
				renderer.SetDrawColor(255, 255, 0, 255)
				renderer.DrawPoint(int32(i_col), int32(i_row))
			case WALL:
				renderer.SetDrawColor(130, 130, 130, 255)
				renderer.DrawPoint(int32(i_col), int32(i_row))
			case WATER:
				renderer.SetDrawColor(0, 0, 255, 10)
				renderer.DrawPoint(int32(i_col), int32(i_row))
			}
		}
	}
}

func updateWorldDirt(world *[][]uint8, i_row, i_col int) {
	rightX, rightY := i_col+1, i_row+1
	rightInBounds := rightY >= 0 && rightY <= GAME_HEIGHT-1 && rightX > 0 && rightX <= GAME_WIDTH-1
	leftX, leftY := i_col-1, i_row+1
	leftInBounds := leftX >= 0 && leftX <= GAME_WIDTH-1 && leftY > 0 && leftY <= GAME_HEIGHT-1
	canMoveRight := rightInBounds && (*world)[rightY][rightX] == AIR && (*world)[i_row][rightX] == AIR
	canMoveLeft := leftInBounds && (*world)[leftY][leftX] == AIR && (*world)[i_row][leftX] == AIR
	canMoveBelow := i_row < GAME_HEIGHT-1 && ((*world)[i_row+1][i_col] == AIR || (*world)[i_row+1][i_col] == WATER)
	
	if canMoveBelow {
		if (*world)[i_row+1][i_col] == WATER {
			(*world)[i_row+1][i_col] = DIRT
			(*world)[i_row][i_col] = WATER
		} else {
			(*world)[i_row+1][i_col] = DIRT
			(*world)[i_row][i_col] = AIR
		}
	} else if canMoveLeft && canMoveRight {
		direction := rand.Intn(21)
		if direction <= 10{
			(*world)[rightY][rightX] = DIRT
			(*world)[i_row][i_col] = AIR
		} else {
			(*world)[leftY][leftX] = DIRT
			(*world)[i_row][i_col] = AIR
		}
	} else if canMoveRight {
		(*world)[rightY][rightX] = DIRT
		(*world)[i_row][i_col] = AIR
	}  else if canMoveLeft {
		(*world)[leftY][leftX] = DIRT
		(*world)[i_row][i_col] = AIR
	}
}

func updateWorldWater(world *[][]uint8, i_row, i_col int){
	rightX, rightY := i_col+1, i_row
	rightInBounds := rightX >= 0 && rightX <= GAME_WIDTH-1 && rightY > 0 && rightY <= GAME_HEIGHT-1
	leftX, leftY := i_col-1, i_row
	leftInBounds := leftX >= 0 && leftX <= GAME_WIDTH-1 && leftY > 0 && leftY <= GAME_HEIGHT-1
	canMoveRight := rightInBounds && (*world)[rightY][rightX] == AIR
	canMoveLeft := leftInBounds && (*world)[leftY][leftX] == AIR 

	if i_row < GAME_HEIGHT && (*world)[i_row+1][i_col] == AIR {
		(*world)[i_row+1][i_col] = WATER
		(*world)[i_row][i_col] = AIR

	} else if canMoveLeft && canMoveRight {
		direction := rand.Intn(21) 
		if direction <= 5 { //random move direction, probably have it so it's weighted by free stuff
			(*world)[rightY][rightX] = WATER
			(*world)[i_row][i_col] = AIR
		} else {
			(*world)[leftY][leftX] = WATER
			(*world)[i_row][i_col] = AIR
		}
	} else if canMoveRight {
		(*world)[rightY][rightX] = WATER
		(*world)[i_row][i_col] = AIR
	} else if canMoveLeft {
		(*world)[leftY][leftX] = WATER
		(*world)[i_row][i_col] = AIR
	}
}

func debugPrintWorld(world *[][]uint8) {
	for i := range *world {
		fmt.Println((*world)[i])
	}

}

func updateWorld(world *[][]uint8) {
	for i_row := GAME_HEIGHT-2; i_row >= 0; i_row-- {
		for i_col := GAME_WIDTH-1; i_col >= 0; i_col-- {
			switch (*world)[i_row][i_col] {
			case DIRT:
				updateWorldDirt(&*world, i_row, i_col)
			case WATER:
				updateWorldWater(&*world, i_row, i_col)
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
	running, paused := true, false
	
	drawType := DRAW_DIRT
	world := setupWorld(GAME_WIDTH, GAME_HEIGHT)
	frameDuration := time.Second / FRAME_RATE
	dragging := false
	for running {
		startTime := time.Now()
		renderer.SetDrawColor(0, 0, 0, 0)
		renderer.Clear()
		mouseX, mouseY, _ := sdl.GetMouseState()
		clampMouse(&mouseX, &mouseY)
		drawWorld(&world, &*renderer)
		if !paused {updateWorld(&world)}
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
					// fmt.Printf("key: %v\n", key.Sym)
					if key.Sym == 114 {
						world = setupWorld(GAME_WIDTH, GAME_HEIGHT)
					} else if key.Sym == 112 {
						paused = !paused
						fmt.Println("Paused is: ", paused)
						//debugPrintWorld(&world)
					} else {
						drawType = (drawType+1) % END_DRAW
						fmt.Println("Draw Type: ", drawNameList[drawType])
					}
				}
			}
		}
		
		if dragging {
			switch drawType {
			case DRAW_DIRT:
				world[mouseY][mouseX] = DIRT
			case DRAW_AIR:
				world[mouseY][mouseX] = AIR
			case DRAW_WALL:
				world[mouseY][mouseX] = WALL
			case DRAW_WATER:
				world[mouseY][mouseX] = WATER
			}
		}
		elapsedTime := time.Since(startTime)
		if elapsedTime < frameDuration {
			remainingTime := frameDuration - elapsedTime
			time.Sleep(remainingTime)
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
