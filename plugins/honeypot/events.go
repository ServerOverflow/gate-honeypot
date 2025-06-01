package honeypot

import (
	"bytes"
	"encoding/json"
	"github.com/mrheinen/p0fclient"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type EventType int

const (
	Pinged EventType = iota
	Joined
	Left
)

type Event struct {
	Type            EventType `json:"type"`
	OperatingSystem string    `json:"operatingSystem"`
	Protocol        int       `json:"protocol"`
	SourceIp        string    `json:"sourceIp"`
	SourcePort      int       `json:"sourcePort"`
	TargetIp        string    `json:"targetIp"`
	TargetPort      int       `json:"targetPort"`
	Username        *string   `json:"username,omitempty"`
	Uuid            *string   `json:"uuid,omitempty"`
}

var (
	queue      = make(chan Event, 100)
	httpClient = &http.Client{Timeout: 5 * time.Second}
	endpoint   = os.Getenv("EVENTS_URL")
	p0fSocket  = os.Getenv("P0F_SOCKET_PATH")
	address    = os.Getenv("SERVER_ADDRESS")
	apiKey     = os.Getenv("API_KEY")
)

func StartEventSender() {
	go func() {
		for evt := range queue {
			sendEvent(evt)
		}
	}()
}

func sendEvent(evt Event) {
	cli := p0fclient.NewP0fClient(p0fSocket)
	err := cli.Connect()
	if err != nil {
		log.Printf("Failed to connect to p0f socket: %v\n", err)
		return
	}

	res, _ := cli.QueryIP(net.ParseIP(evt.SourceIp))
	evt.OperatingSystem = string(res.OsName[:]) + " " + string(res.OsFlavor[:])

	split := strings.Split(address, ":")
	port, _ := strconv.Atoi(split[1])
	evt.TargetIp = split[0]
	evt.TargetPort = port

	data, err := json.Marshal(evt)
	if err != nil {
		log.Printf("Failed to marshal event: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Failed to create request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to send event: %v\n", err)
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v\n", err)
		return
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode >= 300 {
		log.Printf("Failed with status code: %d\n", resp.StatusCode)
		log.Print(string(bodyBytes))
	}
}
