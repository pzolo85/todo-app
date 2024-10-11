package mail

import (
	"fmt"
	"net/http"

	"github.com/pzolo85/todo-app/back/internal/claim"
	"github.com/pzolo85/todo-app/back/internal/config"
	"github.com/pzolo85/todo-app/back/internal/notify"

	"github.com/labstack/echo/v4"
)

type DefaultHandler struct {
	svc Service
	ntf Notify
	cfg *config.Config
}

func NewDefaultHandler(svc Service, ntf Notify, cfg *config.Config) *DefaultHandler {
	return &DefaultHandler{
		svc: svc,
		ntf: ntf,
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
	g.GET("/sub", h.Sub)
}

func (h *DefaultHandler) Sub(c echo.Context) error {
	//clm := c.Get(claim.UserClaimContextKey)
	//usrClaim, ok := clm.(*claim.UserClaim)
	//if !ok {
	//return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("failed to extract user claim"))
	//}
	appID := c.Get(claim.AppIDContextKey)
	appIDStr, ok := appID.(string)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("failed to extract appid > %s", appID))
	}

	go func() {
		challenges := h.svc.ListChallenges()
		i := 1
		for k, v := range challenges {
			h.ntf.Notify(
				[]byte(fmt.Sprintf(`{"key":"%s","val":"%s"}`, k, v)),
				i,
				notify.Topic{
					Type:  "email",
					Topic: appIDStr,
				},
			)
			i++
		}
	}()
	return h.ntf.Subscribe(c, []string{"mail", appIDStr})
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
