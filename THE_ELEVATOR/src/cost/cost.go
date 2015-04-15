package cost

import (
	"log"
	"../elev"
	"../fsm"
	"../queue"
)

func CalculateCost(targetFloor int, targetButton elev.ButtonType) int {
	// Really need some good error taking care of here
	var targetDirection elev.DirnType

	switch targetButton {
	case elev.ButtonCallUp:
		targetDirection = elev.DirnUp
	case elev.ButtonCallDown:
		targetDirection = elev.DirnDown
	default:
		log.Fatalln("Error direction in cost")
	}

	cost := 0

	floor := elev.GetFloor()
	direction := fsm.GetDirection()
	log.Printf("Floor: %d, direction: %d\n")
	if floor == -1 {
		cost += 1
		floor = incrementFloor(floor, direction) // Is this correct?
	}

	// Loop through floors until target found, and accumulate cost:
	for floor != targetFloor && direction != targetDirection {
		// Handle top/bottom floors:
		if floor <= 0 {
			floor = 0
			direction = elev.DirnUp
		} else if floor >= elev.NumFloors - 1 {
			floor = elev.NumFloors - 1
			direction = elev.DirnDown
		}

		// Go to next floor:
		floor = incrementFloor(floor, direction)

		if queue.ShouldStop(floor, direction) {
			if floor == targetFloor {
				break
			}
			cost += 2
		}
		cost += 2
	}
	return cost
}

func incrementFloor(floor int, direction elev.DirnType) int {
	switch direction {
	case elev.DirnDown:
		floor--
	case elev.DirnUp:
		floor++
	default:
		log.Println("Error: Invalid direction, floor not incremented.")
	}

	return floor
}
