package cost

import (
	"../defs"
	"../queue"
	//"errors"
	"fmt"
	"log"
)

// --------------- PUBLIC: ---------------


// CalculateCost calculates how much effort it takes this lift to carry out
// the given order. Each sheduled stop on the way there and each travel
// between adjacent floors will add cost 2. Cost 1 is added if the lift
// starts between floors.
func CalculateCost(targetFloor, targetButton, fsmFloor, fsmDir, currFloor int) (int, error) {
	if (targetButton != defs.ButtonCallUp) && (targetButton != defs.ButtonCallDown) {
		return 0, fmt.Errorf("CalculateCost() called with invalid order: %d\n", targetButton)
	}

	cost := 0
	
	if currFloor == -1 {
		cost += 1
		fsmFloor = incrementFloor(fsmFloor, fsmDir)
	}

	for !(fsmFloor == targetFloor && queue.ShouldStop(fsmFloor, fsmDir)) {
		log.Printf("Floor: %d, direction: %d\n", fsmFloor, fsmDir)
		if queue.ShouldStop(fsmFloor, fsmDir) {
			cost += 2
		}
		fsmDir = queue.ChooseDirection(fsmFloor, fsmDir)
		fsmFloor = incrementFloor(fsmFloor, fsmDir)
		cost += 2
	}

	return cost, nil
}

func incrementFloor(floor int, direction int) int {
	switch direction {
	case defs.DirnDown:
		floor--
	case defs.DirnUp:
		floor++
	case defs.DirnStop:
		log.Println("Error(ish): Direction stop, floor not incremented.")
	default:
		log.Println("Error: Invalid direction, floor not incremented.")
	}

	return floor
}