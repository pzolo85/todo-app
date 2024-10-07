package mail

import (
	"fmt"
	"log/slog"
	"time"
	"todo/internal/config"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
)

type DefaultService struct {
	logger *slog.Logger
	cache  *cache.Cache
	config *config.Config
}

func NewDefaultService(logger *slog.Logger, cache *cache.Cache, cfg *config.Config) *DefaultService {
	return &DefaultService{
		logger: logger,
		cache:  cache,
		config: cfg,
	}
}

func (s *DefaultService) SendChallenge(email string) error {
	challenge := uuid.NewString()
	s.logger.Info("new challenge", "email", email, "challenge", challenge, "url", fmt.Sprintf("http://%s:%d/api/v1/user/validate?email=%s&challenge=%s", s.config.Address, s.config.Port, email, challenge))
	err := s.cache.Add(challenge, email, time.Hour*24)
	if err != nil {
		return fmt.Errorf("failed to store challenge in cache > %w", err)
	}

	return nil

}

func (s *DefaultService) VerifyChallenge(email string, challenge string) error {
	s.logger.Info("verify challenge", "email", email, "challenge", challenge)
	cacheEmail, found := s.cache.Get(challenge)
	if !found {
		return fmt.Errorf("invalid challenge")
	}

	emailString, ok := cacheEmail.(string)
	if !ok {
		return fmt.Errorf("corrupted value in cache > %#v", cacheEmail)
	}

	if emailString != email {
		return fmt.Errorf("invalid challenge email")
	}

	s.cache.Delete(challenge)

	return nil
}

func (s *DefaultService) ListChallenges() map[string]string {
	m := s.cache.Items()
	outmap := make(map[string]string, len(m))
	for k, v := range m {
		s.logger.Debug("challenges", "key", k, "value", v)
		outmap[k] = v.Object.(string)
	}
	return outmap
}
