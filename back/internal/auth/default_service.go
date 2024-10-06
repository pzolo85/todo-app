package auth

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/mitchellh/mapstructure"
	"github.com/patrickmn/go-cache"
)

type DefaultService struct {
	key           []byte
	signingMethod jwt.SigningMethod
	logger        *slog.Logger
	cache         *cache.Cache
	repo          Repo
	m             sync.Mutex
}

type UserClaim struct {
	Email     string    `json:"email" mapstructure:"email"`
	CreatedAt time.Time `json:"created_at" mapstructure:"created_at"`
	SourceIP  string    `json:"source_address" mapstructure:"source_address"`
	UserAgent string    `json:"user_agent" mapstructure:"user_agent"`
	Role      string    `json:"role" mapstructure:"role"`
	Validated bool      `json:"validated" mapstructure:"validated"`
}

func (u UserClaim) Valid() error {
	return validateUser(&u)
}

func NewDefaultService(key []byte, signingMethod jwt.SigningMethod, logger *slog.Logger, cache *cache.Cache, repo Repo) *DefaultService {
	return &DefaultService{
		key:           key,
		signingMethod: signingMethod,
		logger:        logger,
		cache:         cache,
		repo:          repo,
	}
}

func (s *DefaultService) GetJWT(u *UserClaim) (string, error) {
	if err := u.Valid(); err != nil {
		return "", err
	}

	t := jwt.NewWithClaims(s.signingMethod, u)
	tstr, err := t.SignedString(s.key)
	if err != nil {
		return "", fmt.Errorf("failed to sign token > %w", err)
	}

	return tstr, nil
}

func (s *DefaultService) DecodeToken(t string) (*UserClaim, error) {
	token, err := jwt.Parse(t, func(t *jwt.Token) (interface{}, error) {
		if t.Method != s.signingMethod {
			return nil, fmt.Errorf("invalid signing method: %s", t.Method)
		}
		return s.key, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token > %w", err)
	}

	var user UserClaim
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to convert claim to mapclaims")
	}

	for k, v := range token.Claims.(jwt.MapClaims) {
		s.logger.Debug("user claim: ",
			slog.String("key", k),
			slog.Any("value", v),
		)
	}
	mapstructure.Decode(mapClaims, &user)
	createdAt, err := time.Parse(time.RFC3339, mapClaims["created_at"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at")
	}
	user.CreatedAt = createdAt

	if user.Valid() != nil {
		return nil, fmt.Errorf("failed to validate claims > %w", user.Valid())
	}

	//	 check if user is still valid
	s.m.Lock()
	defer s.m.Unlock()
	cachedUser, ok := s.cache.Get(user.Email)
	if !ok {
		userDB, err := s.repo.GetUser(user.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to get user from db > %w", err)
		}
		s.cache.Set(user.Email, userDB, cache.DefaultExpiration)
		cachedUser = userDB
	}
	user.Role = cachedUser.(*UserClaim).Role
	user.Validated = cachedUser.(*UserClaim).Validated

	return &user, nil
}

func (s *DefaultService) ClearUser(email string) {
	s.m.Lock()
	defer s.m.Unlock()
	s.cache.Delete(email)
}

func validateUser(u *UserClaim) error {
	if u.Email == "" || u.SourceIP == "" || u.UserAgent == "" || u.Role == "" {
		return fmt.Errorf("%#v is not a valid user", u)
	}
	return nil
}
