package functions

import (
	"encoding/json"
	"fmt"
	http "github.com/bogdanfinn/fhttp"
	"io"
	"log"
	"main/console"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func handleResponse(resp *http.Response) ([]byte, error) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("failed to close response body: %v", err)
		}
	}(resp.Body)

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return bodyText, nil
}

func CheckGuild(token, guild string) bool {
	resp := BuildClient(http.MethodGet, "https://canary.discord.com/api/v9/guilds/"+guild, nil, &token, nil, nil)

	bodyText, err := handleResponse(resp)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode == http.StatusOK {
		return true
	}

	var prop BodyResponse
	err = json.Unmarshal(bodyText, &prop)
	if err != nil {
		log.Fatalf("failed to unmarshal: %v\n", err)
	}

	message := prop.Message.(string)
	if message == "" {
		message = "unknown error"
	}
	console.DisplayText("FATAL", console.Colors["red"], token[:20], message)
	return false
}

func CheckChannel(token, channel string) bool {
	resp := BuildClient(http.MethodGet, "https://canary.discord.com/api/v9/channels/"+channel+"/messages?limit=50", nil, &token, nil, nil)

	bodyText, err := handleResponse(resp)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode == http.StatusOK {
		return true
	}

	var prop BodyResponse
	err = json.Unmarshal(bodyText, &prop)
	if err != nil {
		log.Fatalf("failed to unmarshal: %v\n", err)
	}

	message := prop.Message.(string)
	if message == "" {
		message = "unknown error"
	}
	console.DisplayText("FATAL", console.Colors["red"], token[:20], message)
	return false
}

func getRandomMember(guildID string, count int) string {
	var message string
	members := ReadFileLines(guildID + ".txt")

	if len(members) == 0 {
		log.Fatalf(console.Colors["white"]+"no members found in guild file: %s.txt", guildID)
	}

	if count > len(members) {
		count = len(members)
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(members), func(i, j int) { members[i], members[j] = members[j], members[i] })

	for i := 0; i < count; i++ {
		message += fmt.Sprintf(" <@!%s>", members[i])
	}
	return message
}

func SendMessage(token, message, channel string, guildID, cookie *string, count *int) {
	for {
		content := message

		if guildID != nil {
			content += getRandomMember(*guildID, *count)
		}

		jsonData := strings.NewReader(fmt.Sprintf(`{"mobile_network_type":"unknown","content":"%s","nonce":"","tts":false,"flags":0}`, content))
		resp := BuildClient(http.MethodPost, fmt.Sprintf("https://canary.discord.com/api/v9/channels/%s/messages", channel), jsonData, &token, cookie, nil)

		bodyText, err := handleResponse(resp)
		if err != nil {
			log.Fatal(err)
		}

		var prop BodyResponse
		err = json.Unmarshal(bodyText, &prop)
		if err != nil {
			log.Fatalf("failed to unmarshal: %v\n", err)
		}

		switch resp.StatusCode {
		case http.StatusOK:
			console.DisplayText("SENT", console.Colors["green"], token[:20], "")
		case http.StatusTooManyRequests:
			delay, err := strconv.ParseFloat(fmt.Sprintf("%v", prop.RetryAfter), 64)
			if err != nil {
				log.Fatalf("failed to parse retry delay: %v\n", err)
			}
			time.Sleep(time.Duration(delay) * time.Millisecond)
		default:
			console.DisplayText("FATAL", console.Colors["red"], token[:20], prop.Message.(string))
			return
		}
	}
}
