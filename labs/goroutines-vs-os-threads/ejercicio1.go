package main

import (
	"os"
	"strconv"
)

func gen(a, b chan int, c int) {
	aux := <-a
	b <- aux
}

func main() {
	f, _ := os.Create("report1.txt")

	defer f.Close()
	// Set up the pipeline.
	a := make(chan int)
	b := make(chan int)
	counter := 0
	for {
		go gen(b, a, counter)
		a = b
		b = make(chan int)
		counter = counter + 1
		counterString := strconv.Itoa(counter)
		counterString = "\r" + counterString
		f.WriteString(counterString)
	}
}
