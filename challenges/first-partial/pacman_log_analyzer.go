package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

type File struct {
	pckgName     string
	installDate  string
	lastUpdtDate string
	nUpdt        int
	removalDate  string
}

func main() {
	file, err := os.Open("pacman.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	m := make(map[string]File)
	upgCntr := 0
	rmvCntr := 0
	instllCntr := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")
		for i, s := range line {
			if s == "installed" || s == "upgraded" || s == "removed" {
				fileName := line[i+1]
				fl := m[fileName]
				fl.pckgName = fileName
				if fl.removalDate == "" || fl.lastUpdtDate == "" {
					fl.removalDate = "-"
					fl.lastUpdtDate = "-"
				}
				if s == "installed" {
					instllCntr = instllCntr + 1
					fl.nUpdt = 0
					fl.installDate = line[i-3] + " " + line[i-2]
					fl.installDate = strings.Trim(fl.installDate, "[")
					fl.installDate = strings.Trim(fl.installDate, "]")
				}
				if s == "upgraded" {
					upgCntr = upgCntr + 1
					fl.nUpdt = fl.nUpdt + 1
					fl.lastUpdtDate = line[i-3] + " " + line[i-2]
					fl.lastUpdtDate = strings.Trim(fl.lastUpdtDate, "[")
					fl.lastUpdtDate = strings.Trim(fl.lastUpdtDate, "]")
				}
				if s == "removed" {
					rmvCntr = rmvCntr + 1
					fl.removalDate = line[i-3] + " " + line[i-2]
					fl.removalDate = strings.Trim(fl.removalDate, "]")
					fl.removalDate = strings.Trim(fl.removalDate, "[")
				}
				m[fileName] = fl
			}
		}
	}

	fmt.Println("Pacman Packages Report")
	fmt.Println("----------------------")
	fmt.Printf("- Installed files: %v\n", instllCntr)
	fmt.Printf("- Removed files: %v\n", rmvCntr)
	fmt.Printf("- Upgraded files: %v\n", upgCntr)
	fmt.Printf("- Current installed: %v\n", instllCntr-rmvCntr)

	for key, value := range m {
		fmt.Println("List of packages")
		fmt.Println("----------------")
		fmt.Printf("- Package Name        : %v\n", key)
		fmt.Printf("  - Install date      : %v\n", value.installDate)
		fmt.Printf("  - Last update date  : %v\n", value.lastUpdtDate)
		fmt.Printf("  - How many updates  : %v\n", value.nUpdt)
		fmt.Printf("  - Removal date      : %v\n", value.removalDate)
	}
}

