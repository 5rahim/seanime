package mal

type (
	Wrapper struct {
		AccessToken string
	}
)

func NewWrapper(accessToken string) *Wrapper {
	return &Wrapper{
		AccessToken: accessToken,
	}
}
