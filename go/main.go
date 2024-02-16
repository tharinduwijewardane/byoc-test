package main

import (
	"log"
	"time"
)

func main() {
	for {
		log.Println("log flooding test: logging current time: " + time.Now().String())
	}
}
