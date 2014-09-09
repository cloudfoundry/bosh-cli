package stemcell

type Stemcell struct {
	ImagePath       string
	Name            string
	Version         string
	Sha1            string
	CloudProperties map[string]interface{} `yaml:"cloud_properties"`
}
