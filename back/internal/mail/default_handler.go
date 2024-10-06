package mail

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type DefaultHandler struct {
}

func NewDefaultHandler() *DefaultHandler {
	return &DefaultHandler{}
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
	g.GET("/list", func(c echo.Context) error {
		return c.JSON(http.StatusOK, &Mails{
			[]Mail{
				{
					Subject: "hello",
					Link:    "http://as.com",
					To:      "ptil",
				},
			},
		})

	})
}
