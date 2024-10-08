// Package claim holds a claim of user type
package claim

import (
	"fmt"
	"time"
)

const (
	UserClaimContextKey = "user_claims"
)

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
