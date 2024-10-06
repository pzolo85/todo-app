package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc Service
	log *slog.Logger
}

type LoginRequest struct {
	Email string `json:"email"`
	Hash  string `json:"hash"`
}
type LoginResponse struct {
	Token string `json:"token"`
}

const (
	AdminRole = "admin"
	UserRole  = "user"
)

func NewDefaultHandler(svc Service, log *slog.Logger) *Handler {
	return &Handler{
		svc: svc,
		log: log.WithGroup("auth_handler"),
	}
}

func (h *Handler) AddHandler(g *echo.Group) {
	g.POST("/login", func(c echo.Context) error {
		var req LoginRequest
		err := c.Bind(&req)
		if err != nil {
			h.log.Error("failed to bind login request", slog.String("error", err.Error()))
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to bind login request > %w", err))
		}

		// verify user details
		// get role details

		token, err := h.svc.GetJWT(&UserClaim{
			Email:     req.Email,
			CreatedAt: time.Now(),
			SourceIP:  c.RealIP(),
			UserAgent: c.Request().UserAgent(),
			Role:      AdminRole,
		})
		if err != nil {
			h.log.Error("failed to create jwt", slog.String("error", err.Error()))
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to generate jwt > %w", err))
		}

		return c.JSON(http.StatusOK, &LoginResponse{
			Token: token,
		})

	})
}

func (h *Handler) GetAuthMiddleware(userLevel string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.Request().Header.Get("x-auth-token")
			if token == "" {
				h.log.Warn("unauthenticated attempt to access protected route",
					"request_ip", c.RealIP(),
					slog.String("request_url", c.Path()),
				)
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			t, err := h.svc.DecodeToken(token)
			if err != nil {
				h.log.Warn("error attempting to decode", "err", err.Error())
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			if userLevel == AdminRole && t.Role != AdminRole {
				h.log.Warn("unauthenticated attempt to access protected route", "user_claim", t, "request", c.Request().URL)
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			h.log.Info("request from user", "user", t)
			c.Set(userClaim, t)
			return next(c)
		}
	}
}

const (
	userClaim = "user_claims"
)
