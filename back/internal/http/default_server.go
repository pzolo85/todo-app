package http

import (
	"fmt"
	"log/slog"
	"todo/internal/auth"
	"todo/internal/mail"
	"todo/internal/user"

	"github.com/labstack/echo/v4"
)

type DefaultServer struct {
	srv       *echo.Echo
	logger    *slog.Logger
	adminRole string
}

func GetDefaultServer(echo *echo.Echo, log *slog.Logger, adminRole string) *DefaultServer {
	return &DefaultServer{
		srv:       echo,
		logger:    log,
		adminRole: adminRole,
	}
}

func (s *DefaultServer) LoadRoutes(authHandler *auth.Handler, mailHandler *mail.DefaultHandler, userHandler *user.DefaultHandler) error {
	// api/v1
	v1grp := s.srv.Group("/api/v1")

	// user
	userGrp := v1grp.Group("/user")

	// auth
	authGrp := v1grp.Group("/auth")

	// admin
	adminGrp := v1grp.Group("/admin",
		authHandler.AddUserClaim(),
		authHandler.VerifyRole([]string{s.adminRole}),
	)

	// admin/mail
	mailGrp := adminGrp.Group("/mail")

	// add handlers
	authHandler.AddHandler(authGrp)
	mailHandler.AddHandler(mailGrp)
	userHandler.AddHandler(userGrp, authHandler.AddUserClaim())

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
