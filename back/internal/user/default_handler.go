package user

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/pzolo85/todo-app/back/internal/auth"
	"github.com/pzolo85/todo-app/back/internal/mail"
	userDB "github.com/pzolo85/todo-app/back/internal/repo/user"

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
type ModifyUserRequest struct {
	Email string `json:"email,omitempty"`
}

func NewDefaultHandler(repo userDB.Repo, logger *slog.Logger, mailSvc mail.Service, userRole string) *DefaultHandler {
	return &DefaultHandler{
		repo:     repo,
		logger:   logger.WithGroup("user_handler"),
		mailSvc:  mailSvc,
		userRole: userRole,
	}
}

func (h *DefaultHandler) AddHandler(userGroup *echo.Group, adminGroup *echo.Group, claimMW echo.MiddlewareFunc, validMW echo.MiddlewareFunc) {
	userGroup.POST("/create", h.CreateUser)
	userGroup.GET("/validate", h.ValidateUser)
	userGroup.GET("/info", h.Info, claimMW, validMW)
	userGroup.GET("/resend-challenge", h.ResendChallenge, claimMW)
	userGroup.DELETE("/", h.DeleteUser, claimMW)

	adminUserGroup := adminGroup.Group("/user")
	adminUserGroup.PUT("/disable", h.DisableUser)
	adminUserGroup.PUT("/make-admin", h.MakeAdmin)
	adminUserGroup.PUT("/disable-admin", h.DisableAdmin)
}

func (h *DefaultHandler) ResendChallenge(c echo.Context) error {
	clm := c.Get(auth.UserClaimContextKey)
	claim, ok := clm.(*auth.UserClaim)
	if !ok {
		h.logger.Error("failed to parse claim from context", "claim", clm)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	err := h.mailSvc.SendChallenge(claim.Email)
	if err != nil {
		return fmt.Errorf("failed to send challenge > %w", err)
	}

	return c.NoContent(http.StatusOK)
}

func (h *DefaultHandler) Info(c echo.Context) error {
	clm := c.Get(auth.UserClaimContextKey)

	claim, ok := clm.(*auth.UserClaim)
	if !ok {
		h.logger.Error("failed to parse claim from context", "claim", clm)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	u, err := h.repo.GetUser(claim.Email)
	if err != nil {
		h.logger.Error("failed to get user", "err", err.Error())
		return echo.NewHTTPError(http.StatusBadGateway, err)
	}

	return c.JSON(http.StatusOK, u)
}

func (h *DefaultHandler) ValidateUser(c echo.Context) error {
	email := c.QueryParam("email")
	challenge := c.QueryParam("challenge")
	err := h.mailSvc.VerifyChallenge(email, challenge)
	if err != nil {
		h.logger.Warn("invalid challenge validation", "email", email, "challenge", challenge)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid challenge validation")
	}

	err = h.repo.EnableUser(email)
	if err != nil {
		h.logger.Error("failed to enable user", "err", err.Error())
		return echo.NewHTTPError(http.StatusBadGateway, err)
	}

	u, err := h.repo.GetUser(email)
	if err != nil {
		h.logger.Error("failed to enable user", "err", err.Error())
		return echo.NewHTTPError(http.StatusBadGateway, err)
	}

	return c.JSON(http.StatusOK, u)
}

func (h *DefaultHandler) MakeAdmin(c echo.Context) error {
	var req ModifyUserRequest
	err := c.Bind(&req)
	if err != nil {
		h.logger.Error("failed to decode modify user request", "err", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = h.repo.MakeAdmin(req.Email)
	if err != nil {
		h.logger.Error("failed to make user admin", "err", err.Error())
		return echo.NewHTTPError(http.StatusBadGateway, err)
	}

	return nil
}

func (h *DefaultHandler) DisableUser(c echo.Context) error {
	var req ModifyUserRequest
	err := c.Bind(&req)
	if err != nil {
		h.logger.Error("failed to decode user create request", "err", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = h.repo.DisableUser(req.Email)
	if err != nil {
		h.logger.Error("failed to disable user", "err", err.Error())
		return echo.NewHTTPError(http.StatusBadGateway, err)
	}

	return nil
}

func (h *DefaultHandler) DisableAdmin(c echo.Context) error {
	var req ModifyUserRequest
	err := c.Bind(&req)
	if err != nil {
		h.logger.Error("failed to decode user modify request", "err", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = h.repo.DisableAdmin(req.Email)
	if err != nil {
		h.logger.Error("failed to disable admin access", "err", err.Error())
		return echo.NewHTTPError(http.StatusBadGateway, err)
	}

	return nil
}

func (h *DefaultHandler) DeleteUser(c echo.Context) error {
	clm := c.Get(auth.UserClaimContextKey)

	claim, ok := clm.(*auth.UserClaim)
	if !ok {
		h.logger.Error("failed to parse claim from context", "claim", clm)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	err := h.repo.DeleteUser(claim.Email)
	if err != nil {
		h.logger.Error("failed to delete user", "err", err.Error())
		return echo.NewHTTPError(http.StatusBadGateway, err)
	}

	return nil

}

func (h *DefaultHandler) CreateUser(c echo.Context) error {
	var req UserCreateRequest
	err := c.Bind(&req)
	if err != nil {
		h.logger.Error("failed to decode user create request", "err", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err)
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
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = h.repo.SaveUser(&user, false)
	if err != nil {
		h.logger.Error("failed to save user to db", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusAccepted, user)
}
