package user

import (
	"log/slog"
	"net/http"
	"time"
	"todo/internal/auth"
	"todo/internal/mail"
	userDB "todo/internal/repo/user"

	"github.com/labstack/echo/v4"
)

type DefaultHandler struct {
	repo     userDB.Repo
	logger   *slog.Logger
	mailSvc  mail.Service
	userRole string
}

type UserCreateRequest struct {
	Email      string `json:"email,omitempty"`
	Salt       string `json:"salt,omitempty"`
	HashedPass string `json:"hashed_pass,omitempty"`
}

func NewDefaultHandler(repo userDB.Repo, logger *slog.Logger, mailSvc mail.Service, userRole string) *DefaultHandler {
	return &DefaultHandler{
		repo:     repo,
		logger:   logger.WithGroup("user_handler"),
		mailSvc:  mailSvc,
		userRole: userRole,
	}
}

func (h *DefaultHandler) AddHandler(g *echo.Group, claimMW echo.MiddlewareFunc) {
	g.POST("/create", h.CreateUser)
	g.DELETE("/", h.DeleteUser, claimMW)
}

func (h *DefaultHandler) DeleteUser(c echo.Context) error {
	claimAny := c.Get(auth.UserClaimContextKey)

	claim, ok := claimAny.(*auth.UserClaim)
	if !ok {
		h.logger.Error("failed to parse claim from context", "claim", claimAny)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	err := h.repo.DeleteUser(claim.Email)
	if err != nil {
		h.logger.Error("failed to delete user", "err", err.Error())
		return echo.NewHTTPError(http.StatusBadGateway)
	}

	return nil

}

func (h *DefaultHandler) CreateUser(c echo.Context) error {
	var req UserCreateRequest
	err := c.Bind(&req)
	if err != nil {
		h.logger.Error("failed to deconde user create request", "err", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	var user = userDB.User{
		Email:        req.Email,
		PassHash:     req.HashedPass,
		Salt:         req.Salt,
		Role:         h.userRole,
		CreatedAt:    time.Now(),
		ValidEmail:   false,
		ActiveJWT:    []string{},
		Notes:        []string{},
		SharedWithMe: []string{},
	}

	err = h.mailSvc.SendChallenge(req.Email)
	if err != nil {
		h.logger.Error("failed to send email challenge", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	err = h.repo.SaveUser(&user)
	if err != nil {
		h.logger.Error("failed to save user to db", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusAccepted, user)

}
