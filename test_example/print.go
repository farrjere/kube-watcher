package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {

	for {
		time.Sleep(time.Second * 2)
		fmt.Printf("Hello, World %v!\n", rand.Int())
	}
}
