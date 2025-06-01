package honeypot

import (
	"context"
	"github.com/robinbraemer/event"
	"go.minekube.com/gate/pkg/edition/java/ping"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func id(s string) uuid.UUID {
	id, _ := uuid.Parse(s) // ignore error here
	return id
}

var players = []ping.SamplePlayer{
	{Name: "ghostbusterx", ID: id("eb0f2383-7f55-4b92-850c-a288d6a9f41a")},
	{Name: "twc_msmc_tz5", ID: id("f5caee1d-2a65-401c-9175-8205e8571e76")},
	{Name: "raininpaine", ID: id("eaa58dc9-e939-47b5-abf1-65c708b04f4f")},
	{Name: "q23w", ID: id("082040a4-5a05-4c23-84f1-6b4450a3455d")},
	{Name: "sovis", ID: id("2d321b2b-a0cb-45bb-b427-7861569ed3cc")},
	{Name: "xnfg", ID: id("88a94809-875a-4c32-a6b7-feb32c38de7d")},
	{Name: "alukyn", ID: id("01b99b3e-556c-4620-b151-b72ac4a7fc33")},
	{Name: "hannah_3", ID: id("846e95a1-9986-4a8f-b6c4-8d2fa0da9dea")},
	{Name: "graskiffer", ID: id("3cd415ae-d3ac-4cc4-8ba1-0eb07fe5f246")},
}

var Plugin = proxy.Plugin{
	Name: "Ping",
	Init: func(ctx context.Context, p *proxy.Proxy) error {
		event.Subscribe(p.Event(), 0, onPing)
		event.Subscribe(p.Event(), 0, onJoin)
		return nil
	},
}

func onPing(e *proxy.PingEvent) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	playersCopy := make([]ping.SamplePlayer, len(players))
	copy(playersCopy, players)

	r.Shuffle(len(playersCopy), func(i, j int) {
		playersCopy[i], playersCopy[j] = playersCopy[j], playersCopy[i]
	})

	n := 3
	if len(playersCopy) < n {
		n = len(playersCopy)
	}

	p := e.Ping()
	p.Players.Sample = append(p.Players.Sample, playersCopy[:n]...)

	split := strings.Split(e.Connection().RemoteAddr().String(), ":")
	port, _ := strconv.Atoi(split[1])

	sendEvent(Event{
		Type:       Pinged,
		Protocol:   int(e.Connection().Protocol()),
		SourceIp:   split[0],
		SourcePort: port,
	})
}

func onJoin(e *proxy.LoginEvent) {
	split := strings.Split(e.Player().RemoteAddr().String(), ":")
	port, _ := strconv.Atoi(split[1])
	username := e.Player().Username()
	uuid := e.Player().ID().String()

	sendEvent(Event{
		Type:       Joined,
		Protocol:   int(e.Player().Protocol()),
		SourceIp:   split[0],
		SourcePort: port,
		Username:   &username,
		Uuid:       &uuid,
	})
}

func onLeave(e *proxy.DisconnectEvent) {
	split := strings.Split(e.Player().RemoteAddr().String(), ":")
	port, _ := strconv.Atoi(split[1])
	username := e.Player().Username()
	uuid := e.Player().ID().String()

	sendEvent(Event{
		Type:       Left,
		Protocol:   int(e.Player().Protocol()),
		SourceIp:   split[0],
		SourcePort: port,
		Username:   &username,
		Uuid:       &uuid,
	})
}
