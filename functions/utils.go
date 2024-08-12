package functions

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"time"
)

func ReadFileLines(path string) []string {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("failed to get file: %v\n", err)
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Fatalf("failed to close file: %v\n", err)
		}
	}(file)

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

func GetRandomString(input []string) string {
	if len(input) == 0 {
		log.Fatalln("empty list detected")
	}

	rand.Seed(time.Now().UnixNano())
	return input[rand.Intn(len(input))]
}
