package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/jroimartin/gocui"
	"log"
	"math/rand"
	"os"
	"strconv"
)

const (
	VERTICAL   = true
	HORIZONTAL = false
)

var (
	grid        = make([][]GridSquare, 16)
	shipTypes   = []string{"aircraft_carrier", "battleship", "submarine", "destroyer", "cruiser", "patrol_boat"}
	shipLengths = map[string]int{
		"aircraft_carrier": 5,
		"battleship":       4,
		"submarine":        3,
		"destroyer":        3,
		"cruiser":          3,
		"patrol_boat":      2,
	}
	shipPoints = map[string]int{
		"aircraft_carrier": 20,
		"battleship":       12,
		"submarine":        6,
		"destroyer":        6,
		"cruiser":          6,
		"patrol_boat":      2,
	}
	boatHits map[string]int

	playerScore int
	g           *gocui.Gui
	gBoard      string
)

func init() {
	for i, _ := range grid {
		grid[i] = make([]GridSquare, 16)
	}
	boatHits = make(map[string]int)
}

type GridSquare struct {
	HasShip  bool
	BeenHit  bool
	ShipType string
}

type Coordinate struct {
	X int
	Y int
}

func placeShips() {
	for _, ship := range shipTypes {
		placeBoat(ship)
	}
}

func isSunk(boat []GridSquare) bool {
	for _, square := range boat {
		if !square.BeenHit {
			return false
		}
	}
	return true
}

func placeBoat(boatType string) {
	orientation := randOrientation()
	if orientation == VERTICAL {
		start := Coordinate{
			X: rand.Intn(16),
			Y: rand.Intn(16 - shipLengths[boatType]),
		}
		squares := grid[start.X][start.Y : start.Y+shipLengths[boatType]]
		for _, square := range squares {
			if square.HasShip {
				placeBoat(boatType)
				return
			}
		}

		for i, _ := range squares {
			squares[i].HasShip = true
			squares[i].ShipType = boatType
		}
		return
	} else {
		start := Coordinate{
			X: rand.Intn(16 - shipLengths[boatType]),
			Y: rand.Intn(16),
		}
		rows := grid[start.X : start.X+shipLengths[boatType]]
		for _, row := range rows {
			if row[start.Y].HasShip {
				placeBoat(boatType)
				return
			}
		}

		for i, _ := range rows {
			rows[i][start.Y].HasShip = true
			rows[i][start.Y].ShipType = boatType
		}
		return
	}
}

func randOrientation() bool {
	randomInt := rand.Intn(2)
	if randomInt == 1 {
		return VERTICAL
	} else {
		return HORIZONTAL
	}
}

func shoot(x int, y int) error {
	if x > 15 || y > 15 {
		return errors.New(fmt.Sprintf("You cannot shoot at location (%d, %d): it is off the board\n", x, y))
	}
	square := grid[x][y]
	if square.BeenHit {
		return errors.New(fmt.Sprintf("You have already shot at square (%d, %d)\n", x, y))
	}
	if square.HasShip {
		fmt.Printf("You hit a ship at location (%d, %d)!\n", x, y)
		boatHits[square.ShipType] += 1
		if boatHits[square.ShipType] == shipLengths[square.ShipType] {
			fmt.Printf("You've sunk my %s!\n", square.ShipType)
			playerScore += shipPoints[square.ShipType]
		}
	} else {
		fmt.Printf("You missed at (%d, %d)\n", x, y)
		playerScore -= 1
	}
	grid[x][y].BeenHit = true

	return nil
}

func getCoordinate() (x int, y int) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Input a coordinate:")
	fmt.Printf("\tx: ")
	xStr, err := reader.ReadString([]byte("\n")[0])
	if err != nil {
		fmt.Printf("Error reading from stdin: %s", err.Error())
		return getCoordinate()
	}
	if len(xStr) > 1 {
		xStr = xStr[:len(xStr)-1]
	}
	fmt.Printf("\ty: ")
	yStr, err := reader.ReadString([]byte("\n")[0])
	if err != nil {
		fmt.Printf("Error reading from stdin: %s", err.Error())
		return getCoordinate()
	}
	if len(yStr) > 1 {
		yStr = yStr[:len(yStr)-1]
	}
	x, err = strconv.Atoi(xStr)
	if err != nil {
		fmt.Printf("Error converting `%s` to int: %s\n", xStr, err.Error())
		return getCoordinate()
	}
	y, err = strconv.Atoi(yStr)
	if err != nil {
		fmt.Printf("Error converting `%s` to int: %s\n", xStr, err.Error())
		return getCoordinate()
	}
	return x, y
}

