package http

import (
	"fmt"
	"log/slog"
	"todo/internal/auth"
	"todo/internal/mail"

	"github.com/labstack/echo/v4"
)

type DefaultServer struct {
	srv    *echo.Echo
	logger *slog.Logger
}

func GetDefaultServer(echo *echo.Echo, log *slog.Logger) *DefaultServer {
	return &DefaultServer{
		srv:    echo,
		logger: log,
	}
}

func (s *DefaultServer) LoadRoutes(authHandler *auth.Handler, mailHandler *mail.DefaultHandler) error {

	// api/v1
	v1grp := s.srv.Group("/api/v1")

	// auth
	authGrp := v1grp.Group("/auth")

	// admin
	adminGrp := v1grp.Group("/admin",
		authHandler.AddUserClaim(),
		authHandler.VerifyRole([]string{auth.AdminRole}),
	)

	// admin/mail
	mailGrp := adminGrp.Group("/mail")

	// add handlers
	authHandler.AddHandler(authGrp)
	mailHandler.AddHandler(mailGrp)

	return nil
}

func (s *DefaultServer) Start(add string, port int) {
	s.logger.Error("server crashed",
		slog.String(
			"err",
			s.srv.Start(fmt.Sprintf("%s:%d", add, port)).Error(),
		),
	)
}
