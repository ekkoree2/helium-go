package console

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var Colors = map[string]string{
	"green":      "\033[32m",
	"red":        "\033[31m",
	"yellow":     "\033[33m",
	"magenta":    "\033[35m",
	"blue":       "\033[34m",
	"cyan":       "\033[36m",
	"gray":       "\033[90m",
	"white":      "\033[97m",
	"pink":       "\033[95m",
	"light_blue": "\033[94m",
}

func Prompt(text string, ask bool) string {
	response := fmt.Sprintf("%s[%s%s%s", Colors["white"], Colors["light_blue"], text, Colors["white"])

	if ask {
		response += fmt.Sprintf("? %s(y/n)%s]: ", Colors["gray"], Colors["white"])
	} else {
		response += "]: "
	}
	fmt.Print(response)

	reader := bufio.NewReader(os.Stdin)
	inputText, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	return strings.TrimSpace(inputText)
}

func DisplayText(text string, color string, token interface{}, log string) {
	response := fmt.Sprintf("%s[%s]%s ", Colors["cyan"], time.Now().Format("15:04:05"), Colors["white"])

	if text != "" {
		response += fmt.Sprintf("%s[%s]%s ", color, text, Colors["white"])
	}
	if token != nil {
		response += fmt.Sprintf("%v", token)
	}
	if log != "" {
		response += fmt.Sprintf(" %s(%s)%s", Colors["gray"], log, Colors["white"])
	}
	fmt.Println(response)
}
