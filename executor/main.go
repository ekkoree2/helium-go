package executor

import (
	"log"
	"main/console"
	"main/functions"
	"strconv"
)

func Main() {
	console.ClearConsole()
	console.DisplayArt()
	console.DisplayMenu()
	tokens := functions.ReadFileLines("tokens.txt")

	result := console.Prompt("option", false)
	parse, err := strconv.Atoi(result)
	if err != nil {
		log.Fatalf("failed to parse int: %v\n", err)
	}
	module := console.GetFunctions()
	strModule := module[int64(parse)]
	ExecuteFunction(strModule, tokens)
}
