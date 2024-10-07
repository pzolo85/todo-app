// Package auth handles JWT generation and parsing
//
// # Middlewares
//
// Package auth provides middlewares for authn/authz
//
//	AddUserClaim() echo.MiddlewareFunc // Decodes the JWT into a UserClaim in the echo Context
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

// Holds the Claim section of the JWT
type UserClaim struct {
	Email     string    `json:"email" mapstructure:"email"`
	CreatedAt time.Time `json:"created_at" mapstructure:"created_at"`
	ExpiresAt time.Time `json:"expires_at" mapstructure:"expires_at"`
	IsAdmin   bool      `json:"is_admin" mapstructure:"is_admin"`
	SourceIP  string    `json:"source_address" mapstructure:"source_address"`
	UserAgent string    `json:"user_agent" mapstructure:"user_agent"`
	ClaimID   string    `json:"claim_id" mapstructure:"claim_id"`
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

	t := jwt.NewWithClaims(s.signingMethod, u)
	tstr, err := t.SignedString(s.key)
	if err != nil {
		return "", fmt.Errorf("failed to sign token > %w", err)
	}

	return tstr, nil
}

func (s *DefaultService) DecodeToken(t string) (*UserClaim, error) {
	token, err := jwt.Parse(t, func(t *jwt.Token) (any, error) {
		if t.Method != s.signingMethod {
			return nil, fmt.Errorf("invalid signing method: %s", t.Method)
		}
		return s.key, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token > %w", err)
	}

	var userClaim UserClaim
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

	dhf := mapstructure.StringToTimeHookFunc(time.RFC3339)
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: dhf,
		Result:     &userClaim,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create custom decoder > %w", err)
	}

	err = dec.Decode(mapClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to create custom decoder > %w", err)
	}

	if userClaim.Valid() != nil {
		return nil, fmt.Errorf("failed to validate claims > %w", userClaim.Valid())
	}

	return &userClaim, nil
}

func validateUser(u *UserClaim) error {
	errFmt := "invalid user claim > %s cannot be nil"
	switch {
	case u.Email == "":
		return fmt.Errorf(errFmt, "email")
	case u.SourceIP == "":
		return fmt.Errorf(errFmt, "source_address")
	case u.UserAgent == "":
		return fmt.Errorf(errFmt, "user_agent")
	case u.ClaimID == "":
		return fmt.Errorf(errFmt, "claim_id")
	}
	return nil
}
