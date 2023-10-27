// TODO
// Add some kind of extra 2d array for already affected points
// Fix the water upwarp

package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"log"
	"math/rand"
	"os"
	"time"
)

const WIN_WIDTH, WIN_HEIGHT = 1000, 1000
const GAME_WIDTH, GAME_HEIGHT = 200, 200
const WIDTH_SCALE, HEIGHT_SCALE = WIN_WIDTH / GAME_WIDTH, WIN_HEIGHT / GAME_HEIGHT

const FRAME_RATE = 150

const (
	fontPath = "/usr/share/fonts/noto/NotoSans-Regular.ttf"
	fontSize = 20
)

const ( //draw methods
	DRAW_AIR = iota
	DRAW_DIRT
	DRAW_WALL
	DRAW_WATER
	DRAW_WOOD
	DRAW_FIRE
	DRAW_LAVA
	END_DRAW //used for like switching
)

var drawNameList = [...]string{"Air", "Dirt", "Wall", "Water", "Wood", "Fire", "Lava"}

const (
	AIR = iota
	DIRT
	WALL
	WATER
	WOOD
	FIRE
	LAVA
)

func setupWorld(w, h int32) [][]uint8 {
	world := make([][]uint8, h)
	for i := range world {
		world[i] = make([]uint8, w)
	}

	return world
}

// drawInfo ...
func drawInfo(font *ttf.Font, renderer *sdl.Renderer, drawType int) {
	var text *sdl.Surface
	var err error
	message := fmt.Sprintf("Current draw: %v", drawNameList[drawType])
	if text, err = font.RenderUTF8Blended(message, sdl.Color{R: 255, G: 255, B: 255, A: 255}); err != nil {
		return
	}
	defer text.Free()

	tex, err := renderer.CreateTextureFromSurface(text)
	if err != nil {
		log.Fatal(err)
	}
	defer tex.Destroy()

	_, _, width, height, _ := tex.Query()
	destRect := &sdl.Rect{X: 0, Y: 0, W: width / 4, H: height / 4}
	renderer.Copy(tex, nil, destRect)
}

// drawInitInfo draws the basic info about the program
// Doesn't work for some reason on a window manager (at least on xmonad)
// so if it doesn't work then it just prints to stdout
func drawInitInfo(window *sdl.Window) {
	message := "right/left = next/prev type\n-/= = make brush size bigger or smaller\np = toggle pause"
	if err := sdl.ShowSimpleMessageBox(sdl.MESSAGEBOX_INFORMATION, "welcome to sand", message, nil); err != nil {
		fmt.Println("MessageBox either doesn't work\nOr you're on a WM lmao\nprinting to Term\n")
		fmt.Println(message)
	}
}

func drawMouse(renderer *sdl.Renderer, mouseX, mouseY int32, brushSize uint8) {
	renderer.SetDrawColor(255, 0, 0, 100) // Set color to red

	// Calculate the position and dimensions of the rectangle
	startX := int32(mouseX) - int32(brushSize)/2
	startY := int32(mouseY) - int32(brushSize)/2
	width := int32(brushSize)
	height := int32(brushSize)

	// Draw the rectangle
	renderer.DrawRect(&sdl.Rect{X: startX, Y: startY, W: width, H: 1})              // Top
	renderer.DrawRect(&sdl.Rect{X: startX, Y: startY + height - 1, W: width, H: 1}) // Bottom
	renderer.DrawRect(&sdl.Rect{X: startX, Y: startY, W: 1, H: height})             // Left
	renderer.DrawRect(&sdl.Rect{X: startX + width - 1, Y: startY, W: 1, H: height}) // Right
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
				renderer.SetDrawColor(0, 0, 255, 150)
				renderer.DrawPoint(int32(i_col), int32(i_row))
			case WOOD:
				renderer.SetDrawColor(70, 60, 10, 255)
				renderer.DrawPoint(int32(i_col), int32(i_row))
			case FIRE:
				renderer.SetDrawColor(uint8(rand.Intn(30)+200), 10, 10, 255)
				renderer.DrawPoint(int32(i_col), int32(i_row))
			case LAVA:
				renderer.SetDrawColor(240, 170, 58, 255)
				renderer.DrawPoint(int32(i_col), int32(i_row))
			}
		}
	}
}

