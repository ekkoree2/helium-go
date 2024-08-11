package functions

import (
	"encoding/json"
	"fmt"
	http "github.com/bogdanfinn/fhttp"
	"io"
	"log"
	"main/console"
	"strings"
)

func LeaveServer(token, guild, cookie string) {
	data := strings.NewReader(`{"lurking":false}`)
	resp := BuildClient(http.MethodDelete, fmt.Sprintf("https://canary.discord.com/api/v9/users/@me/guilds/%s", guild), data, &token, &cookie, nil)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("failed to close client: %v\n", err)
		}
	}(resp.Body)
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	switch resp.StatusCode {
	case http.StatusNoContent:
		console.DisplayText("LEFT", console.Colors["green"], token[:20], guild)
	case http.StatusTooManyRequests:
		console.DisplayText("RATE LIMITED", console.Colors["magenta"], token[:20], guild)
	default:
		var prop BodyResponse
		err = json.Unmarshal(bodyText, &prop)
		if err != nil {
			log.Fatalf("failed to unmarshal: %v\n", err)
		}
		console.DisplayText("FATAL", console.Colors["red"], token[:20], prop.Message.(string))
	}
}
