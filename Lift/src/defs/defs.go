package defs

import (
	"net"
	"strings"
	"time"
)

// Global system constants
const NumButtons = 3
const NumFloors = 4

const (
	ButtonUp int = iota
	ButtonDown
	ButtonCommand // Rename to ButtonInternal or something
)

const (
	DirDown int = iota - 1
	DirStop
	DirUp
)

const MaxInt = int(^uint(0) >> 1)

//constants for sending aliveMsg, and detecting deaths
const SpamInterval = 400 * time.Millisecond
const ResetTime = 2 * time.Second

const (
	Alive int = iota + 1
	NewOrder
	CompleteOrder
	Cost
)

// Generic network message. No other messages are ever sent on the network.
type Message struct {
	Kind   int
	Floor  int
	Button int
	Cost   int
	Addr   string `json:"-"`
}

var MessageChan = make(chan Message) // vurder buff
var SyncLightsChan = make(chan bool)

var Laddr *net.UDPAddr //Local address

func LastPartOfIp(ip string) string {
	return strings.Split(strings.Split(ip, ".")[3], ":")[0]
}