func updateWorldLava(world *[][]uint8, i_row, i_col int, compWorld *[][]uint8) {
	var moveChance, fireChance int32
	var spreadThresh, spreadFireThresh int32 = 10, 2
	rightX, rightY := i_col+1, i_row
	rightInBounds := rightX >= 0 && rightX <= GAME_WIDTH-1 && rightY > 0 && rightY <= GAME_HEIGHT-1
	leftX, leftY := i_col-1, i_row
	leftInBounds := leftX >= 0 && leftX <= GAME_WIDTH-1 && leftY > 0 && leftY <= GAME_HEIGHT-1
	canMoveRight := rightInBounds && (*world)[rightY][rightX] == AIR
	canMoveLeft := leftInBounds && (*world)[leftY][leftX] == AIR
	topX, topY := i_col, i_row-1
	canSpreadTop := (topY >= 0) && ((*world)[topY][topX] == AIR || (*world)[topY][topX] == WOOD)

	moveChance = rand.Int31n(40)
	fireChance = rand.Int31n(50)
	if spreadThresh > moveChance {
		if i_row < GAME_HEIGHT && ((*world)[i_row+1][i_col] == AIR || (*world)[i_row+1][i_col] == FIRE) {
			(*world)[i_row+1][i_col] = LAVA
			(*world)[i_row][i_col] = AIR
			(*compWorld)[i_row+1][i_col] = 1
		} else if canMoveLeft && canMoveRight {
			direction := rand.Intn(21)
			if direction < 7 { //random move direction, probably have it so it's weighted by free stuff
				(*world)[rightY][rightX] = LAVA
				(*world)[i_row][i_col] = AIR
				(*compWorld)[rightY][rightX] = 1
			} else {
				(*world)[leftY][leftX] = LAVA
				(*world)[i_row][i_col] = AIR
				(*compWorld)[leftY][leftX] = 1
			}
		} else if canMoveRight {
			(*world)[rightY][rightX] = LAVA
			(*world)[i_row][i_col] = AIR
			(*compWorld)[rightY][rightX] = 1
		} else if canMoveLeft {
			(*world)[leftY][leftX] = LAVA
			(*world)[i_row][i_col] = AIR
			(*compWorld)[leftY][leftX] = 1
		}
	}
	if canSpreadTop && spreadFireThresh > fireChance {
		(*world)[topY][topX] = FIRE
	}
}

func updateWorldFire(world *[][]uint8, i_row, i_col int) {
	var spreadChance int32
	var airSpreadChance, woodSpreadChance, dieChance int32 = 1, 6, 4
	rightX, rightY := i_col+1, i_row
	leftX, leftY := i_col-1, i_row
	rightInBounds, leftInBounds := rightX >= 0 && rightX < GAME_WIDTH, leftX >= 0 && leftX < GAME_WIDTH
	topX, topY, btmX, btmY := i_col, i_row-1, i_col, i_row+1
	topInBounds, btmInBounds := topY >= 0 && topY < GAME_HEIGHT, btmY >= 0 && btmY < GAME_HEIGHT
	canSpreadRight := rightInBounds && ((*world)[rightY][rightX] == AIR || (*world)[rightY][rightX] == WOOD)
	canSpreadLeft := leftInBounds && ((*world)[leftY][leftX] == AIR || (*world)[leftY][leftX] == WOOD)
	canSpreadTop := topInBounds && ((*world)[topY][topX] == AIR || (*world)[topY][topX] == WOOD)
	canSpreadBtm := btmInBounds && ((*world)[btmY][btmX] == AIR || (*world)[btmY][btmX] == WOOD)

	if canSpreadRight {
		rightCell := &(*world)[rightY][rightX]
		spreadChance = rand.Int31n(40)
		switch *rightCell {
		case AIR:
			if spreadChance < airSpreadChance {
				*rightCell = FIRE
			}
		case WOOD:
			if spreadChance < woodSpreadChance {
				*rightCell = FIRE
			}
		}
	}
	if canSpreadLeft {
		leftCell := &(*world)[leftY][leftX]
		spreadChance = rand.Int31n(40)
		switch *leftCell {
		case AIR:
			if spreadChance < airSpreadChance {
				*leftCell = FIRE
			}
		case WOOD:
			if spreadChance < woodSpreadChance {
				*leftCell = FIRE
			}
		}
	}
	if canSpreadTop {
		topCell := &(*world)[topY][topX]
		spreadChance = rand.Int31n(40)
		switch *topCell {
		case AIR:
			if spreadChance < airSpreadChance {
				*topCell = FIRE
			}
		case WOOD:
			if spreadChance < woodSpreadChance {
				*topCell = FIRE
			}
		}
	}
	if canSpreadBtm {
		btmCell := &(*world)[btmY][btmX]
		spreadChance = rand.Int31n(40)
		switch *btmCell {
		case AIR:
			if spreadChance < airSpreadChance {
				*btmCell = FIRE
			}
		case WOOD:
			if spreadChance < woodSpreadChance {
				*btmCell = FIRE
			}
		}
	}
	disappearChance := rand.Int31n(40)
	if (leftInBounds && (*world)[leftY][leftX] == WATER) || (rightInBounds && (*world)[rightY][rightX] == WATER) {
		dieChance += 20
	}
	if topInBounds && (*world)[topY][topX] == WATER {
		dieChance += 20
	}
	if disappearChance < dieChance {
		(*world)[i_row][i_col] = AIR
	}
}

