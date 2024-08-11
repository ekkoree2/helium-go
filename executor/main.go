package executor

import (
	"bufio"
	"log"
	"main/console"
	"os"
	"strconv"
)

func readFileLines(path string) []string {
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

func Main() {
	console.ClearConsole()
	console.DisplayArt()
	console.DisplayMenu()
	tokens := readFileLines("tokens.txt")

	result := console.Prompt("option", false)
	parse, err := strconv.Atoi(result)
	if err != nil {
		log.Fatalf("failed to parse int: %v\n", err)
	}
	module := console.GetFunctions()
	strModule := module[int64(parse)]
	ExecuteFunction(strModule, tokens)
}
