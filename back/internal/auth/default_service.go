package auth

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/mitchellh/mapstructure"
)

type DefaultService struct {
	key           []byte
	signingMethod jwt.SigningMethod
	logger        *slog.Logger
}

type UserClaim struct {
	Email     string    `json:"email" mapstructure:"email"`
	CreatedAt time.Time `json:"created_at" mapstructure:"created_at"`
	SourceIP  string    `json:"source_address" mapstructure:"source_address"`
	UserAgent string    `json:"user_agent" mapstructure:"user_agent"`
	Role      string    `json:"role" mapstructure:"role"`
}

func (u UserClaim) Valid() error {
	return validateUser(&u)
}

func NewDefaultService(key []byte, signingMethod jwt.SigningMethod, logger *slog.Logger) *DefaultService {
	return &DefaultService{
		key:           key,
		signingMethod: signingMethod,
		logger:        logger,
	}
}

func (s *DefaultService) GetJWT(u *UserClaim) (string, error) {
	if err := u.Valid(); err != nil {
		return "", err
	}
	//if err := validateUser(u); err != nil {
	//return "", err
	//}

	//t := jwt.NewWithClaims(s.signingMethod, jwt.MapClaims{
	//"email":          u.Email,
	//"created_at":     u.CreatedAt.Format(time.RFC3339),
	//"source_address": string(u.SourceIP),
	//"user_agent":     u.UserAgent,
	//})
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

	return &user, nil
}

func validateUser(u *UserClaim) error {
	fmt.Println("about to validate userclaim", "claim", u)
	if u.Email == "" || u.SourceIP == "" || u.UserAgent == "" {
		return fmt.Errorf("%#v is not a valid user", u)
	}
	return nil
}
