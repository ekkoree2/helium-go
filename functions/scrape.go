package functions

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

type Hello struct {
	HeartbeatInterval int    `json:"heartbeat_interval"`
	Op                int    `json:"op"`
	T                 string `json:"t"`
}

type Identify struct {
	Op int `json:"op"`
	D  struct {
		Token      string `json:"token"`
		Properties struct {
			OS      string `json:"$os"`
			Browser string `json:"$browser"`
			Device  string `json:"$device"`
		} `json:"properties"`
	} `json:"d"`
}

func sendScrapeMessage(guildID string, channelID string, ws *websocket.Conn, r int) error {
	payload := map[string]interface{}{
		"op": 14,
		"d": map[string]interface{}{
			"guild_id":            guildID,
			"typing":              true,
			"threads":             true,
			"activities":          true,
			"members":             []interface{}{},
			"channels":            map[string][][]int{channelID: {{r * 100, r*100 + 100 - 1}}},
			"thread_member_lists": []interface{}{},
		},
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return ws.WriteMessage(websocket.TextMessage, jsonData)
}

func connect(token string, guildID string, channelID string) ([]string, error) {
	ws, _, err := websocket.DefaultDialer.Dial("wss://gateway.discord.gg/?v=9&encoding=json", nil)
	if err != nil {
		return nil, err
	}
	defer func(ws *websocket.Conn) {
		err := ws.Close()
		if err != nil {
			log.Fatalf("failed to close ws conn: %v\n", err)
		}
	}(ws)

	var users []string
	first := false
	totalRanges := 0
	rangesScraped := 0

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			return nil, err
		}

		var hello Hello
		err = json.Unmarshal(message, &hello)
		if err != nil {
			return nil, err
		}

		if hello.Op == 10 {
			if err := sendIdentify(token, ws); err != nil {
				return nil, err
			}
		}

		if hello.T == "READY" {
			if err := sendScrapeMessage(guildID, channelID, ws, 0); err != nil {
				return nil, err
			}
		}

		if hello.T == "GUILD_MEMBER_LIST_UPDATE" {
			var jsonData map[string]interface{}
			if err := json.Unmarshal(message, &jsonData); err != nil {
				log.Printf("failed to parse json: %v", err)
				continue
			}

			data := jsonData["d"].(map[string]interface{})

			if onlineCountVal, ok := data["online_count"]; ok && onlineCountVal != nil {
				if onlineCountFloat, ok := onlineCountVal.(float64); ok {
					onlineCount := int(onlineCountFloat)
					totalRanges = (onlineCount + 99) / 100

					if !first {
						for i := 0; i <= totalRanges; i++ {
							time.Sleep(time.Second)
							if err := sendScrapeMessage(guildID, channelID, ws, i); err != nil {
								return nil, err
							}
						}
						first = true
					}
				}
			}

			if opsVal, ok := data["ops"]; ok && opsVal != nil {
				for _, ops := range opsVal.([]interface{}) {
					op := ops.(map[string]interface{})
					if op["op"] == "SYNC" {
						for _, member := range op["items"].([]interface{}) {
							memberMap := member.(map[string]interface{})
							if user, exists := memberMap["member"]; exists && user != nil {
								userID := extractUserID(user)
								if userID != "" {
									users = append(users, userID)
								}
							}
						}
						rangesScraped++
					}
				}
			}

			if rangesScraped >= totalRanges {
				fmt.Printf("Scraped %d users from %s\n", len(users), guildID)
				break
			}
		}
	}
	return users, nil
}

func extractUserID(data interface{}) string {
	if dataMap, ok := data.(map[string]interface{}); ok {
		if memberData, ok := dataMap["user"]; ok {
			if memberDataMap, ok := memberData.(map[string]interface{}); ok {
				if id, exists := memberDataMap["id"]; exists && id != nil {
					if idStr, ok := id.(string); ok {
						return idStr
					}
				}
			}
		}
	}
	return ""
}

func sendIdentify(token string, ws *websocket.Conn) error {
	identify := Identify{
		Op: 2,
	}
	identify.D.Token = token
	identify.D.Properties.OS = "windows"
	identify.D.Properties.Browser = "Discord"
	identify.D.Properties.Device = "desktop"

	return ws.WriteJSON(identify)
}

func Scrape(token, guildID, channelID string) {
	users, err := connect(token, guildID, channelID)
	if err != nil {
		log.Fatalf("failed to scrape: %v", err)
	}

	file, err := os.Create(guildID + ".txt")
	if err != nil {
		log.Fatalf("failed to create file: %v\n", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("failed to close file: %v\n", err)
		}
	}(file)

	for _, userID := range users {
		_, err := file.WriteString(userID + "\n")
		if err != nil {
			log.Fatalf("failed to write to file: %v\n", err)
		}
	}
}
