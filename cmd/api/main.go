package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "go.uber.org/automaxprocs/maxprocs"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	cachepkg "azret/internal/cache"
	cfgpkg "azret/internal/config"
	routerpkg "azret/internal/http"
	handlerpkg "azret/internal/http/handlers"
	repositorypkg "azret/internal/repository"
	servicepkg "azret/internal/service"
)

func main() {

	//runtime.GOMAXPROCS(4)

	// Logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Config
	cfg, err := cfgpkg.Load()
	if err != nil {
		logger.Fatal("config_load_error", zap.Error(err))
	}

	// Gin mode
	gin.SetMode(gin.ReleaseMode)

	// PostgreSQL pool
	pgCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("pg_parse_config_error", zap.Error(err))
	}
	pgCfg.MaxConns = int32(cfg.PGMaxConns)
	pgCfg.HealthCheckPeriod = cfg.PGHealthCheckPeriod()
	pgCfg.MaxConnIdleTime = cfg.PGMaxConnIdleTime()
	pgCfg.MaxConnLifetime = cfg.PGMaxConnLifetime()
	dbpool, err := pgxpool.NewWithConfig(context.Background(), pgCfg)
	if err != nil {
		logger.Fatal("pg_pool_create_error", zap.Error(err))
	}
	defer dbpool.Close()

	// Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:            cfg.RedisAddr,
		Username:        cfg.RedisUsername,
		Password:        cfg.RedisPassword,
		DB:              cfg.RedisDB,
		DialTimeout:     time.Duration(cfg.RedisDialTimeout) * time.Millisecond,
		ReadTimeout:     time.Duration(cfg.RedisReadTimeout) * time.Millisecond,
		WriteTimeout:    time.Duration(cfg.RedisWriteTimeout) * time.Millisecond,
		PoolSize:        cfg.RedisPoolSize,
		MinIdleConns:    cfg.RedisMinIdleConns,
		PoolTimeout:     time.Duration(cfg.RedisPoolTimeoutMs) * time.Millisecond,
		MaxRetries:      cfg.RedisMaxRetries,
		ConnMaxIdleTime: time.Duration(cfg.RedisConnMaxIdleSec) * time.Second,
	})
	defer rdb.Close()

	// Compose dependencies
	repo := repositorypkg.NewUserRepository(dbpool)
	cache := cachepkg.NewUserCache(rdb, cfg.CacheTTL())
	svc := servicepkg.NewUserService(repo, cache, cfg.RequestTimeout())
	usersHandler := handlerpkg.NewUsersHandler(svc, logger)
	router := routerpkg.NewRouter(usersHandler)

	// HTTP server
	srv := &http.Server{
		Addr:           ":" + strconv.Itoa(cfg.Port),
		Handler:        router,
		ReadTimeout:    cfg.ReadTimeout(),
		WriteTimeout:   cfg.WriteTimeout(),
		IdleTimeout:    cfg.IdleTimeout(),
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		logger.Info("server_start", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server_error", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server_shutdown_error", zap.Error(err))
	}
	logger.Info("server_stopped")
}
