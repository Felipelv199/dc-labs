package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func main() {
	bucket := flag.String("bucket", "s", "the bucket url")
	flag.Parse()
	bucketName := string(*bucket)
	url := "https://" + bucketName + ".s3.amazonaws.com"
	resp, err := http.Get(url)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	html, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	content := string(html)
	contentSplited := strings.Split(content, "</Contents>")
	numberOfFiles := 0
	directories := make(map[string]int)
	fileTypes := make(map[string]int)

	for _, s := range contentSplited {
		sSplited := strings.Split(s, "<Contents>")

		if len(sSplited) > 1 {
			key := strings.Split(sSplited[1], "<Key>")
			keySplited := strings.Split(key[1], "</Key>")

			if strings.Contains(keySplited[0], ".") {
				fileRoute := strings.Split(keySplited[0], ".")
				fileTypes[fileRoute[len(fileRoute)-1]] = fileTypes[fileRoute[len(fileRoute)-1]] + 1
				numberOfFiles = numberOfFiles + 1

				if strings.Contains(fileRoute[0], "/") {
					directory := strings.Split(fileRoute[0], "/")
					directories[directory[len(directory)-2]] = directories[directory[len(directory)-2]] + 1
				}
			} else {
				if strings.Contains(keySplited[0], " ") == false {
					directory := strings.Split(keySplited[0], "/")
					directories[directory[0]] = directories[directory[0]] + 1
				}
			}
		}
	}

	fmt.Println("AWS S3 Explorer")
	fmt.Printf("Bucket Name            : %v\n", bucketName)
	fmt.Printf("Number of objects      : %v\n", numberOfFiles)
	fmt.Printf("Number of directories  : %v\n", len(directories))
	fmt.Printf("Extensions             : ")
	counter := 0

	for i, s := range fileTypes {
		counter = counter + 1
		if counter < len(fileTypes) {
			fmt.Printf("%v(%v), ", i, s)
		} else {
			fmt.Printf("%v(%v)\n", i, s)
		}
	}
}

