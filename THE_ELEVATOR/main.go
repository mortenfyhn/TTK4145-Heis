package main

import (
	"./src/cost"
	"./src/defs"
	"./src/fsm"
	"./src/hw"
	"./src/network"
	"./src/queue"
	"errors"
	"fmt"
	"log"
	"math"
	"time"
)

const debugPrint = false
const resetTime = 1 * time.Second

var _ = log.Println
var _ = fmt.Println
var _ = errors.New

type keypress struct {
	button int
	floor  int
}

var onlineLifts = make(map[string]network.UdpConnection)

var deadChan = make(chan network.UdpConnection)
var costChan = make(chan defs.Message)

type reply struct {
	cost int
	lift string
}
type order struct {
	floor  int
	button int
}

func main() {
	if err := hw.Init(); err != nil {
		log.Fatal(err)
	}
	queue.Init()
	fsm.Init()
	network.Init()

	go run()
}

func run() {
	buttonChan := pollButtons()
	floorChan := pollFloors()

	for {
		select {
		case keypress := <-buttonChan:
			fsm.EventButtonPressed(keypress.floor, keypress.button)
		case floor := <-floorChan:
			fsm.EventFloorReached(floor)
		case udpMessage := <-network.ReceiveChan:
			handleMessage(network.ParseMessage(udpMessage))
		case connection := <-deadChan:
			handleDeadLift(connection.Addr)
		}
	}
}

func pollButtons() <-chan keypress {
	c := make(chan keypress)

	go func() {
		var buttonState [defs.NumFloors][defs.NumButtons]bool

		for {
			for f := 0; f < defs.NumFloors; f++ {
				for b := 0; b < defs.NumButtons; b++ {
					if (f == 0 && b == defs.ButtonCallDown) ||
						(f == defs.NumFloors-1 && b == defs.ButtonCallUp) {
						continue
					}
					if hw.ReadButton(f, b) {
						if !buttonState[f][b] {
							c <- keypress{button: b, floor: f}
						}
						buttonState[f][b] = true
					} else {
						buttonState[f][b] = false
					}
				}
			}
			time.Sleep(time.Millisecond)
		}
	}()

	return c
}

func pollFloors() <-chan int {
	c := make(chan int)

	go func() {
		oldFloor := hw.GetFloor()

		for {
			newFloor := hw.GetFloor()
			if newFloor != oldFloor && newFloor != -1 {
				c <- newFloor
			}
			oldFloor = newFloor
			time.Sleep(time.Millisecond)
		}
	}()

	return c
}

func handleMessage(message defs.Message) {
	switch message.Kind {
	case defs.Alive:
		if connection, exist := onlineLifts[message.Addr]; exist {
			connection.Timer.Reset(resetTime)
			if debugPrint {
				fmt.Printf("Timer reset for IP %s\n", message.Addr)
			}
		} else {
			newConnection := network.UdpConnection{message.Addr, time.NewTimer(resetTime)}
			onlineLifts[message.Addr] = newConnection
			if debugPrint {
				fmt.Printf("New connection with IP %s\n", message.Addr)
			}
			go connectionTimer(&newConnection)
		}
	case defs.NewOrder:
		cost, err := cost.CalculateCost(message.Floor, message.Button, fsm.GetFloor(), fsm.GetDirection(), hw.GetFloor())
		if err != nil {
			log.Println(err)
		}
		costMessage := &defs.Message{
			Kind:   defs.Cost,
			Floor:  message.Floor,
			Button: message.Button,
			Cost:   cost}
		network.Send(costMessage)
	case defs.CompleteOrder:
		// remove from queues
		queue.RemoveSharedOrder(message.Floor, message.Button)
		// prob more to do here
	case defs.Cost:
		costChan <- message
	}
}

func handleDeadLift(deadAddr string) {
	fmt.Printf("Connection to IP %s is dead!\n", deadAddr)
	delete(onlineLifts, deadAddr)
	queue.ReassignOrders(deadAddr)
}

func connectionTimer(connection *network.UdpConnection) {
	for {
		<-connection.Timer.C
		deadChan <- *connection
	}
}

func liftAssigner() {
	// collect cost values from all lifts
	// decide which lift gets the order when all lifts
	// in alive-list have answered or after a timeout
	// either send the decision on network or pray that all
	// lifts make the same choice every time
	go func() {
		assignmentQueue := make(map[order][]reply)
		for {
			select {
			case message := <-costChan:
				newOrder, newReply := split(message)
				// Check if order in queue
				if value, exist := assignmentQueue[newOrder]; exist {
					// Check if lift in list of that order
					found := false
					for _, e := range value {
						if e == newReply {
							found = true
						}
					}
					// Add it if not found
					if !found {
						assignmentQueue[newOrder] = append(assignmentQueue[newOrder], newReply)
					}
				} else {
					// If order not in queue at all, init order list with it
					assignmentQueue[newOrder] = []reply{newReply}
				}
				evaluateLists(assignmentQueue)
			default:
				// worry
			}
		}
	}()
}

func split(m defs.Message) (order, reply) {
	return order{floor: m.Floor, button: m.Button}, reply{cost: m.Cost, lift: m.Addr}
}

// This is very cryptic and ungood.
func evaluateLists(queue map[order][]reply) {
	// Loop thru all lists
	for key, replyList := range queue {
		// Check if the list is complete
		if len(replyList) == len(onlineLifts) {
			var (
				lowKey  int
				lowCost = math.Inf(1)
				lowAddr
			)
			// Loop thru costs in each complete list
			for i, reply := range replyList {
				if reply.cost < lowCost {
					lowCost = reply.cost
					lowKey = i
					lowAddr = reply.lift
				}
			}
			// Assign order key to lift
			queue.AddSharedOrder(lowKey.floor, lowKey.button, lowAddr)
		}
	}
}
