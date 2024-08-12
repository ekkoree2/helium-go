package functions

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var users []map[string]interface{}
var usersMutex sync.Mutex

func rangeCorrector(ranges [][]int) [][]int {
	if !containsRange(ranges, []int{0, 99}) {
		ranges = append([][]int{{0, 99}}, ranges...)
	}
	return ranges
}

func getRanges(index, multiplier, memberCount int) [][]int {
	initialNum := index * multiplier
	rangesList := [][]int{{initialNum, initialNum + 99}}
	if memberCount > initialNum+99 {
		rangesList = append(rangesList, []int{initialNum + 100, initialNum + 199})
	}
	return rangeCorrector(rangesList)
}

func parseGuildMemberListUpdate(response map[string]interface{}) map[string]interface{} {
	memberdata := map[string]interface{}{
		"online_count":  response["d"].(map[string]interface{})["online_count"],
		"member_count":  response["d"].(map[string]interface{})["member_count"],
		"id":            response["d"].(map[string]interface{})["id"],
		"guild_id":      response["d"].(map[string]interface{})["guild_id"],
		"hoisted_roles": response["d"].(map[string]interface{})["groups"],
		"types":         []interface{}{},
		"locations":     []interface{}{},
		"updates":       []interface{}{},
	}
	ops, ok := response["d"].(map[string]interface{})["ops"].([]interface{})
	if !ok {
		log.Println("Expected ops to be []interface{}, but got something else")
		return nil
	}

	for _, chunk := range ops {
		memberdata["types"] = append(memberdata["types"].([]interface{}), chunk.(map[string]interface{})["op"])
		if chunk.(map[string]interface{})["op"] == "SYNC" || chunk.(map[string]interface{})["op"] == "INVALIDATE" {
			memberdata["locations"] = append(memberdata["locations"].([]interface{}), chunk.(map[string]interface{})["range"])
			if chunk.(map[string]interface{})["op"] == "SYNC" {
				memberdata["updates"] = append(memberdata["updates"].([]interface{}), chunk.(map[string]interface{})["items"])
			} else {
				memberdata["updates"] = append(memberdata["updates"].([]interface{}), []interface{}{})
			}
		} else if chunk.(map[string]interface{})["op"] == "INSERT" || chunk.(map[string]interface{})["op"] == "UPDATE" || chunk.(map[string]interface{})["op"] == "DELETE" {
			memberdata["locations"] = append(memberdata["locations"].([]interface{}), chunk.(map[string]interface{})["index"])
			if chunk.(map[string]interface{})["op"] == "DELETE" {
				memberdata["updates"] = append(memberdata["updates"].([]interface{}), []interface{}{})
			} else {
				memberdata["updates"] = append(memberdata["updates"].([]interface{}), chunk.(map[string]interface{})["item"])
			}
		}
	}

	return memberdata
}

type DiscordSocket struct {
	Token         string
	GuildID       string
	ChannelID     string
	SocketHeaders map[string]string
	Conn          *websocket.Conn
	EndScraping   bool
	Guilds        map[string]map[string]interface{}
	Members       map[string]interface{}
	Ranges        [][]int
	LastRange     int
	PacketsRecv   int
	Mutex         sync.Mutex
}

func (ds *DiscordSocket) run() {
	ds.runForever()
}

func (ds *DiscordSocket) scrapeUsers() {
	if !ds.EndScraping {
		rangesJSON, err := json.Marshal(ds.Ranges)
		if err != nil {
			log.Fatal(err)
			return
		}

		requestJSON1 := `{"op":14,"d":{"guild_id":"` + ds.GuildID + `","typing":true,"activities":true,"threads":true,"channels":{"` + ds.ChannelID + `":` + string(rangesJSON) + `}}}`
		ds.send(requestJSON1)
		requestJSON2 := `{"op":14,"d":{"guild_id":"` + ds.GuildID + `","typing":true,"activities":true,"threads":true,"channels":{"` + ds.ChannelID + `":[[0, 99], [100, 199]]}}}`
		ds.send(requestJSON2)
	}
}

func (ds *DiscordSocket) sockOpen() {
	ds.send(fmt.Sprintf(`{"op":2,"d":{"token":"%s","capabilities":125,"properties":{"os":"Windows","browser":"Firefox","device":"","system_locale":"it-IT","browser_user_agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0) Gecko/20100101 Firefox/94.0","browser_version":"94.0","os_version":"10","referrer":"","referring_domain":"","referrer_current":"","referring_domain_current":"","release_channel":"stable","client_build_number":103981,"client_event_source":null},"presence":{"status":"online","since":0,"activities":[],"afk":false},"compress":false,"client_state":{"guild_hashes":{},"highest_last_message_id":"0","read_state_version":0,"user_guild_settings_version":-1,"user_settings_version":-1}}}`, ds.Token))
}

func (ds *DiscordSocket) heartbeatThread(interval time.Duration) {
	for {
		ds.send(fmt.Sprintf(`{"op":1,"d":%d}`, ds.PacketsRecv))
		time.Sleep(interval)
	}
}

