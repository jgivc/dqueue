package main

import (
	"flag"
	"log"

	"github.com/jgivc/dqueue/config"
	"github.com/jgivc/dqueue/internal/app"
	"github.com/jgivc/dqueue/pkg/logger"
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