func takeTurn() {
	coordinates := make([]Coordinate, 5)
	for i := 0; i < 5; i++ {
		//x, y := getCoordinate()
		coordinates[i] = Coordinate{
			X: i,
			Y: i,
		}
	}

	var fails int
	for _, point := range coordinates {
		err := shoot(point.X, point.Y)
		if err != nil {
			fmt.Printf("Error shooting at (%d, %d): %s", point.X, point.Y, err.Error())
			fails += 1
		}
		if haveYouWon() {
			win()
			return
		}
	}

	for fails > 0 {
		coordinates := make([]Coordinate, fails)
		for i := 0; i < fails; i++ {
			x, y := getCoordinate()
			coordinates[i] = Coordinate{
				X: x,
				Y: y,
			}
		}

		fails = 0
		for _, point := range coordinates {
			err := shoot(point.X, point.Y)
			if err != nil {
				fmt.Printf("Error shooting at (%d, %d): %s\n", point.X, point.Y, err.Error())
				fails += 1
			}
		}
	}
}

func win() {
	fmt.Printf("You won!\n")
	fmt.Printf("Your final score was: %d\n", playerScore)
	return
}

func lose() {
	fmt.Printf("Unfortunately, you lost.\n")
	fmt.Printf("Your final score was: %d\n", playerScore)
}

func haveYouWon() bool {
	for _, boat := range shipTypes {
		if boatHits[boat] < shipLengths[boat] {
			return false
		}
	}
	return true
}

func playGame(g *gocui.Gui) {
	for i := 0; i < 1; i++ {
		takeTurn()
		if haveYouWon() {
			win()
			return
		}
		//fmt.Printf("After turn %d, your score is %d\n", i+1, playerScore)
		s := prepBoardAndCheat()
		//fmt.Print(s)
		gBoard = s
		g.Flush()
		renderBoard(g, s)
	}

	if haveYouWon() {
		win()
		return
	} else {
		lose()
		return
	}
}

func prepBoard() (s string) {
	s = ""
	for _, row := range grid {
		line := ""
		for _, cell := range row {
			var c string
			if cell.BeenHit {
				c = "X"
			} else {
				c = "."
			}
			line += c
		}
		line += fmt.Sprintln()
		s += line
	}
	return s
}

func prepBoardAndCheat() (s string) {
	s = ""
	for _, row := range grid {
		line := ""
		for _, cell := range row {
			var c string
			switch {
			case cell.BeenHit:
				c = "X"
			case cell.HasShip:
				c = "B"
			default:
				c = "."
			}
			line += c
		}
		line += fmt.Sprintln()
		s += line
	}
	return s
}
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("center", 0, 0, maxX, maxY); err != nil {
		if err != gocui.ErrorUnkView {
			return err
		}
		fmt.Fprintln(v, gBoard)
	}
	return nil
}
func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.Quit
}
func renderBoard(g *gocui.Gui, board string) {
	x := 0
	y := 0
	for _, c := range board {
		if c == '\n' {
			y += 1
			x = 0
			continue
		}
		g.SetRune(x, y, c)
		x++
	}
}

func main() {
	placeShips()
	for i, row := range grid {
		for j, cell := range row {
			var c string
			if cell.ShipType != "" {
				c = fmt.Sprintf("X")
			} else {
				c = fmt.Sprintf("%d,%d", i, j)
			}
			fmt.Printf("%v\t", c)
		}
		fmt.Printf("\n")
	}

	var err error
	g = gocui.NewGui()
	if err := g.Init(); err != nil {
		log.Panicln(err)
	}
	defer g.Close()
	g.SetLayout(layout)
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	err = g.MainLoop()
	if err != nil && err != gocui.Quit {
		log.Panicln(err)
	}
	playGame(g)
}
