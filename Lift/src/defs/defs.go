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
	ButtonCallUp int = iota
	ButtonCallDown
	ButtonCommand // Rename to ButtonInternal or something
)

const ( // Rename to DirDown etc.
	DirnDown int = iota - 1
	DirnStop
	DirnUp
)

const (
	Alive int = iota + 1
	NewOrder
	CompleteOrder
	Cost
)

const MaxInt = int(^uint(0) >> 1)

const SpamInterval = 5000 * time.Millisecond
const ResetTime = 11 * time.Second

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
