package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"wb-order-service/internal/cache"
	"wb-order-service/internal/config"
	httpapi "wb-order-service/internal/http"
	"wb-order-service/internal/kafka"
	"wb-order-service/internal/repo"
	"wb-order-service/internal/service"
)

func main() {
	cfg := config.Load()

	lvl := new(slog.LevelVar)
	switch cfg.LogLevel {
	case "DEBUG": lvl.Set(slog.LevelDebug)
	case "INFO": lvl.Set(slog.LevelInfo)
	case "WARN": lvl.Set(slog.LevelWarn)
	default: lvl.Set(slog.LevelInfo)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
	logger.Info("starting app", "name", cfg.AppName, "port", cfg.HTTPPort)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pgcfg, err := pgxpool.ParseConfig(cfg.PostgresURL)
	if err != nil { logger.Error("pg parse", "err", err); os.Exit(1) }
	pgcfg.MaxConns = int32(cfg.PGMaxConns)
	pool, err := pgxpool.NewWithConfig(ctx, pgcfg)
	if err != nil { logger.Error("pg connect", "err", err); os.Exit(1) }
	defer pool.Close()

	c := cache.NewTTL()
	c.StartJanitor(cfg.CacheJanitor)
	defer c.Stop()

	repository := repo.NewPG(pool)
	svc := service.New(repository, c, cfg.CacheTTL)

	cons := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID, svc, logger)
	go func(){ 
		if err := cons.Run(ctx); err != nil {
			if ctx.Err() == nil { logger.Error("consumer", "err", err) }
		}
	}()
	defer func(){ _ = cons.Close() }()

	h := httpapi.New(svc, logger, pool, cfg.KafkaBrokers, cfg.KafkaTopic)
	srv := &http.Server{
		Addr: ":" + cfg.HTTPPort,
		Handler: h,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func(){
		logger.Info("http server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http serve", "err", err)
			cancel()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	_ = srv.Shutdown(shutdownCtx)
	logger.Info("graceful shutdown complete")
}
