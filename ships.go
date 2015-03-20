package main

const (
	HORIZONTAL = true
	VERTICAL   = false
)

var (
	allShips        []ship
	shipNameToIndex map[string]int
)

type ship struct {
	name      string
	points    int
	length    int
	hitsSoFar int
}

func init() {

	aircraftCarrier := ship{
		name:   "aircraft carrier",
		points: 20,
		length: 5,
	}
	battleship := ship{
		name:   "battleship",
		points: 12,
		length: 4,
	}
	submarine := ship{
		name:   "submarine",
		points: 6,
		length: 3,
	}
	destroyer := ship{
		name:   "destroyer",
		points: 6,
		length: 3,
	}
	cruiser := ship{
		name:   "cruiser",
		points: 6,
		length: 3,
	}
	patrolBoat := ship{
		name:   "patrol boat",
		points: 2,
		length: 2,
	}
	allShips = []ship{
		aircraftCarrier,
		battleship,
		submarine,
		destroyer,
		cruiter,
		patrolBoat,
	}

	for i, s := range allShips {
		shipNameToIndex[s.name] = i
	}
}

func shipByName(name string) ship {
	return allShips[shipNameToIndex[name]]
}
