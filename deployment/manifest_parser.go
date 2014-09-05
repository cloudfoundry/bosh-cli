package deployment

type ManifestParser interface {
	Parse(manifestPath string) (Deployment, error)
}
