package main

import (
	"flag"
	"log"

	"github.com/jgivc/vapp/config"
	"github.com/jgivc/vapp/internal/app"
	"github.com/jgivc/vapp/pkg/logger"
)

const (
	defaultConfigFileName = "config.yml"
)

func main() {
	configFileName := flag.String("c", defaultConfigFileName, "Path to config file.")
	flag.Parse()

	logger := logger.New()
	cfg, err := config.New(*configFileName)
	if err != nil {
		log.Fatal(err)
	}

	app.Run(cfg, logger)

}
