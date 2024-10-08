// Copyright 2024 pzolo85. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// todo-app
package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/pzolo85/todo-app/back/internal/auth"
	"github.com/pzolo85/todo-app/back/internal/claim"
	"github.com/pzolo85/todo-app/back/internal/config"
	"github.com/pzolo85/todo-app/back/internal/http"
	"github.com/pzolo85/todo-app/back/internal/log"
	"github.com/pzolo85/todo-app/back/internal/mail"
	"github.com/pzolo85/todo-app/back/internal/user"

	"github.com/boltdb/bolt"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
)

// Services is a group of services and handlers.
type Services struct {
	logger  *slog.Logger
	AuthSvc *auth.DefaultService
	AuthHdl *auth.Handler
	Server  *http.DefaultServer
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config > %s", err.Error())
		os.Exit(2)
	}

	if cfg.GenerateKey || cfg.SignAdminToken {
		// we don't want to lock here waiting for the default db when loading the services
		file, err := os.CreateTemp(os.TempDir(), "todo_db_*")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create tmp db > %s", err.Error())
		}

		cfg.DBPath = file.Name()
	}

	svc, err := loadServices(cfg)
	if err != nil {
		panic(err)
	}

	// cli options
	switch {
	case cfg.GenerateKey:
		if err := GenerateKey(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(2)
		}
		os.Exit(0)
	case cfg.SignAdminToken:
		if err := GenerateToken(cfg, svc.AuthSvc); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(2)
		}
		os.Exit(0)
	}

	svc.logger.Debug("config", "cfg", cfg)
	svc.Server.Start(cfg.Address, cfg.Port)

}

// GenerateKey generates a passphrase for JWT validation
func GenerateKey(cfg *config.Config) error {
	defer os.Remove(cfg.DBPath)
	key := uuid.NewString()
	err := os.WriteFile(config.KeyFile, []byte(key), 0400)
	if err != nil {
		return fmt.Errorf("failed to generate key > %w", err)
	}
	fmt.Fprintf(os.Stdout, "new key generated: %s\n", config.KeyFile)
	return nil
}

// GenerateToken generates a JWT with admin permissions
func GenerateToken(cfg *config.Config, authSvc auth.Service) error {
	defer os.Remove(cfg.DBPath)
	now := time.Now()
	c := claim.UserClaim{
		Email:     cfg.SignEmail,
		CreatedAt: now,
		ExpiresAt: now.Add(cfg.SignDuration),
		IsAdmin:   true,
		ClaimID:   uuid.NewString(),
		SourceIP:  "127.0.0.1",
		UserAgent: "curl",
	}

	token, err := authSvc.GetJWT(&c)
	if err != nil {
		return fmt.Errorf("failed to generate JWT token > %w", err)
	}

	fmt.Fprintf(os.Stdout, "%s", token)
	return nil
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
	mailSvc := mail.NewDefaultService(logger, mailCache, cfg)
	mailHandler := mail.NewDefaultHandler(mailSvc, cfg)

	// user
	userCache := cache.New(time.Hour, time.Minute*20)
	userRepo, err := user.NewDefaultRepo(db, userCache, cfg.AdminRole, cfg.UserRole)
	if err != nil {
		return nil, fmt.Errorf("failed to create userRepo > %w", err)
	}
	userHandler := user.NewDefaultHandler(userRepo, logger, mailSvc, cfg.UserRole)

	// auth
	authSvc := auth.NewDefaultService(cfg.Key, jwt.SigningMethodHS256, logger)
	authHandler := auth.NewDefaultHandler(authSvc, logger, userRepo)

	// server
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
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
