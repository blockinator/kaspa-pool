package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/onemorebsmith/kaspa-pool/src/kaspastratum"
	"gopkg.in/yaml.v2"
)

func main() {
	pwd, _ := os.Getwd()
	fullPath := path.Join(pwd, "config.yaml")
	log.Printf("loading config @ `%s`", fullPath)
	rawCfg, err := ioutil.ReadFile(fullPath)
	if err != nil {
		log.Printf("config file not found: %s", err)
		os.Exit(1)
	}
	cfg := kaspastratum.BridgeConfig{}
	if err := yaml.Unmarshal(rawCfg, &cfg); err != nil {
		log.Printf("failed parsing config file: %s", err)
		os.Exit(1)
	}

	flag.StringVar(&cfg.StratumPort, "stratum", cfg.StratumPort, "stratum port to listen on, default `:5555`")
	flag.StringVar(&cfg.RPCServer, "kaspa", cfg.RPCServer, "address of the kaspad node, default `localhost:16110`")
	flag.StringVar(&cfg.PromPort, "prom", cfg.PromPort, "address to serve prom stats, default `:2112`")
	flag.StringVar(&cfg.HealthCheckPort, "hcp", cfg.HealthCheckPort, `(rarely used) if defined will expose a health check on /readyz, default ""`)
	flag.Parse()

	log.Println("----------------------------------")
	log.Printf("initializing bridge")
	log.Printf("\tkaspad:        %s", cfg.RPCServer)
	log.Printf("\tstratum:       %s", cfg.StratumPort)
	log.Printf("\tprom:          %s", cfg.PromPort)
	log.Printf("\thealth check:  %s", cfg.HealthCheckPort)
	log.Println("----------------------------------")

	if err := kaspastratum.ListenAndServe(cfg); err != nil {
		log.Println(err)
	}
}
