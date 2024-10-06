package auth

type Service interface {
	DecodeToken(t string) (*UserClaim, error)
	GetJWT(u *UserClaim) (string, error)
}
