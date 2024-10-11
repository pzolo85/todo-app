package notify

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"

	"github.com/dunglas/mercure"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

type Service struct {
	key []byte
	hub *mercure.Hub
	log *slog.Logger
}

type Topic struct {
	Type  string
	Topic string
}

var publishAllToken string

func NewDefaultService(k []byte, h *mercure.Hub, l *slog.Logger) (*Service, error) {
	var err error
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"mercure": map[string][]string{
			"publish": {"*"},
		},
	})
	publishAllToken, err = t.SignedString(k)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT > %wl", err)
	}

	return &Service{
		key: k,
		hub: h,
		log: l.With(slog.String("service", "notify")),
	}, nil
}

func (s *Service) Notify(data []byte, id int, topics ...Topic) error {
	idStr := "="
	if id >= 0 {
		idStr += strconv.Itoa(id)
	}

	for _, t := range topics {
		typeStr := "="
		if t.Type != "" {
			typeStr += t.Type
		}

		mercureData := fmt.Sprintf("data=%s&id%s&type%s&retry=&topic=%s", data, idStr, typeStr, t.Topic)
		w := httptest.NewRecorder()
		bodyReader := strings.NewReader(mercureData)
		r, err := http.NewRequest(
			http.MethodPost,
			"mercure",
			bodyReader,
		)
		if err != nil {
			return fmt.Errorf("failed to create request > %w", err)
		}

		r.Header.Add(echo.HeaderContentType, echo.MIMEApplicationForm)
		r.Header.Add(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", publishAllToken))

		s.hub.PublishHandler(w, r)
		if w.Code != http.StatusOK {
			msg := w.Body.String()
			s.log.Error("failed to publish", "err", msg, "code", w.Code, "data", mercureData)
			return fmt.Errorf("failed to publish code(%d) msg: %s", w.Code, msg)
		}
	}
	return nil
}

func (s *Service) Subscribe(c echo.Context, topics []string) error {
	var topicPrefix []string
	for _, t := range topics {
		topicPrefix = append(topicPrefix, "topic="+t)
	}

	topicsString := strings.Join(topicPrefix, "&")
	w := c.Response().Writer
	r, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("mercure?%s", topicsString),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request > %s", err)
	}

	s.hub.SubscribeHandler(w, r)
	return nil
}
