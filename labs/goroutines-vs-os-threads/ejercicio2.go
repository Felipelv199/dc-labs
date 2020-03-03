package main

import (
	"os"
	"strconv"
	"time"
)

func f1(a, b chan int, value int) {
	a <- value
	x := <-a
	b <- x
}

func f2(a, b chan int) {
	x := <-b
	x = x + 1
	a <- x
}

func main() {
	f, _ := os.Create("report2.txt")
	defer f.Close()
	// Set up the pipeline.
	a := make(chan int)
	b := make(chan int)
	seconds := 10
	sumaVal := 0
	for i := 0; i < seconds; i++ {
		val := 0
		for start := time.Now(); ; {
			go f1(a, b, val)
			go f2(a, b)
			val = <-a
			if time.Since(start) > time.Second {
				s := "Goroutines in second " + strconv.Itoa(i+1) + ": " + strconv.Itoa(val) + "\n"
				f.WriteString(s)
				break
			}
		}
		sumaVal = sumaVal + val
	}
	promedio := float64(sumaVal) / float64(seconds)
	s := "Average of goroutines in a second: " + strconv.FormatFloat(promedio, 'f', -1, 64)
	f.WriteString(s)
}

