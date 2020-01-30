package main

import "golang.org/x/tour/pic"

func Pic(dx, dy int) [][]uint8 {
	p := make([][]uint8, dy)
	for i := 0; i < dy; i++ {
		p[i] = make([]uint8, dx)
	}
	for y := 0; y < dy; y++{
		for x := 0; x < dx; x++{
			p[y][x] = uint8((x+y) / 2)
		}
	}
	return p
}

func main() {
	pic.Show(Pic)
}

