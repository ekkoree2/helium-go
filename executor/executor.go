package executor

import (
	"bufio"
	"fmt"
	"log"
	"main/console"
	"main/functions"
	"os"
	"strconv"
	"sync"
)

func ExecuteFunction(module string, tokens []string) {
	console.ClearConsole()
	console.DisplayArt()

	stringChannels := console.Prompt("workers amount", false)

	var wg sync.WaitGroup
	numChannels, err := strconv.Atoi(stringChannels)
	if err != nil {
		log.Fatalf("failed to parse workers int: %v\n", err)
	}

	jobs := make(chan []interface{}, len(tokens))
	cookie := functions.GetCookies()

	for i := 0; i < numChannels; i++ {
		wg.Add(1)
		switch module {
		case "joiner":
			go Worker(&wg, jobs, functions.JoinServer)
		case "leaver":
			go Worker(&wg, jobs, functions.LeaveServer)
		}
	}

	switch module {
	case "joiner":
		invite := console.Prompt("invite", false)
		properties := functions.GetProperties(invite)
		console.ClearConsole()
		console.DisplayArt()
		for _, token := range tokens {
			jobs <- []interface{}{token, invite, cookie, properties}
		}

	case "leaver":
		guild := console.Prompt("guild id", false)
		console.ClearConsole()
		console.DisplayArt()
		for _, token := range tokens {
			jobs <- []interface{}{token, guild, cookie}
		}
	}

	close(jobs)
	wg.Wait()

	fmt.Println("\n~/> press enter to continue")
	_, err = bufio.NewReader(os.Stdin).ReadBytes('\n')
	if err != nil {
		return
	}
	Main()
}
