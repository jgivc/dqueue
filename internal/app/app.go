package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"

	"github.com/jgivc/dqueue/config"
	"github.com/jgivc/dqueue/internal/adapter"
	"github.com/jgivc/dqueue/internal/handler"
	"github.com/jgivc/dqueue/internal/service"
	"github.com/jgivc/dqueue/pkg/ami"
	"github.com/jgivc/dqueue/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	promClientsGauge := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "app_client_repo_clients_length",
		Help: "Clients count in clients repo",
	})

	clientRepo := adapter.NewClientRepo(promClientsGauge, logger)
	operatorRepo := adapter.NewOperatorRepo(&cfg.OperatorRepo, logger)
	strategy := service.NewRrStrategy(&cfg.DialerConfig, voip, operatorRepo, logger)
	dialer := service.NewDialerService(&cfg.DialerConfig, queue, operatorRepo, strategy, logger)
	dialer.Start(ctx)

	clientService := service.NewClientService(&cfg.ClientService, voip, queue, clientRepo, dialer, logger)
	clientService.Start(ctx)

	operatorService := service.NewOperatorService(operatorRepo, dialer, logger)

	clientHandler := handler.NewClientHandler(clientService, cfg.Context, logger)
	clientHandler.Register(ami)

	operatorHandler := handler.NewOperatorHandler(operatorService, logger)
	operatorHandler.Register(ami)

	httpServer := http.Server{
		Addr:              cfg.ListenAddr,
		ReadHeaderTimeout: time.Second,
	}

	http.Handle(cfg.MetricPath, promhttp.Handler())

	if err := httpServer.ListenAndServe(); err != nil {
		logger.Fatal("msg", "HTTP server ListenAndServe Error", "error", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-c

	cancel()

	if err := httpServer.Shutdown(context.Background()); err != nil {
		logger.Error("msg", "Cannot shutdown http server", "error", err)
	}
	voip.Close()
	operatorRepo.Close()
	operatorRepo.Close()
	clientRepo.Close()
	queue.Close()
	ami.Close()
}
