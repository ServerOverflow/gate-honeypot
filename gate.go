package main

import (
	"github.com/minekube/gate-plugin-template/plugins/honeypot"
	"go.minekube.com/gate/cmd/gate"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func main() {
	proxy.Plugins = append(
		proxy.Plugins,
		honeypot.Plugin,
	)

	honeypot.StartEventSender()
	gate.Execute()
}
