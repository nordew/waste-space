package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"waste-space/internal/config"
	"waste-space/internal/controller/v1"
	"waste-space/internal/middleware"
	"waste-space/internal/service"
	"waste-space/internal/storage/cache"
	"waste-space/internal/storage/repository"
	"waste-space/pkg/auth"
	"waste-space/pkg/db"

	"github.com/gin-gonic/gin"
	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

type App struct {
	server *http.Server
	db     *gorm.DB
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	database, err := db.NewPostgres(db.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if err := runMigrations(sqlDB); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	redisClient, err := db.NewRedis(db.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())

	tokenService := auth.NewJWTService(cfg.JWT.Secret)
	tokenCache := cache.NewTokenCache(redisClient)
	userRepo := repository.NewUserRepository(database)
	userService := service.NewUserService(userRepo, tokenService, tokenCache)
	dumpsterRepo := repository.NewDumpsterRepository(database)
	dumpsterService := service.NewDumpsterService(dumpsterRepo)

	handler := v1.NewHandler(userService, dumpsterService, tokenService)
	handler.InitRoutes(router)

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &App{
		server: server,
		db:     database,
	}, nil
}

func (a *App) Run() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting server on %s", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}

	sqlDB, err := a.db.DB()
	if err == nil {
		sqlDB.Close()
	}

	log.Println("Server stopped")
	return nil
}

func runMigrations(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	log.Println("Migrations applied successfully")
	return nil
}