func updateWorldDirt(world *[][]uint8, i_row, i_col int) {
	rightX, rightY := i_col+1, i_row+1
	rightInBounds := rightY >= 0 && rightY <= GAME_HEIGHT-1 && rightX > 0 && rightX <= GAME_WIDTH-1
	leftX, leftY := i_col-1, i_row+1
	leftInBounds := leftX >= 0 && leftX <= GAME_WIDTH-1 && leftY > 0 && leftY <= GAME_HEIGHT-1
	canMoveRight := rightInBounds && (((*world)[rightY][rightX] == AIR && (*world)[i_row][rightX] == AIR) || (*world)[rightY][rightX] == WATER)
	canMoveLeft := leftInBounds && (((*world)[leftY][leftX] == AIR && (*world)[i_row][leftX] == AIR) || (*world)[leftY][leftX] == WATER)
	canMoveBelow := i_row < GAME_HEIGHT-1 && ((*world)[i_row+1][i_col] == AIR || (*world)[i_row+1][i_col] == WATER)
	
	if canMoveBelow { //probably fix the water upwarp
		if (*world)[i_row+1][i_col] == WATER {
			(*world)[i_row+1][i_col] = DIRT
			(*world)[i_row][i_col] = WATER
		} else {
			(*world)[i_row+1][i_col] = DIRT
			(*world)[i_row][i_col] = AIR
		}
	} else if canMoveLeft && canMoveRight {
		rightPix := (*world)[rightY][rightX]
		leftPix := (*world)[leftY][leftX]
		direction := rand.Intn(21)
		if direction <= 10 {
			(*world)[i_row][i_col] = rightPix
			(*world)[rightY][rightX] = DIRT
		} else {
			(*world)[i_row][i_col] = leftPix
			(*world)[leftY][leftX] = DIRT
		}
	} else if canMoveRight {
		(*world)[i_row][i_col] = (*world)[rightY][rightX]
		(*world)[rightY][rightX] = DIRT
	} else if canMoveLeft {
		(*world)[i_row][i_col] = (*world)[leftY][leftX]
		(*world)[leftY][leftX] = DIRT
	}
}

func updateWorldWater(world *[][]uint8, i_row, i_col int, compWorld *[][]uint8) { //maybe unmark the like moved pos
	rightX, rightY := i_col+1, i_row
	rightInBounds := rightX >= 0 && rightX <= GAME_WIDTH-1 && rightY > 0 && rightY <= GAME_HEIGHT-1
	leftX, leftY := i_col-1, i_row
	leftInBounds := leftX >= 0 && leftX <= GAME_WIDTH-1 && leftY > 0 && leftY <= GAME_HEIGHT-1
	canMoveRight := rightInBounds && (*world)[rightY][rightX] == AIR
	canMoveLeft := leftInBounds && (*world)[leftY][leftX] == AIR

	if i_row < GAME_HEIGHT && (*world)[i_row+1][i_col] == AIR {
		(*world)[i_row+1][i_col] = WATER
		(*world)[i_row][i_col] = AIR
		(*compWorld)[i_row+1][i_col] = 1
	} else if canMoveLeft && canMoveRight {
		direction := rand.Intn(21)
		if direction < 7 { //random move direction, probably have it so it's weighted by free stuff
			(*world)[rightY][rightX] = WATER
			(*world)[i_row][i_col] = AIR
			(*compWorld)[rightY][rightX] = 1
		} else {
			(*world)[leftY][leftX] = WATER
			(*world)[i_row][i_col] = AIR
			(*compWorld)[leftY][leftX] = 1
		}
	} else if canMoveRight {
		(*world)[rightY][rightX] = WATER
		(*world)[i_row][i_col] = AIR
		(*compWorld)[rightY][rightX] = 1
	} else if canMoveLeft {
		(*world)[leftY][leftX] = WATER
		(*world)[i_row][i_col] = AIR
		(*compWorld)[leftY][leftX] = 1
	}
}

