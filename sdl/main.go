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

const (
	snakeHW         = 10
	widowDimensions = 600
)

const (
	RuneDArrow = '↓'
	RuneLArrow = '←'
	RuneRArrow = '→'
	RuneUArrow = '↑'
)

const (
	UP    = RuneUArrow
	DOWN  = RuneDArrow
	LEFT  = RuneLArrow
	RIGHT = RuneRArrow
)

var (
	Score        int
	FoodPlaced   = false
	FoodLocation sdl.Rect
	MOVEMENT     = map[rune]segment{UP: {0, -1 * snakeHW}, DOWN: {0, 1 * snakeHW}, LEFT: {-1 * snakeHW, 0}, RIGHT: {1 * snakeHW, 0}}
	heading      = map[rune]map[rune]segment{
		UP: {
			RIGHT: MOVEMENT[RIGHT],
			LEFT:  MOVEMENT[LEFT],
		},
		DOWN: {
			RIGHT: MOVEMENT[RIGHT],
			LEFT:  MOVEMENT[LEFT],
		},
		LEFT: {
			DOWN: MOVEMENT[DOWN],
			UP:   MOVEMENT[UP],
		},
		RIGHT: {
			DOWN: MOVEMENT[DOWN],
			UP:   MOVEMENT[UP],
		},
	}

	mapping = map[sdl.Scancode]rune{sdl.SCANCODE_UP: UP, sdl.SCANCODE_DOWN: DOWN, sdl.SCANCODE_LEFT: LEFT, sdl.SCANCODE_RIGHT: RIGHT}

	RestrictedLower int32
	RestrictedUpper int32
)

var (
	snakeBoi snake
	Places   []int32
)

type snake struct {
	bodySegmentPositions []sdl.Rect
	userDirection        rune
	heading              rune
}

type segment struct {
	x int32
	y int32
}

func newSnake() snake {
	return snake{
		bodySegmentPositions: []sdl.Rect{newBodySegment(330, 250), newBodySegment(320, 250), newBodySegment(310, 250)}, //initial position for snakeBoi
		userDirection:        0,
		heading:              RIGHT,
	}
}

func newBodySegment(x, y int32) sdl.Rect {
	return sdl.Rect{W: snakeHW, H: snakeHW, X: x, Y: y}
}

func (s *snake) move() error {
	if s.userDirection != 0 {
		fmt.Print("")
	}
	head := newBodySegment(0, 0)
	head.X, head.Y = s.calculatePosition(s.bodySegmentPositions[0].X, s.bodySegmentPositions[0].Y)

	collision := checkWallCollision(head)
	if collision {
		sdl.Delay(2000)
		return fmt.Errorf("Hit a wall, you ded 🐍")
	}

	bodyCollision := s.overlapsWithSegment(head.X, head.Y)
	if bodyCollision {
		sdl.Delay(2000)
		return fmt.Errorf("Hit yourself, you ded 🐍")
	}

	s.bodySegmentPositions = append([]sdl.Rect{head}, s.bodySegmentPositions...)

	foodFound := checkFoodCollision(head)
	if foodFound {
		Score += 5
		placeFood()
	}
	//IF there is food present at the new head location, delete the food item and keep tail
	if !foodFound {
		s.bodySegmentPositions = s.bodySegmentPositions[:len(s.bodySegmentPositions)-1]
	}

	//set new heading
	if s.userDirection != 0 {
		_, ok := heading[s.heading][s.userDirection]
		if ok {
			s.heading = s.userDirection
		}
	}
	//reset user direction
	s.userDirection = 0

	return nil
}

func checkFoodCollision(head sdl.Rect) bool {
	if head.X == FoodLocation.X && head.Y == FoodLocation.Y {
		fmt.Println("Food collision at", "Head:", head, "Food:", FoodLocation)
		return true
	}
	return false
}

func (s *snake) calculatePosition(x, y int32) (int32, int32) {
	n, ok := heading[s.heading][s.userDirection]

	if !ok {
		n = MOVEMENT[s.heading]
	}

	return x + n.x, y + n.y
}

func (s snake) overlapsWithSegment(X, Y int32) bool {
	for _, segment := range s.bodySegmentPositions {
		if Y == segment.Y && X == segment.X {
			return true
		}
	}
	return false
}

