package main

import (
	"fmt"
	"log/slog"
	"os"
	"todo/internal/auth"
	"todo/internal/config"
	"todo/internal/http"
	"todo/internal/log"
	"todo/internal/mail"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Services struct {
	logger  *slog.Logger
	AuthSvc *auth.DefaultService
	AuthHdl *auth.Handler
	Server  *echo.Echo
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

	svc.Server.Start(fmt.Sprintf("%s:%d", cfg.Address, cfg.Port))

}

func loadServices(cfg *config.Config) (*Services, error) {
	// logger
	appID := uuid.NewString()
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to extract hostname > %w", err)
	}

	logger := log.NewDefaultService(cfg.Level, appID, hostname)

	// auth
	authSvc := auth.NewDefaultService(cfg.Key, jwt.SigningMethodHS256, logger)
	authHandler := auth.NewDefaultHandler(authSvc, logger)

	// mail
	mailHandler := mail.NewDefaultHandler()

	// server
	server := http.GetDefaultServer()
	v1grp := server.Group("/api/v1")
	authGrp := v1grp.Group("/auth")
	adminGrp := v1grp.Group("/admin", authHandler.GetAuthMiddleware(auth.AdminRole))

	mailGrp := adminGrp.Group("/mail")

	// add handlers
	authHandler.AddHandler(authGrp)
	mailHandler.AddHandler(mailGrp)

	return &Services{
		logger:  logger,
		AuthSvc: authSvc,
		AuthHdl: authHandler,
		Server:  server,
	}, nil

}
