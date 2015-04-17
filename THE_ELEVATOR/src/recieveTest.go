package main

import (
	"./network"
	"fmt"
)

func main() {
	network.Init()
	go network.ReceiveMsg()
	for {
		connection := <- network.ConnectionTimer
		fmt.Println(connection.Addr, "is dead")
	}
}
