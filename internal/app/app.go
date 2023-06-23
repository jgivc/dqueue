package app

import (
	"github.com/jgivc/vapp/config"
	"github.com/jgivc/vapp/pkg/logger"
)

func Run(cfg *config.Config, logger logger.Logger) {
	logger.Info("App start")
	panic("not implemented")
}
