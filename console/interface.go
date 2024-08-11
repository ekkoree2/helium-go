package console

import (
	"fmt"
	"golang.org/x/term"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

func stripColors(text string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(text, "")
}

func center(text string) string {
	fd := int(os.Stdout.Fd())

	width, _, err := term.GetSize(fd)
	if err != nil {
		log.Fatal(err)
	}

	plain := stripColors(text)

	padding := (width - len(plain)) / 2

	if padding < 0 {
		padding = 0
	}

	return fmt.Sprintf("%s%s", strings.Repeat(" ", padding), text)
}

func DisplayArt() {
	art := []string{
		`   __       ___          `,
		`  / /  ___ / (_)_ ____ _ `,
		` / _ \/ -_) / / // /  ' \`,
		`/_//_/\__/_/_/\_,_/_/_/_/`,
		`                         `,
	}

	for _, line := range art {
		render := fmt.Sprintf("%s%s", Colors["light_blue"], center(line))
		fmt.Println(render)
	}
}

func DisplayMenu() {
	functions := GetFunctions()
	keys := make([]int64, 0, len(functions))
	var lines []string

	for key := range functions {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	columns := 4
	rows := (len(keys) + columns - 1) / columns

	for i := 0; i < rows; i++ {
		line := ""

		for j := 0; j < columns; j++ {
			num := i + j*rows
			if num < len(keys) {
				line += fmt.Sprintf("%s%02d: %s%-20s", Colors["blue"], keys[num], Colors["white"], functions[keys[num]])
			}
		}
		lines = append(lines, line)
	}

	for _, line := range lines {
		fmt.Println(center(line))
	}
}

func ClearConsole() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	default:
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}
