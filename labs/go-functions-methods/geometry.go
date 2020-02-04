package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type Point struct {
	x, y float64
}

func (p Point) X() float64 {
	return p.x
}

func (p Point) Y() float64 {
	return p.y
}

func (p Point) Distance(q Point) float64 {
	return math.Hypot(q.X()-p.X(), q.Y()-p.Y())
}

type Path []Point

func (path Path) Distance() float64 {
	sum := 0.0
	a := len(path)
	for i := range path {
		if i > 0 {
			sum += path[i-1].Distance(path[i])
			if i == a-1 {
				fmt.Printf("%v ", path[i-1].Distance(path[i]))
			} else {
				fmt.Printf("%v + ", path[i-1].Distance(path[i]))
			}
		}
	}
	return sum
}

func main() {
	var sides int
	fmt.Scan(&sides)
	var max int = 100
	var min int = -100
	var operation = max - min
	var path Path

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < sides; i++ {
		var point = Point{float64(rand.Intn(operation) + min), float64(rand.Intn(operation) + min)}
		path = append(path, point)
	}

	fmt.Printf("Generating a [%v] sides figure\n", sides)
	fmt.Println("Figure's vertices")
	for i := 0; i < sides; i++ {
		fmt.Printf("( %v, %v)\n", path[i].X(), path[i].Y())
	}
	fmt.Println("Figure's Perimeter")
	fmt.Printf("= %v \n", path.Distance())
}


