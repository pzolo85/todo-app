package mail

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
)

type DefaultService struct {
	logger *slog.Logger
	cache  *cache.Cache
}

func NewDefaultService(logger *slog.Logger, cache *cache.Cache) *DefaultService {
	return &DefaultService{
		logger: logger,
		cache:  cache,
	}
}

func (s *DefaultService) SendChallenge(email string) error {
	challenge := uuid.NewString()
	s.logger.Info("new challenge", "email", email, "challenge", challenge)
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

	return nil
}
