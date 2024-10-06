package mail

type Service interface {
	SendChallenge(email string) error
	VerifyChallenge(email string, challenge string) error
}
