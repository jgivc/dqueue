package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jgivc/vapp/config"
	"github.com/jgivc/vapp/internal/adapter"
	"github.com/jgivc/vapp/internal/handler"
	"github.com/jgivc/vapp/internal/service"
	"github.com/jgivc/vapp/pkg/ami"
	"github.com/jgivc/vapp/pkg/logger"
)

func Run(cfg *config.Config, logger logger.Logger) {
	logger.Info("msg", "App start")
	ctx, cancel := context.WithCancel(context.Background())

	ami := ami.New(&cfg.AmiConfig, logger)
	if err := ami.Start(ctx); err != nil {
		logger.Fatal("msg", "Cannot start ami", "error", err)
	}

	voip := adapter.NewVoipAdapter(&cfg.VoipAdapterConfig, ami)
	queue := adapter.NewQueue(int(cfg.QueueConfig.MaxClients))
	clientRepo := adapter.NewClientRepo(logger)
	operatorRepo := adapter.NewOperatorRepo(&cfg.OperatorRepo, logger)
	strategy := service.NewRrStrategy(&cfg.DialerConfig, voip, operatorRepo, logger)
	dialer := service.NewDialerService(&cfg.DialerConfig, queue, operatorRepo, strategy, logger)
	dialer.Start(ctx)

	clientService := service.NewClientService(&cfg.ClientService, voip, queue, clientRepo, dialer, logger)
	clientService.Start(ctx)

	operatorService := service.NewOperatorService(operatorRepo, dialer, logger)

	clientHandler := handler.NewClientHandler(clientService, logger)
	clientHandler.Register(ami)

	operatorHandler := handler.NewOperatorHandler(operatorService, logger)
	operatorHandler.Register(ami)

	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-c

	cancel()

	operatorRepo.Close()
	operatorRepo.Close()
	clientRepo.Close()
	queue.Close()
	ami.Close()
}
