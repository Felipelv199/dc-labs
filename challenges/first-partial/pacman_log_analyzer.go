package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
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
				if line[i-1] == "[ALPM]" && s == "installed" {
					instllCntr = instllCntr + 1
					fl.nUpdt = 0
					fl.installDate = line[i-3] + " " + line[i-2]
					fl.installDate = strings.Trim(fl.installDate, "[")
					fl.installDate = strings.Trim(fl.installDate, "]")
					fl.removalDate = "-"
					fl.lastUpdtDate = "-"
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

	f, err := os.Create("packages_report.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	l, err := f.WriteString("Pacman Packages Report\n")
	l, err = f.WriteString("----------------------\n")
	s := strconv.FormatInt(int64(instllCntr), 10)
	l, err = f.WriteString("- Installed files: " + s + "\n")
	s = strconv.FormatInt(int64(rmvCntr), 10)
	l, err = f.WriteString("- Removed files: " + s + "\n")
	s = strconv.FormatInt(int64(upgCntr), 10)
	l, err = f.WriteString("- Upgraded files: " + s + "\n")
	s = strconv.FormatInt(int64(instllCntr-rmvCntr), 10)
	l, err = f.WriteString("- Current installed: " + s + "\n")
	l, err = f.WriteString("\nList of packages\n")
	l, err = f.WriteString("----------------\n")
	for key, value := range m {
		l, err = f.WriteString("- Package Name        : " + key + "\n")
		l, err = f.WriteString("  - Install date      : " + value.installDate + "\n")
		l, err = f.WriteString("  - Last update date  : " + value.lastUpdtDate + "\n")
		l, err = f.WriteString("  - How many updates  : " + strconv.FormatInt(int64(value.nUpdt), 10) + "\n")
		l, err = f.WriteString("  - Removal date      : " + value.removalDate + "\n")

	}

	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}
	fmt.Println(l, "bytes written successfully")
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}

