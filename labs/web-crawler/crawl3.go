// Copyright © 2016 Alan A. A. Donovan & Brian W. Kernighan.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// See page 241.

// Crawl2 crawls web links starting with the command-line arguments.
//
// This version uses a buffered channel as a counting semaphore
// to limit the number of concurrent calls to links.Extract.
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gopl.io/ch5/links"
)

//!+sema
// tokens is a counting semaphore used to
// enforce a limit of 20 concurrent requests.
var tokens = make(chan struct{}, 20)

func crawl(url string, depth int, actualDep int) []string {
	var list []string
	if actualDep <= depth {
		fmt.Println(url)
		tokens <- struct{}{} // acquire a token
		list, err := links.Extract(url)
		<-tokens // release the token

		if err != nil {
			log.Print(err)
		}
		return list
	}
	return list
}

//!-sema

//!+
func main() {
	worklist := make(chan []string)
	var n int // number of pending sends to worklist

	// Start with the command-line arguments.
	n++
	args := os.Args[1:]
	s := strings.Split(args[0], "=")
	var start []string
	start = os.Args[2:]
	depth, _ := strconv.Atoi(s[1])
	go func() { worklist <- start }()

	// Crawl the web concurrently.
	seen := make(map[string]bool)
	urlDep := make(map[string]int)
	actualDep := 0
	for ; n > 0; n-- {
		list := <-worklist
		for _, link := range list {
			if !seen[link] {
				aux := actualDep
				aux = aux + 1
				seen[link] = true
				n++
				urlDep[link] = aux
				go func(link string) {
					worklist <- crawl(link, depth, aux)
				}(link)
			}
		}
		actualDep = actualDep + 1
	}
}