func debugPrintWorld(world *[][]uint8) {
	for i := range *world {
		fmt.Println((*world)[i])
	}
}

func updateWorld(world *[][]uint8) {
	completedWorld := setupWorld(GAME_WIDTH, GAME_HEIGHT)
	for i_row := GAME_HEIGHT - 2; i_row >= 0; i_row-- {
		for i_col := GAME_WIDTH - 1; i_col >= 0; i_col-- {
			if completedWorld[i_row][i_col] == 0 {
				switch (*world)[i_row][i_col] {
				case DIRT:
					updateWorldDirt(&*world, i_row, i_col)
				case WATER:
					updateWorldWater(&*world, i_row, i_col, &completedWorld)
				case FIRE:
					updateWorldFire(&*world, i_row, i_col)
				case LAVA:
					updateWorldLava(&*world, i_row, i_col, &completedWorld)
				}
			}
		}
	}
}

func clampMouse(mouseX *int32, mouseY *int32) {
	*mouseX /= WIDTH_SCALE
	*mouseY /= HEIGHT_SCALE
	*mouseX = max(min(*mouseX, GAME_WIDTH-1), 0)
	*mouseY = max(min(*mouseY, GAME_HEIGHT-1), 0)
}

func addCell(world *[][]uint8, mouseX, mouseY, drawType int, brushSize uint8) {
	startX, startY := int(mouseX)-int(brushSize)/2, int(mouseY)-int(brushSize)/2
	for i := 0; i < int(brushSize); i++ {
		for j := 0; j < int(brushSize); j++ {
			if startX+i >= 0 && startX+i < len((*world)[0]) && startY+j >= 0 && startY+j < len(*world) {
				(*world)[startY+j][startX+i] = uint8(drawType)
			}
		}
	}
}

func run() (err error) {
	var font *ttf.Font
	if err = ttf.Init(); err != nil {
		log.Fatal(err)
	}
	defer ttf.Quit()

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Fatal(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Sandy", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		WIN_WIDTH, WIN_HEIGHT, sdl.WINDOW_SHOWN|sdl.WINDOW_ALLOW_HIGHDPI|sdl.WINDOW_OPENGL)
	if err != nil {
		log.Fatal(err)
	}
	defer window.Destroy()
	//window.SetWindowOpacity(0.4)

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		println("Creating renderer:", err)
		log.Fatal(err)
	}
	defer renderer.Destroy()
	renderer.SetScale(WIDTH_SCALE, HEIGHT_SCALE)
	renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

	font, err = ttf.OpenFont(fontPath, fontSize)
	if err != nil {
		fmt.Println("Unable to load font, continuing anyway")
	}

	running, paused := true, false
	drawInitInfo(&*window)

	drawType := DRAW_DIRT
	var brushSize uint8 = 1
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
		if !paused {
			updateWorld(&world)
		}
		drawMouse(&*renderer, mouseX, mouseY, brushSize)
		drawInfo(font, renderer, drawType)
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
					} else if key.Sym == 112 {
						paused = !paused
						fmt.Println("Paused is: ", paused)
						//debugPrintWorld(&world)
					} else if key.Sym == 61 { // '='
						brushSize += 1
						fmt.Println("Brush Size: ", brushSize)
					} else if key.Sym == 45 { // '-'
						brushSize -= 1
						fmt.Println("Brush Size: ", brushSize)
					} else if key.Sym == 48 {
						brushSize = 1
						fmt.Println("Brush Size: ", brushSize)
					} else if key.Sym == 1073741903 {
						drawType = (drawType + 1) % END_DRAW
						fmt.Println("Draw Type: ", drawNameList[drawType])
					} else if key.Sym == 1073741904 {
						newDraw := drawType - 1
						if newDraw < 0 {
							newDraw = END_DRAW - 1
						}
						drawType = newDraw
						fmt.Println("Draw Type: ", drawNameList[drawType])
					} else {
						
					}
				}
			}
		}

		if dragging {
			addCell(&world, int(mouseX), int(mouseY), drawType, brushSize)
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
