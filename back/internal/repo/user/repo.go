package user

type Repo interface {
	GetUser(email string) (*User, error)
	SaveUser(u *User) error
	DeleteUser(email string) error
	DisableUser(email string) error
	MakeAdmin(email string) error
	DisableAdmin(email string) error
	EnableUser(email string) error
}
