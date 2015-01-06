package release

type Validator interface {
	Validate(release Release) error
}
