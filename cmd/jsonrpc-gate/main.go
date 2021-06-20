package main

import (
	"flag"

	"github.com/nspcc-dev/neofs-node/cmd/neofs-node/config"
)

func main() {
	configFile := flag.String("config", "", "path to config")
	flag.Parse()

	var p config.Prm

	appCfg := config.New(p,
		config.WithConfigFile(*configFile),
	)

	serveGate(appCfg)
}
