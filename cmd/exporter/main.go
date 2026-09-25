package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"mc-monitor/internal/collector"
	"mc-monitor/internal/probe"
)

func main() {
	// Flags
	listenAddr := flag.String("listen", ":9092", "address to listen on")
	probeInterval := flag.Duration("interval", 30*time.Second, "probe interval")
	probeTimeout := flag.Duration("timeout", 5*time.Second, "probe timeout")
	probeLimit := flag.Int("limit", 20, "probe limit")
	flag.Parse()
	// Logger Init
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	// Cache Init
	cache := probe.NewCache()
	// targets
	targets := []probe.Target{
		{Name: "main", Addr: "127.0.0.1:25565"},
		{Name: "main1", Addr: "127.0.0.1:25565"},
		{Name: "main2", Addr: "127.0.0.1:25565"},
		{Name: "main3", Addr: "127.0.0.1:25565"},
		{Name: "main4", Addr: "127.0.0.1:25565"},
		{Name: "main5", Addr: "127.0.0.1:25565"},
		{Name: "main6", Addr: "127.0.0.1:25565"},
		{Name: "main7", Addr: "127.0.0.1:25565"},
		{Name: "main8", Addr: "127.0.0.1:25565"},
		{Name: "main9", Addr: "127.0.0.1:25565"},
		{Name: "main10", Addr: "127.0.0.1:25565"},
		{Name: "main11", Addr: "127.0.0.1:25565"},
	}
	// prober init
	prober := probe.NewProber(targets, cache, *probeInterval, *probeTimeout, *probeLimit, logger)
	// context Init
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go prober.Run(ctx)

	// collector regist and init

	mc, err := collector.NewMCCollector(cache)
	if err != nil {
		logger.Error("fail to load mc collector", "err", err)
		os.Exit(1)
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(mc)

	// http service init and routes regist
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:    *listenAddr,
		Handler: mux,
	}
	// http service run
	go func() {
		logger.Info("exporter listening", "addr", *listenAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server run failed", "err", err)
			cancel()
		}
	}()
	// wait for shutdown signal
	<-ctx.Done()
	logger.Info("find shutdown signal,exiting")
	// wait 5s for http requests before program shutdown
	shutdownCtx, shutdownCacel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCacel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown error", "err", err)
	}

	logger.Info("exporter stopped")
}
