package main

import(
	. "fmt"
	"runtime"
	"time"
)

var(
	i int = 0
) 

func increase(channel chan int){
	for j := 0; j < 1000000; j++{
		i = <- channel
		i++
		channel <- i
	}
}

func decrease(channel chan int){
	for j := 0; j < 1000000; j++{
		i = <- channel
		i--
		channel <- i
	}
}

func main() {
	
	channel := make(chan int, 1);

	channel <- i

	runtime.GOMAXPROCS(runtime.NumCPU())

	go increase(channel)

	go decrease(channel)

	time.Sleep(400*time.Millisecond)
	
	Println(i)
}
