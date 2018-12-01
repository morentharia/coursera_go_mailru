package main

import (
	"fmt"
	"time"
)

func sayHello() {
	fmt.Println("Hello World")
}

func main() {
	timer := time.AfterFunc(1*time.Second, sayHello)

	fmt.Println("wait")
	fmt.Scanln()
	timer.Stop()

	// fmt.Scanln()
}
