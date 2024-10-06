package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"
	"todo/internal/auth"
	"todo/internal/config"
	"todo/internal/http"
	"todo/internal/log"
	"todo/internal/mail"
	userDB "todo/internal/repo/user"
	"todo/internal/user"

	"github.com/boltdb/bolt"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
)

type Services struct {
	logger  *slog.Logger
	AuthSvc *auth.DefaultService
	AuthHdl *auth.Handler
	Server  *http.DefaultServer
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	svc, err := loadServices(cfg)
	if err != nil {
		panic(err)
	}

	svc.Server.Start(cfg.Address, cfg.Port)

}

func loadServices(cfg *config.Config) (*Services, error) {
	// logger
	appID := uuid.NewString()
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to extract hostname > %w", err)
	}

	logger := log.NewDefaultService(cfg.Level, appID, hostname)

	db, err := bolt.Open(cfg.DBPath, 0777, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open db > %w", err)
	}

	// mail
	mailCache := cache.New(time.Hour*24, time.Hour)
	mailSvc := mail.NewDefaultService(logger, mailCache)
	mailHandler := mail.NewDefaultHandler()

	// user
	userCache := cache.New(time.Hour, time.Minute*20)
	userRepo, err := userDB.NewDefaultRepo(db, userCache)
	if err != nil {
		return nil, fmt.Errorf("failed to create userRepo > %w", err)
	}
	userHandler := user.NewDefaultHandler(userRepo, logger, mailSvc, cfg.UserRole)

	// auth
	authSvc := auth.NewDefaultService(cfg.Key, jwt.SigningMethodHS256, logger)
	authHandler := auth.NewDefaultHandler(authSvc, logger, userRepo)

	// server
	e := echo.New()
	srv := http.GetDefaultServer(e, logger, cfg.AdminRole)
	err = srv.LoadRoutes(authHandler, mailHandler, userHandler)
	if err != nil {
		return nil, err
	}

	return &Services{
		logger:  logger,
		AuthSvc: authSvc,
		AuthHdl: authHandler,
		Server:  srv,
	}, nil

}
