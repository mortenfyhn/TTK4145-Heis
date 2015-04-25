package config

import (
	"net"
	"time"
	"os/exec"
)

// Global system constants
const NumButtons = 3
const NumFloors = 4

const (
	ButtonUp int = iota
	ButtonDown
	ButtonCommand //todo: Rename to ButtonInternal or something
)

const (
	DirDown int = iota - 1
	DirStop
	DirUp
)

const MaxInt = int(^uint(0) >> 1)

//constants for sending aliveMsg, and detecting deaths
const SpamInterval = 400 * time.Millisecond
const OnlineLiftResetTime = 2 * time.Second //todo: name?

//constants setting the time the elevators wait before, the cost and the order times out
const OrderTime = 10 * time.Second //todo: name? //set this to 30 Seconds after done debuging
const CostTime = 10 * time.Second

//message kind constants
const (
	Alive int = iota + 1
	NewOrder
	CompleteOrder
	Cost
)

type Keypress struct {
	Button int
	Floor  int
}

// Generic network message. No other messages are ever sent on the network.
type Message struct {
	Kind   int
	Floor  int
	Button int
	Cost   int
	Addr   string `json:"-"`
}

var MessageChan = make(chan Message) // vurder buff //todo: change name to outgoingMessages or something
var SyncLightsChan = make(chan bool)
var CloseConnectionChan = make(chan bool)

//Local address
var Laddr *net.UDPAddr //todo: make this string

//Start a new terminal when restart.Run()
var Restart = exec.Command("gnome-terminal", "-x", "sh", "-c", "lift")

