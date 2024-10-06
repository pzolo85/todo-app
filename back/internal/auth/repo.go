package auth

type Repo interface {
	GetUser(email string) (*UserClaim, error)
	SaveUser(user *UserClaim) error
}
