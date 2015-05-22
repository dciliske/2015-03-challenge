package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"

	"github.com/jroimartin/gocui"
)

const (
	VERTICAL   = true
	HORIZONTAL = false

	BOARD  = "board"
	STATUS = "status"
	SCORE  = "score"
	SHOTS  = "shots"
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
	shotsLeft   int

	cheatFlag = flag.Bool("cheat", false, "Use this flag to cheat.")
)

func init() {
	for i, _ := range grid {
		grid[i] = make([]GridSquare, 16)
	}
	boatHits = make(map[string]int)
}

func resetGrid() {
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

func haveYouWon() bool {
	for _, boat := range shipTypes {
		if boatHits[boat] < shipLengths[boat] {
			return false
		}
	}
	return true
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.Quit
}

func renderBoard(g *gocui.Gui) {
	g.SetCurrentView(BOARD)
	g.CurrentView().Clear()
	for _, row := range grid {
		for _, cell := range row {
			var c string
			switch {
			case cell.BeenHit:
				c = "X"
			case *cheatFlag && cell.HasShip:
				c = "B"
			default:
				c = "."
			}
			fmt.Fprintf(g.CurrentView(), "%v ", c)
		}
		fmt.Fprintf(g.CurrentView(), "\n")
	}
}

func renderScore(g *gocui.Gui) {
	g.SetCurrentView(SCORE)
	defer g.SetCurrentView(BOARD)
	g.CurrentView().Clear()
	fmt.Fprintf(g.CurrentView(), "score\n%d", playerScore)
}

func renderShotsLeft(g *gocui.Gui) {
	g.SetCurrentView(SHOTS)
	defer g.SetCurrentView(BOARD)
	g.CurrentView().Clear()
	fmt.Fprintf(g.CurrentView(), "left:\n%d", shotsLeft)
}

func thisMyLayout(gui *gocui.Gui) error {
	renderBoard(gui)
	renderScore(gui)
	renderShotsLeft(gui)
	return nil
}

func MoveRight(gui *gocui.Gui, view *gocui.View) error {
	x, _ := view.Cursor()
	if x < 32 {
		view.MoveCursor(2, 0, false)
	}
	return nil
}

func MoveLeft(gui *gocui.Gui, view *gocui.View) error {
	x, _ := view.Cursor()
	if x > 0 {
		view.MoveCursor(-2, 0, false)
	}
	return nil
}

func MoveUp(gui *gocui.Gui, view *gocui.View) error {
	_, y := view.Cursor()
	if y > 0 {
		view.MoveCursor(0, -1, false)
	}
	return nil
}

func MoveDown(gui *gocui.Gui, view *gocui.View) error {
	_, y := view.Cursor()
	if y < 15 {
		view.MoveCursor(0, 1, false)
	}
	return nil
}

func WriteStatus(gui *gocui.Gui, statusMsg string, a ...interface{}) error {
	v, err := gui.View(STATUS)
	if err != nil {
		return err
	}
	v.Clear()
	_, err = fmt.Fprintf(v, statusMsg, a...)
	return err
}

func ShootAt(gui *gocui.Gui, view *gocui.View) error {
	defer gui.Flush()
	if shotsLeft <= 0 {
		return WriteStatus(gui, "You're out of shots. You have lost. You are dead. Press tab to play again.")
	}
	if haveYouWon() {
		return WriteStatus(gui, "You won! Press tab to play again.")
	}
	defer func() {
		if haveYouWon() {
			WriteStatus(gui, "You won! Press tab to play again.")
		} else if shotsLeft <= 0 {
			WriteStatus(gui, "You're out of shots. You have lost. You are dead. Press tab to play again.")
		}
	}()
	x, y := view.Cursor()
	x /= 2

	square := grid[y][x]
	if square.BeenHit {
		return WriteStatus(gui, "You already shot there. Try again")
	}
	shotsLeft -= 1
	grid[y][x].BeenHit = true

	if square.HasShip {
		boatHits[square.ShipType] += 1
		if boatHits[square.ShipType] == shipLengths[square.ShipType] {
			playerScore += shipPoints[square.ShipType]
			return WriteStatus(gui, "You sunk my %s at (%d, %d)", square.ShipType, x, y)
		} else {
			return WriteStatus(gui, "You hit a ship at square (%d, %d)", x, y)
		}
	} else {
		playerScore--
		return WriteStatus(gui, "You missed at (%d, %d)", x, y)
	}
	return nil
}

func Restart(gui *gocui.Gui, view *gocui.View) error {
	defer gui.Flush()
	resetGrid()
	shotsLeft = 6 * 5
	placeShips()
	playerScore = 0
	return nil
}

func main() {
	flag.Parse()

	g = gocui.NewGui()
	if err := g.Init(); err != nil {
		log.Panicln(err.Error())
	}
	defer g.Close()
	g.SetLayout(thisMyLayout)
	g.ShowCursor = true

	boardView, err := g.SetView(BOARD, 0, 0, 34, 17)
	if err != nil && err != gocui.ErrorUnkView {
		log.Panicln(err.Error())
	}
	boardView.Overwrite = true
	boardView.Wrap = false
	boardView.Autoscroll = false
	err = g.SetCurrentView(BOARD)
	if err != nil && err != gocui.ErrorUnkView {
		log.Panicln(err.Error())
	}

	Restart(g, boardView)

	statusView, err := g.SetView(STATUS, 0, 18, 34, 25)
	statusView.Autoscroll = false
	statusView.Wrap = true
	statusView.Overwrite = true
	err = g.SetCurrentView(BOARD)
	if err != nil {
		log.Panicln(err.Error())
	}

	scoreView, err := g.SetView(SCORE, 36, 0, 45, 3)
	scoreView.Autoscroll = false
	scoreView.Wrap = false
	scoreView.Overwrite = true

	shotsLeftView, err := g.SetView(SHOTS, 36, 5, 45, 8)
	shotsLeftView.Autoscroll = false
	shotsLeftView.Wrap = false
	shotsLeftView.Overwrite = true

	g.SetKeybinding(BOARD, gocui.KeyArrowUp, gocui.ModNone, MoveUp)
	g.SetKeybinding(BOARD, gocui.KeyArrowDown, gocui.ModNone, MoveDown)
	g.SetKeybinding(BOARD, gocui.KeyArrowLeft, gocui.ModNone, MoveLeft)
	g.SetKeybinding(BOARD, gocui.KeyArrowRight, gocui.ModNone, MoveRight)
	g.SetKeybinding(BOARD, gocui.KeyEnter, gocui.ModNone, ShootAt)
	g.SetKeybinding(BOARD, gocui.KeySpace, gocui.ModNone, ShootAt)
	g.SetKeybinding(BOARD, gocui.KeyEsc, gocui.ModNone, Restart)
	g.SetKeybinding(BOARD, gocui.KeyBackspace, gocui.ModNone, Restart)
	g.SetKeybinding(BOARD, gocui.KeyTab, gocui.ModNone, Restart)

	//g.SetLayout(layout)
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	err = g.MainLoop()
	if err != nil && err != gocui.Quit {
		log.Panicln(err)
	}
}
