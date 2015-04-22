// This file is a golang port of elev.c from the hand out driver
// https://github.com/TTK4145/Project
package hw

import (
	def "../config"
	"log"
)

var lampChannelMatrix = [def.NumFloors][def.NumButtons]int{
	{LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1},
	{LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2},
	{LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
	{LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4},
}
var buttonChannelMatrix = [def.NumFloors][def.NumButtons]int{
	{BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
	{BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
	{BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
	{BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4},
}

func Init(motorDir chan int,
	doorLamp chan bool,
	floorLamp chan int) (floor int) {
	if !ioInit() {
		log.Fatalln("hw.Init(): ioInit() failed!")
	}
	log.Println("ioInit() completed...")
	// Zero all floor button lamps
	for f := 0; f < def.NumFloors; f++ {
		if f != 0 {
			SetButtonLamp(f, def.ButtonDown, false)
		}
		if f != def.NumFloors-1 {
			SetButtonLamp(f, def.ButtonUp, false)
		}
		SetButtonLamp(f, def.ButtonIn, false)
	}
	
	stopLamp := make(chan bool)
	
	// Start reading hw control channels:
	go setMotorDir(motorDir)
	go setDoorLamp(doorLamp)
	go setStopLamp(stopLamp)
	go setFloorLamp(floorLamp)
	
	doorLamp <- false
	stopLamp <- false
	
	// Move to defined state:
	motorDir <- def.DirDown
	for floor = Floor(); floor == -1; {
		floor = Floor()
	}
	motorDir <- def.DirStop
	floorLamp <- floor
	
	log.Println("hw.Init() returning...")
	return floor
}

func setMotorDir(motorDir <-chan int) {
	var dir int
	for {
		dir = <-motorDir
		if dir == 0 {
			ioWriteAnalog(MOTOR, 0)
		} else if dir > 0 {
			ioClearBit(MOTORDIR)
			ioWriteAnalog(MOTOR, 2800)
		} else if dir < 0 {
			ioSetBit(MOTORDIR)
			ioWriteAnalog(MOTOR, 2800)
		}
	}
}

func setDoorLamp(doorLamp <-chan bool) {
	for {
		value := <- doorLamp
		if value {
			ioSetBit(LIGHT_DOOR_OPEN)
		} else {
			ioClearBit(LIGHT_DOOR_OPEN)
		}
	}
}

func Floor() int {
	if ioReadBit(SENSOR_FLOOR1) {
		return 0
	} else if ioReadBit(SENSOR_FLOOR2) {
		return 1
	} else if ioReadBit(SENSOR_FLOOR3) {
		return 2
	} else if ioReadBit(SENSOR_FLOOR4) {
		return 3
	} else {
		return -1
	}
}

func setFloorLamp(floorLamp <-chan int) {
	for {
		floor := <- floorLamp
		if floor < 0 || floor >= def.NumFloors {
			log.Printf("Error: Floor %d out of range!\n", floor)
			log.Println("No floor indicator will be set.")
			return
		}

		// Binary encoding. One light must always be on.
		if floor&0x02 > 0 {
			ioSetBit(LIGHT_FLOOR_IND1)
		} else {
			ioClearBit(LIGHT_FLOOR_IND1)
		}

		if floor&0x01 > 0 {
			ioSetBit(LIGHT_FLOOR_IND2)
		} else {
			ioClearBit(LIGHT_FLOOR_IND2)
		}
	}
}

func ReadButton(floor int, button int) bool {
	if floor < 0 || floor >= def.NumFloors {
		log.Printf("Error: Floor %d out of range!\n", floor)
		return false
	}
	if button < 0 || button >= def.NumButtons {
		log.Printf("Error: Button %d out of range!\n", button)
		return false
	}
	if button == def.ButtonUp && floor == def.NumFloors-1 {
		log.Println("Button up from top floor does not exist!")
		return false
	}
	if button == def.ButtonDown && floor == 0 {
		log.Println("Button down from ground floor does not exist!")
		return false
	}

	if ioReadBit(buttonChannelMatrix[floor][button]) {
		return true
	} else {
		return false
	}
}

func SetButtonLamp(floor int, button int, value bool) {
	if floor < 0 || floor >= def.NumFloors {
		log.Printf("Error: Floor %d out of range!\n", floor)
		return
	}
	if button == def.ButtonUp && floor == def.NumFloors-1 {
		log.Println("Button up from top floor does not exist!")
		return
	}
	if button == def.ButtonDown && floor == 0 {
		log.Println("Button down from ground floor does not exist!")
		return
	}
	if button != def.ButtonUp &&
		button != def.ButtonDown &&
		button != def.ButtonIn {
		log.Printf("Invalid button %d\n", button)
		return
	}

	if value {
		ioSetBit(lampChannelMatrix[floor][button])
	} else {
		ioClearBit(lampChannelMatrix[floor][button])
	}
}

// Not used:

func setStopLamp(stopLamp <-chan bool) {
	for {
		value := <- stopLamp
		if value {
			ioSetBit(LIGHT_STOP)
		} else {
			ioClearBit(LIGHT_STOP)
		}
	}
}

func GetObstructionSignal() bool {
	return ioReadBit(OBSTRUCTION)
}

func GetStopSignal() bool {
	return ioReadBit(STOP)
}