package main

import (
	"github.com/dianpeng/hi-doctor/config"
	"github.com/dianpeng/hi-doctor/dvar"
	"github.com/dianpeng/hi-doctor/server"
	"github.com/dianpeng/hi-doctor/trigger"

	_ "net/http/pprof"

	"flag"
	"fmt"
	"os"
)

var configPath = flag.String("config", "./hi-doctor.yaml", "specify configuration file path")

func bailout(msg string) {
	fmt.Fprintf(os.Stderr, "%s", msg)
	os.Exit(-1)
}

func main() {
	flag.Parse()
	if cfg, err := config.LoadConfigFile(*configPath); err != nil {
		bailout(fmt.Sprintf("cannot load configuration file, %s", err))
	} else {
		trigger.Start()
		assets, err := dvar.PopulateAssetsMap(cfg.Assets)
		if err != nil {
			bailout(fmt.Sprintf("%s", err))
		}
		server.StartServer(cfg.ServiceDiscovery, assets, cfg.ServerAddress)
	}
}
