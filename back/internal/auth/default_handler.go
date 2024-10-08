package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/pzolo85/todo-app/back/internal/claim"
	"github.com/pzolo85/todo-app/back/internal/user"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc  Service
	repo user.Repo
	log  *slog.Logger
}

type LoginRequest struct {
	Email string `json:"email"`
	Hash  string `json:"hash"`
}
type LoginResponse struct {
	Token string `json:"token"`
}

const AuthHeader = "x-auth-token"

func NewDefaultHandler(svc Service, log *slog.Logger, repo user.Repo) *Handler {
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
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	user, err := h.repo.GetUser(req.Email)
	if err != nil {
		h.log.Warn("invalid user login attempt", "email", req.Email)
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if user.PassHash != req.Hash {
		h.log.Warn("invalid password login attempt", "email", req.Email)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("unknown user: %s", req.Email))
	}

	token, err := h.svc.GetJWT(&claim.UserClaim{
		Email:     req.Email,
		CreatedAt: time.Now(),
		SourceIP:  c.RealIP(),
		UserAgent: c.Request().UserAgent(),
		ClaimID:   uuid.NewString(),
	})

	user.ActiveJWT = append(user.ActiveJWT, token)
	err = h.repo.SaveUser(user, true)
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
			userClaim, ok := c.Get(claim.UserClaimContextKey).(*claim.UserClaim)
			if !ok {
				h.log.Warn("failed to extract claims from context")
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to extract claims")
			}

			if userClaim.IsAdmin {
				return next(c)
			}

			user, err := h.repo.GetUser(userClaim.Email)
			if err != nil {
				h.log.Error("failed to get user from db", "err", err.Error())
				return echo.NewHTTPError(http.StatusInternalServerError, err)
			}

			if !slices.Contains(validRoles, user.Role) {
				h.log.Warn("unauthorized access to protected resource",
					slog.String("path", c.Request().RequestURI),
					slog.String("real_ip", c.RealIP()),
				)
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			return next(c)
		}
	}
}

func (h *Handler) VerifyValidAccount() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userClaim, ok := c.Get(claim.UserClaimContextKey).(*claim.UserClaim)
			if !ok {
				h.log.Warn("failed to extract claims from context")
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to extract claims")
			}

			if userClaim.IsAdmin {
				return next(c)
			}

			user, err := h.repo.GetUser(userClaim.Email)
			if err != nil {
				h.log.Error("failed to get user from db", "err", err.Error())
				return echo.NewHTTPError(http.StatusInternalServerError, err)
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
			token := c.Request().Header.Get(AuthHeader)
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
				return echo.NewHTTPError(http.StatusUnauthorized, err)
			}

			h.log.Debug("user claim decoded from request", "claim", t)
			if t.IsAdmin && t.ExpiresAt.Before(time.Now()) {
				h.log.Warn("auth attempt with expired JWT admin token ", "token", t)
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			if t.IsAdmin {
				c.Set(claim.UserClaimContextKey, t)
				return next(c)
			}

			// verify if token is allowed
			user, err := h.repo.GetUser(t.Email)
			if err != nil {
				h.log.Error("failed to get user from db", "err", err.Error())
				return echo.NewHTTPError(http.StatusInternalServerError, err)

			}

			if !slices.Contains(user.ActiveJWT, token) {
				h.log.Warn("auth attempt with removed JWT token ", "token", t)
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			c.Set(claim.UserClaimContextKey, t)
			return next(c)
		}
	}
}