func checkWallCollision(head sdl.Rect) bool {
	if (head.X <= RestrictedLower || head.X >= RestrictedUpper) || (head.Y <= RestrictedLower || head.Y >= RestrictedUpper) {
		fmt.Printf("X:%d -- Y:%d, Lower:%d Upper: %d\n", head.X, head.Y, RestrictedLower, RestrictedUpper)
		return true
	}
	return false
}

func drawBoarder(r *sdl.Renderer) {
	length := int32(widowDimensions)
	thickness := int32(snakeHW / 2)
	topBoarder := sdl.Rect{X: 0, Y: 0, W: length, H: thickness}
	leftBoarder := sdl.Rect{X: 0, Y: 0, W: thickness, H: length}
	bottomBoarder := sdl.Rect{X: 0, Y: length - thickness, W: length, H: thickness}
	rightBoarder := sdl.Rect{X: length - thickness, Y: 0, W: thickness, H: length}
	r.SetDrawColor(0xff, 0x66, 0x00, 0xff)
	r.FillRects([]sdl.Rect{topBoarder, leftBoarder, rightBoarder, bottomBoarder})
	RestrictedLower = thickness
	RestrictedUpper = length - snakeHW
}

func placeFood() {
	//Try to place food, if random location occupied by snake keep trying
	for {
		randX := int32(Places[rand.Intn(len(Places))])
		randY := int32(Places[rand.Intn(len(Places))])

		if !snakeBoi.overlapsWithSegment(randX, randY) {
			FoodLocation = newBodySegment(randX, randY)
			break
		}

	}

}

func drawFood(r *sdl.Renderer) {
	r.SetDrawColor(0x00, 0xFF, 0x00, 0xff)
	r.FillRect(&FoodLocation)
}

func drawDeadScreen(r *sdl.Renderer) {
	ttf.Init()
	r.SetDrawColor(0, 90, 0, 255)
	r.Clear()
	font, err := ttf.OpenFont("font.ttf", 30)
	if err != nil {
		log.Fatalln("font", err)
	}
	scoreText, err := font.RenderUTF8Solid(fmt.Sprint("Score:", int(Score)), sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err != nil {
		log.Fatalln("scoreText")
	}

	gameOverText, err := font.RenderUTF8Solid("~Game Over~", sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err != nil {
		log.Fatalln("text")
	}

	scoreTexture, err := r.CreateTextureFromSurface(scoreText)
	if err != nil {
		log.Fatalln("texture")
	}

	GameOvertexture, err := r.CreateTextureFromSurface(gameOverText)
	if err != nil {
		log.Fatalln("texture")
	}
	gameOverTextPosition := sdl.Rect{X: 100, Y: 250, W: 400, H: 100}
	scoreTextPosition := sdl.Rect{X: 100, Y: 350, W: 400, H: 100}
	r.Copy(GameOvertexture, nil, &gameOverTextPosition)
	r.Copy(scoreTexture, nil, &scoreTextPosition)
	r.Present()
	sdl.Delay(3000)
	os.Exit(0)
}

func runGame(r *sdl.Renderer) {
	if !FoodPlaced {
		placeFood()
		FoodPlaced = true
	}
	r.SetDrawColor(0, 0, 0, 0)
	r.Clear()
	drawBoarder(r)
	r.SetDrawColor(0xff, 0xff, 0xff, 0xff)
	err := snakeBoi.move()
	if err != nil {
		drawDeadScreen(r)
	}
	r.FillRects(snakeBoi.bodySegmentPositions)
	drawFood(r)
	r.Present()
	sdl.Delay(100)
}

func init() {
	rand.Seed(time.Now().UnixNano())
	for i := int32(10); i <= (widowDimensions - 20); i += 10 {
		Places = append(Places, i)
	}
}

func main() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Fatal(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Snake", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		widowDimensions, widowDimensions, sdl.WINDOW_SHOWN)
	if err != nil {
		log.Fatal(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_SOFTWARE)
	if err != nil {
		log.Fatal(err)
	}
	defer renderer.Destroy()

	surface, err := window.GetSurface()
	if err != nil {
		log.Fatal(err)
	}

	_ = surface.FillRect(nil, 0)

	snakeBoi = newSnake()

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				fmt.Println("Quit")
				running = false
				break
			case *sdl.KeyboardEvent:
				key, ok := mapping[t.Keysym.Scancode]
				if ok && t.Type == sdl.KEYDOWN {
					//fmt.Printf("%+v --- %s\n", event, string(key))
					snakeBoi.userDirection = key
				}
			default:

			}
		}

		runGame(renderer)

	}
}
