package functions

import (
	"fmt"
	http "github.com/bogdanfinn/fhttp"
	tlsclient "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"io"
	"log"
	"strings"
)

type BodyResponse struct {
	Message    interface{} `json:"message"`
	RetryAfter interface{} `json:"retry_after"`
}

// GetHeaders TODO: flexible x-super-properties (getting latest build)
func getHeaders(token, cookie, context *string) map[string]string {
	headers := make(map[string]string)

	headers["accept"] = "*/*"
	headers["accept-language"] = "en-US"
	if token != nil {
		headers["authorization"] = *token
	}
	headers["content-type"] = "application/json"
	if cookie != nil {
		headers["cookie"] = *cookie
	}
	headers["origin"] = "https://canary.discord.com"
	headers["priority"] = "u=1, i"
	headers["referer"] = "https://canary.discord.com/channels/@me"
	headers["sec-ch-ua"] = `"Not/A)Brand";v="8", "Chromium";v="126"`
	headers["sec-ch-ua-mobile"] = "?0"
	headers["sec-ch-ua-platform"] = `"Windows"`
	headers["sec-fetch-dest"] = "empty"
	headers["sec-fetch-mode"] = "cors"
	headers["sec-fetch-site"] = "same-origin"
	headers["user-agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) discord/1.0.427 Chrome/126.0.6478.127 Electron/31.2.1 Safari/537.36"
	if context != nil {
		headers["x-context-properties"] = *context
	}
	headers["x-debug-options"] = "bugReporterEnabled"
	headers["x-discord-locale"] = "pl"
	headers["x-discord-timezone"] = "Europe/Warsaw"
	headers["x-super-properties"] = `eyJvcyI6IldpbmRvd3MiLCJicm93c2VyIjoiRGlzY29yZCBDbGllbnQiLCJyZWxlYXNlX2NoYW5uZWwiOiJjYW5hcnkiLCJjbGllbnRfdmVyc2lvbiI6IjEuMC40MjciLCJvc192ZXJzaW9uIjoiMTAuMC4xOTA0NSIsIm9zX2FyY2giOiJ4NjQiLCJhcHBfYXJjaCI6Ing2NCIsInN5c3RlbV9sb2NhbGUiOiJlbi1VUyIsImJyb3dzZXJfdXNlcl9hZ2VudCI6Ik1vemlsbGEvNS4wIChXaW5kb3dzIE5UIDEwLjA7IFdpbjY0OyB4NjQpIEFwcGxlV2ViS2l0LzUzNy4zNiAoS0hUTUwsIGxpa2UgR2Vja28pIGRpc2NvcmQvMS4wLjQyNyBDaHJvbWUvMTI2LjAuNjQ3OC4xMjcgRWxlY3Ryb24vMzEuMi4xIFNhZmFyaS81MzcuMzYiLCJicm93c2VyX3ZlcnNpb24iOiIzMS4yLjEiLCJjbGllbnRfYnVpbGRfbnVtYmVyIjozMTc0NDMsIm5hdGl2ZV9idWlsZF9udW1iZXIiOjUwNzE5LCJjbGllbnRfZXZlbnRfc291cmNlIjpudWxsfQ==`

	return headers
}

func BuildClient(method, url string, body io.Reader, token, cookie, context *string) *http.Response {
	options := []tlsclient.HttpClientOption{
		tlsclient.WithClientProfile(profiles.Chrome_124),
		tlsclient.WithRandomTLSExtensionOrder(),
	}

	client, err := tlsclient.NewHttpClient(tlsclient.NewNoopLogger(), options...)
	if err != nil {
		log.Fatalf("failed to create client: %v\n", err)
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Fatalf("failed to set request: %v\n", err)
	}

	for key, value := range getHeaders(token, cookie, context) {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to send request: %v\n", err)
	}
	return resp
}

func GetCookies() string {
	var cookieStrings []string

	resp := BuildClient(http.MethodGet, "https://canary.discord.com", nil, nil, nil, nil)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("failed to close client: %v\n", err)
		}
	}(resp.Body)

	cookies := resp.Cookies()
	for _, cookie := range cookies {
		cookieStrings = append(cookieStrings, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}

	return strings.Join(cookieStrings, "; ")
}
