package main

import (
	"log"

	"github.com/jgivc/vapp/config"
	"github.com/jgivc/vapp/internal/app"
	"github.com/jgivc/vapp/pkg/logger"
)

const (
	defaultConfigFileName = "config.yml"
)

func main() {
	logger := logger.New()
	cfg, err := config.New(defaultConfigFileName)
	if err != nil {
		log.Fatal(err)
	}

	app.Run(cfg, logger)

}
