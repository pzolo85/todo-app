package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
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
const (
	userClaim = "user_claims"
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

func (h *Handler) VerifyRole(validRoles []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get(userClaim).(*UserClaim)
			if !ok {
				h.log.Warn("failed to extract claims from context")
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to extract claims")
			}

			if !slices.Contains(validRoles, user.Role) {
				h.log.Warn("unauthorized access to protected resource",
					slog.String("path", c.Request().RequestURI),
					slog.String("real_ip", c.RealIP()),
				)
			}

			return next(c)
		}
	}
}

func (h *Handler) VerifyValidAccount() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get(userClaim).(*UserClaim)
			if !ok {
				h.log.Warn("failed to extract claims from context")
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to extract claims")
			}

			if !user.Validated {
				h.log.Warn("unvalidated user trying to access are for validated",
					slog.String("user", user.Email),
					slog.String("path", c.QueryString()),
				)
				return echo.NewHTTPError(http.StatusUnauthorized, "please validate your account")
			}

			return next(c)

		}
	}
}

func (h *Handler) AddUserClaim() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.Request().Header.Get("x-auth-token")
			if token == "" {
				h.log.Warn("x-auth-token header missing",
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

			h.log.Debug("request from user", "user", t)
			c.Set(userClaim, t)
			return next(c)
		}
	}
}
