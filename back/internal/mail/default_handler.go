package mail

import (
	"fmt"
	"net/http"

	"github.com/pzolo85/todo-app/back/internal/config"

	"github.com/labstack/echo/v4"
)

type DefaultHandler struct {
	svc Service
	cfg *config.Config
}

func NewDefaultHandler(svc Service, cfg *config.Config) *DefaultHandler {
	return &DefaultHandler{
		svc: svc,
		cfg: cfg,
	}
}

type Mails struct {
	Mails []Mail `json:"mails"`
}
type Mail struct {
	Subject string `json:"subject,omitempty"`
	Link    string `json:"link,omitempty"`
	To      string `json:"to,omitempty"`
}

func (h *DefaultHandler) AddHandler(g *echo.Group) {
	g.GET("/list", h.List)
}

func (h *DefaultHandler) List(c echo.Context) error {
	chmap := h.svc.ListChallenges()
	mails := make([]Mail, 0, len(chmap))
	for challenge, email := range chmap {
		mails = append(mails, Mail{
			Subject: fmt.Sprintf("verify your email"),
			To:      email,
			Link:    fmt.Sprintf("http://%s:%d/api/v1/user/validate?email=%s&challenge=%s", h.cfg.Address, h.cfg.Port, email, challenge),
		})
	}

	return c.JSON(http.StatusOK, Mails{
		Mails: mails,
	})
}
