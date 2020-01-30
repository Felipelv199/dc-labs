package main

import (
	"golang.org/x/tour/wc"
	"strings"
	"fmt"
)

func WordCount(s string) map[string]int {
	m := make(map[string]int)

	a := strings.Split(s, " ")
	count := 0
	fmt.Println(len(a))

	for i := 0; i < len(a); i++{
		count++
		for j := 0; j < len(a); j++{
			if i != j{
				if a[i] == a[j]{
					count++
				}
			}
		}
		m[a[i]] = count
		count = 0
	}
	return m
}

func main() {
	wc.Test(WordCount)
}

