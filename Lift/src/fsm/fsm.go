package fsm

import (
	def "../config"
	"../hw"
	"../queue"
	"fmt"
	"log"
	"time"
)

type stateType int // kill this

const (
	idle stateType = iota
	moving
	doorOpen
)

var state stateType
var floor int
var direction int
var departDirection int

var doorReset = make(chan bool)

const doorOpenTime = 1 * time.Second

// --------------- PUBLIC: ---------------

var DoorTimeoutChan = make(chan bool)

func Init() {
	log.Println("FSM Init")
	go startTimer()
	state = idle
	direction = def.DirStop
	floor = hw.Floor()
	if floor == -1 {
		floor = hw.MoveToDefinedState()
	}
	departDirection = def.DirDown
	go syncLights()
}

func EventInternalButtonPressed(buttonFloor int, buttonType int) {
	fmt.Printf("\n\n   ☺      Event internal button (floor %d %s) pressed in state %s\n",
		buttonFloor, buttonString(buttonType), stateString(state))
	queue.Print()
	switch state {
	case idle:
		queue.AddLocalOrder(buttonFloor, buttonType)
		switch direction = queue.ChooseDirection(floor, direction); direction {
		case def.DirStop:
			hw.SetDoorOpenLamp(true)
			queue.RemoveOrdersAt(floor)
			doorReset <- true
			state = doorOpen
		case def.DirUp, def.DirDown:
			hw.SetMotorDirection(direction)
			departDirection = direction
			state = moving
		}
	case doorOpen:
		if floor == buttonFloor {
			doorReset <- true
		} else {
			queue.AddLocalOrder(buttonFloor, buttonType)
		}
	case moving:
		queue.AddLocalOrder(buttonFloor, buttonType)
	default:
		log.Fatalf("State %d is invalid!\n", state)
	}

	def.SyncLightsChan <- true
}

func EventExternalButtonPressed(buttonFloor int, buttonType int) {
	fmt.Printf("\n\n   ☺      Event external button (floor %d %s) pressed in state %s\n",
		buttonFloor, buttonString(buttonType), stateString(state))
	queue.Print()
	switch state {
	case idle, doorOpen, moving:
		// send order on network
		message := def.Message{Kind: def.NewOrder, Floor: buttonFloor, Button: buttonType, Cost: -1}
		def.Outgoing <- message
	default:
		//
	}

	def.SyncLightsChan <- true
}

func EventExternalOrderGivenToMe() {
	fmt.Printf("\n\n   ☺      Event external order given to me.\n")
	queue.Print()

	if queue.IsLocalEmpty() {
		// strange
	}
	switch state {
	case idle:
		switch direction = queue.ChooseDirection(floor, direction); direction {
		case def.DirStop:
			hw.SetDoorOpenLamp(true)
			queue.RemoveOrdersAt(floor)
			doorReset <- true
			state = doorOpen
		case def.DirUp, def.DirDown:
			hw.SetMotorDirection(direction)
			departDirection = direction
			state = moving
		}
	default:
		fmt.Println("   ☺      EventExternalOrderGivenToMe(): Not in idle, will ignore.")
	}
	def.SyncLightsChan <- true
}

func EventFloorReached(newFloor int) {
	fmt.Printf("\n\n   ☺      Event floor %d reached in state %s\n", newFloor, stateString(state))
	queue.Print()
	floor = newFloor
	hw.SetFloorLamp(floor)
	switch state {
	case moving:
		if queue.ShouldStop(floor, direction) {
			hw.SetMotorDirection(def.DirStop)
			hw.SetDoorOpenLamp(true)
			queue.RemoveOrdersAt(floor)
			go queue.SendOrderCompleteMessage(floor)
			doorReset <- true
			state = doorOpen
		} else {
			departDirection = direction
		}
	default:
		log.Printf("Makes no sense to arrive at a floor in state %s.\n", stateString(state))
	}
	def.SyncLightsChan <- true
}

func EventDoorTimeout() {
	fmt.Printf("\n\n   ☺      Event door timeout in state %s\n", stateString(state))
	queue.Print()
	switch state {
	case doorOpen:
		direction = queue.ChooseDirection(floor, direction)
		hw.SetDoorOpenLamp(false)
		hw.SetMotorDirection(direction)
		if direction == def.DirStop {
			state = idle
		} else {
			state = moving
			departDirection = direction
		}
	default:
		log.Fatalf("Makes no sense to time out when not in state door open\n")
	}
	def.SyncLightsChan <- true
}

func Direction() int {
	return direction
}

func DepartDirection() int {
	return departDirection
}

func Floor() int {
	return floor
}

func startTimer() {
	timer := time.NewTimer(0)
	timer.Stop()
	for {
		select {
		case <-doorReset:
			timer.Reset(doorOpenTime)
		case <-timer.C:
			timer.Stop()
			EventDoorTimeout()
		}
	}
}

func syncLights() {
	for {
		<-def.SyncLightsChan

		for f := 0; f < def.NumFloors; f++ {
			for b := 0; b < def.NumButtons; b++ {
				if (b == def.ButtonUp && f == def.NumFloors-1) ||
					(b == def.ButtonDown && f == 0) {
					continue
				} else {
					hw.SetButtonLamp(f, b, queue.IsOrder(f, b))
				}
			}
		}
		time.Sleep(time.Millisecond)
	}
}

func stateString(state stateType) string {
	switch state {
	case idle:
		return "idle"
	case moving:
		return "moving"
	case doorOpen:
		return "door open"
	default:
		return "error: bad state"
	}
}

func buttonString(button int) string {
	switch button {
	case def.ButtonUp:
		return "up"
	case def.ButtonDown:
		return "down"
	case def.ButtonIn:
		return "command"
	default:
		return "error: bad button"
	}
}
