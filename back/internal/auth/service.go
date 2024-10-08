package auth

import "github.com/pzolo85/todo-app/back/internal/claim"

type Service interface {
	DecodeToken(t string) (*claim.UserClaim, error)
	GetJWT(u *claim.UserClaim) (string, error)
}