func (ds *DiscordSocket) sockMessage(message []byte) {
	var decoded map[string]interface{}
	err := json.Unmarshal(message, &decoded)
	if err != nil || decoded == nil {
		log.Println(err)
	}

	ds.Mutex.Lock()
	defer ds.Mutex.Unlock()

	if decoded["op"].(float64) != 11 {
		ds.PacketsRecv++
	}

	switch decoded["op"].(float64) {
	case 10:
		go ds.heartbeatThread(time.Duration(decoded["d"].(map[string]interface{})["heartbeat_interval"].(float64)) * time.Millisecond)
	case 0:
		switch decoded["t"].(string) {
		case "READY":
			for _, guild := range decoded["d"].(map[string]interface{})["guilds"].([]interface{}) {
				guildID := guild.(map[string]interface{})["id"].(string)
				ds.Guilds[guildID] = map[string]interface{}{
					"member_count": guild.(map[string]interface{})["member_count"].(float64),
				}
			}
		case "READY_SUPPLEMENTAL":
			if decoded["t"].(string) == "READY_SUPPLEMENTAL" {
				guildID := ds.GuildID
				if guildID == "" {
					log.Println("Error: GuildID is empty")
					return
				}

				guild := ds.Guilds[guildID]
				memberCount, _ := guild["member_count"].(float64)

				ds.Ranges = getRanges(0, 100, int(memberCount))
				ds.scrapeUsers()
			}
		case "GUILD_MEMBER_LIST_UPDATE":
			parsed := parseGuildMemberListUpdate(decoded)
			if parsed["guild_id"].(string) == ds.GuildID && (containsType(parsed["types"].([]interface{}), "SYNC") || containsType(parsed["types"].([]interface{}), "UPDATE")) {
				for elem, index := range parsed["types"].([]interface{}) {
					if index == "SYNC" {
						if len(parsed["updates"].([]interface{})[elem].([]interface{})) == 0 {
							ds.EndScraping = true
							break
						}
						for _, item := range parsed["updates"].([]interface{})[elem].([]interface{}) {
							if item.(map[string]interface{})["member"] != nil {
								mem := item.(map[string]interface{})["member"].(map[string]interface{})
								obj := map[string]interface{}{
									"tag": mem["user"].(map[string]interface{})["username"].(string) + "#" + mem["user"].(map[string]interface{})["discriminator"].(string),
									"id":  mem["user"].(map[string]interface{})["id"].(string),
								}
								usersMutex.Lock()
								if !userExists(obj["id"].(string)) {
									users = append(users, obj)
								}
								usersMutex.Unlock()
							}
						}
					} else if index == "UPDATE" {
						updates, ok := parsed["updates"].([]interface{})
						if !ok {
							return
						}

						for _, update := range updates {
							updateItems, ok := update.([]interface{})
							if !ok {
								continue
							}

							for _, item := range updateItems {
								member, ok := item.(map[string]interface{})["member"].(map[string]interface{})
								if !ok || member == nil {
									continue
								}

								username, _ := member["user"].(map[string]interface{})["username"].(string)
								discriminator, _ := member["user"].(map[string]interface{})["discriminator"].(string)
								id, _ := member["user"].(map[string]interface{})["id"].(string)

								obj := map[string]interface{}{
									"tag": username + "#" + discriminator,
									"id":  id,
								}

								usersMutex.Lock()
								if !userExists(obj["id"].(string)) {
									users = append(users, obj)
								}
								usersMutex.Unlock()
							}
						}
					}
					ds.LastRange++
					ds.Ranges = getRanges(ds.LastRange, 100, int(ds.Guilds[ds.GuildID]["member_count"].(float64)))
					time.Sleep(350 * time.Millisecond)
					ds.scrapeUsers()
				}
			}
			if ds.EndScraping {
				err := ds.Conn.Close()
				if err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}

func userExists(userID string) bool {
	for _, user := range users {
		if user["id"] == userID {
			return true
		}
	}
	return false
}

func containsRange(ranges [][]int, target []int) bool {
	for _, r := range ranges {
		if r[0] == target[0] && r[1] == target[1] {
			return true
		}
	}
	return false
}

func containsType(types []interface{}, target string) bool {
	for _, t := range types {
		if t.(string) == target {
			return true
		}
	}
	return false
}

func (ds *DiscordSocket) runForever() {
	url := "wss://gateway.discord.gg/?encoding=utf&v=9"
	conn, _, err := websocket.DefaultDialer.Dial(url, http.Header{"Sec-WebSocket-Extensions": {"permessage-deflate; client_max_window_bits"}})
	if err != nil {
		log.Println(err)
		return
	}
	ds.Conn = conn

	go ds.sockOpen()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			go ds.sockMessage(message)
		}
	}()

	<-done
}

func (ds *DiscordSocket) send(message string) {
	err := ds.Conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Println(err)
		return
	}
}

func Scrape(token, guildID, channelID string) {
	ds := &DiscordSocket{
		Token:         token,
		GuildID:       guildID,
		ChannelID:     channelID,
		SocketHeaders: map[string]string{"Accept-Encoding": "gzip, deflate, br", "Accept-Language": "en-US,en;q=0.9", "Cache-Control": "no-cache", "Pragma": "no-cache", "Sec-WebSocket-Extensions": "permessage-deflate; client_max_window_bits", "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.1.2 Safari/605.1.15"},
		Guilds:        make(map[string]map[string]interface{}),
		Members:       make(map[string]interface{}),
		Ranges:        [][]int{{0, 0}},
		LastRange:     0,
		PacketsRecv:   0,
		Mutex:         sync.Mutex{},
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		ds.run()
	}()

	<-done

	userIDs := make([]string, len(users))
	for i, member := range users {
		userIDs[i] = member["id"].(string)
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

	for _, userID := range userIDs {
		_, err := file.WriteString(userID + "\n")
		if err != nil {
			log.Fatalf("failed to write to file: %v\n", err)
		}
	}
}
