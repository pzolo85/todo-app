package user

type Repo interface {
	GetUser(email string) (*User, error)
	SaveUser(u *User) error
	DeleteUser(email string) error
}
