package license

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . DirReader

type DirReader interface {
	Read(string) (*License, error)
}
