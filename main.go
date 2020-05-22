package main

import (
	"fmt"
	"github.com/pkg/term"
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	rows    = 25 //X
	columns = 40 //Y
)

const (
	RuneSq       = '■'
	RuneDArrow   = '↓'
	RuneLArrow   = '←'
	RuneRArrow   = '→'
	RuneUArrow   = '↑'
	RuneDiamond  = '◆'
	RunePi       = 'π'
	RuneHLine    = '─'
	RuneLLCorner = '└'
	RuneLRCorner = '┘'
	RuneULCorner = '┌'
	RuneURCorner = '┐'
	RuneVLine    = '│'
)

const (
	UP    = RuneUArrow
	DOWN  = RuneDArrow
	LEFT  = RuneLArrow
	RIGHT = RuneRArrow
	FOOD  = RunePi
)

type snake struct {
	bodySegmentPositions []segment
	segmentChar          rune
	userDirection        rune
	heading              rune
}

type segment struct {
	x int
	y int
}

var (
	Score        int
	FoodPlaced   = false
	FoodLocation segment
	MOVEMENT     = map[rune]segment{UP: {0, -1}, DOWN: {0, 1}, LEFT: {-1, 0}, RIGHT: {1, 0}}
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
	snakeBoi = snake{
		bodySegmentPositions: []segment{{10, 5}, {10, 4}, {10, 3}}, //initial position for snakeBoi
		segmentChar:          RuneSq,
		userDirection:        0,
		heading:              RIGHT,
	}
	userMove = make(chan rune, 1)
)

var grid [][]rune

func init() {
	rand.Seed(time.Now().UnixNano())
}

func containsFood(head segment) bool {
	if head == FoodLocation {
		return true
	}
	return false
}

func placeFood() {

	for {
		randX := rand.Intn(columns-2) + 2
		randY := rand.Intn(rows-1) + 1

		if grid[randY][randX] == 0 {
			FoodLocation = segment{
				x: randX,
				y: randY,
			}
			break
		}
	}

}

func drawFood() {
	grid[FoodLocation.y][FoodLocation.x] = FOOD
}

func (s *snake) move() {
	if s.userDirection != 0 {
		fmt.Print("")
	}
	head := segment{}
	head.x, head.y = s.calculatePosition(s.bodySegmentPositions[0].x, s.bodySegmentPositions[0].y)
	//fmt.Printf("current head at %d,%d heading %s user sent %s - new head at %d,%d\r\n",
	//	s.bodySegmentPositions[0].x,
	//	s.bodySegmentPositions[0].y,
	//	string(s.heading),
	//	string(s.userDirection),
	//	head.x,
	//	head.y)

	s.bodySegmentPositions = append([]segment{head}, s.bodySegmentPositions...)

	foodFound := containsFood(head)
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
}

func (s *snake) calculatePosition(x, y int) (int, int) {
	if s.heading == DOWN && s.userDirection == UP {
		fmt.Println()
	}
	n, ok := heading[s.heading][s.userDirection]

	if !ok {
		n = MOVEMENT[s.heading]
	}

	return x + n.x, y + n.y
}

func (s *snake) draw() error {
	for _, segment := range s.bodySegmentPositions {
		if grid[segment.y][segment.x] != 0 && grid[segment.y][segment.x] != FOOD {
			return fmt.Errorf("🐍DED🐍")
		}
		grid[segment.y][segment.x] = s.segmentChar
	}
	return nil
}

func boxContent(char string) {
	for row := 0; row < rows; row++ {
		for column := 0; column < columns; column++ {
			s := string(grid[row][column])
			if grid[row][column] == 0 {
				fmt.Print(char)
			} else {
				fmt.Printf("%s", string(s))
			}

		}
		fmt.Print("\r\n")
	}
}

func redrawBox() {
	print("\033[H\033[2J")
	grid = make([][]rune, rows)
	for i := 0; i < rows; i++ {
		grid[i] = make([]rune, columns)
	}

	for index, value := range grid {
		grid[index][len(value)-1] = RuneVLine
		grid[index][0] = RuneVLine
	}

	grid[0] = runeMultiplier(RuneHLine, columns)
	grid[len(grid)-1] = runeMultiplier(RuneHLine, columns)

	grid[0][0] = RuneULCorner
	grid[0][columns-1] = RuneURCorner
	grid[rows-1][0] = RuneLLCorner
	grid[rows-1][columns-1] = RuneLRCorner
}

func runeMultiplier(i rune, count int) []rune {
	var x []rune
	for v := 0; v < count; v++ {
		x = append(x, i)
	}
	return x
}

// Returns rune representing arrow direction
func getChar(t *term.Term) {

	for {
		var keyCode rune

		bytes := make([]byte, 3)
		var numRead int
		numRead, err := t.Read(bytes)
		if err != nil {
			log.Fatal(err)
		}
		//catch escape key and allow user to leave
		if len(bytes) >= 1 {
			if bytes[0] == 27 && bytes[1] == 0 {
				cleanUp(t)
			}
		}

		if numRead == 3 && bytes[0] == 27 && bytes[1] == 91 {
			// Three-character control sequence, beginning with "ESC-[".

			// Since there are no ASCII codes for arrow keys, we use
			if bytes[2] == 65 {
				// Up
				keyCode = UP
			} else if bytes[2] == 66 {
				// Down
				keyCode = DOWN
			} else if bytes[2] == 67 {
				// Right
				keyCode = RIGHT
			} else if bytes[2] == 68 {
				// Left
				keyCode = LEFT
			}
		}

		select {
		case userMove <- keyCode:
		default:

		}
		time.Sleep(time.Millisecond * 25)
	}
}

func main() {
	t, _ := term.Open("/dev/tty")
	term.RawMode(t)
	fmt.Print("\033[?25l")
	defer cleanUp(t)
	go getChar(t)

	for {
		redrawBox()
		select {
		case d := <-userMove:
			//fmt.Print(string(d),"-.\r\n")
			snakeBoi.userDirection = d
		default:
			//fmt.Print(".d\r\n")
		}

		snakeBoi.move()
		err := snakeBoi.draw()
		if err != nil {
			fmt.Println(err.Error(), "\r")
			cleanUp(t)
		}
		if !FoodPlaced {
			placeFood()
			FoodPlaced = true
		}
		drawFood()
		boxContent(" ")
		fmt.Printf("Current Score %d", Score)
		if snakeBoi.heading == DOWN || snakeBoi.heading == UP {
			time.Sleep(time.Millisecond * 160)
		} else {
			time.Sleep(time.Millisecond * 80)
		}

	}
	cleanUp(t)
}

func cleanUp(t *term.Term) {
	t.Restore()
	t.Close()
	fmt.Print("\033[?25h")
	os.Exit(0)
}
