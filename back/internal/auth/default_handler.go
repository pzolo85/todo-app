package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"time"
	userDB "todo/internal/repo/user"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc  Service
	repo userDB.Repo
	log  *slog.Logger
}

type LoginRequest struct {
	Email string `json:"email"`
	Hash  string `json:"hash"`
}
type LoginResponse struct {
	Token string `json:"token"`
}

const (
	UserClaimContextKey = "user_claims"
)

func NewDefaultHandler(svc Service, log *slog.Logger, repo userDB.Repo) *Handler {
	return &Handler{
		svc:  svc,
		log:  log.WithGroup("auth_handler"),
		repo: repo,
	}
}

func (h *Handler) AddHandler(g *echo.Group) {
	g.POST("/login", h.LoginHandler)
}

func (h *Handler) LoginHandler(c echo.Context) error {
	var req LoginRequest
	err := c.Bind(&req)
	if err != nil {
		h.log.Error("failed to bind login request", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to bind login request > %w", err))
	}

	user, err := h.repo.GetUser(req.Email)
	if err != nil {
		h.log.Warn("invalid user login attempt", "email", req.Email)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("unknown user: %s", req.Email))
	}

	if user.PassHash != req.Hash {
		h.log.Warn("invalid password login attempt", "email", req.Email)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("unknown user: %s", req.Email))
	}

	token, err := h.svc.GetJWT(&UserClaim{
		Email:     req.Email,
		CreatedAt: time.Now(),
		SourceIP:  c.RealIP(),
		UserAgent: c.Request().UserAgent(),
		ClaimID:   uuid.NewString(),
	})

	user.ActiveJWT = append(user.ActiveJWT, token)
	err = h.repo.SaveUser(user)
	if err != nil {
		h.log.Error("failed to store user changes to db", "err", err.Error())
	}

	return c.JSON(http.StatusOK, LoginResponse{
		Token: token,
	})
}

// middlewares
func (h *Handler) VerifyRole(validRoles []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userClaim, ok := c.Get(UserClaimContextKey).(*UserClaim)
			if !ok {
				h.log.Warn("failed to extract claims from context")
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to extract claims")
			}

			user, err := h.repo.GetUser(userClaim.Email)
			if err != nil {
				h.log.Error("failed to get user from db", "err", err.Error())
				return echo.NewHTTPError(http.StatusInternalServerError)
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
			userClaim, ok := c.Get(UserClaimContextKey).(*UserClaim)
			if !ok {
				h.log.Warn("failed to extract claims from context")
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to extract claims")
			}

			user, err := h.repo.GetUser(userClaim.Email)
			if err != nil {
				h.log.Error("failed to get user from db", "err", err.Error())
				return echo.NewHTTPError(http.StatusInternalServerError)
			}

			if !user.ValidEmail {
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

			// verify if token is allowed
			user, err := h.repo.GetUser(t.Email)
			if err != nil {
				h.log.Error("failed to get user from db", "err", err.Error())
				return echo.NewHTTPError(http.StatusInternalServerError)

			}

			if !slices.Contains(user.ActiveJWT, token) {
				h.log.Warn("auth attempt with removed JWT token ", "token", t)
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			h.log.Debug("request from user", "user", t)
			c.Set(UserClaimContextKey, t)
			return next(c)
		}
	}
}
