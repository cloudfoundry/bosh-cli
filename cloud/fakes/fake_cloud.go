package fakes

type FakeCloud struct {
	CreateStemcellInputs []CreateStemcellInput
	CreateStemcellCID    string
	CreateStemcellErr    error

	CreateVMInput CreateVMInput
	CreateVMCID   string
	CreateVMErr   error
}

type CreateStemcellInput struct {
	CloudProperties map[string]interface{}
	ImagePath       string
}

type CreateVMInput struct {
	StemcellCID     string
	CloudProperties map[string]interface{}
	NetworksSpec    map[string]interface{}
	Env             map[string]interface{}
}

func NewFakeCloud() *FakeCloud {
	return &FakeCloud{
		CreateStemcellInputs: []CreateStemcellInput{},
	}
}

func (c *FakeCloud) CreateStemcell(cloudProperties map[string]interface{}, imagePath string) (string, error) {
	c.CreateStemcellInputs = append(c.CreateStemcellInputs, CreateStemcellInput{
		CloudProperties: cloudProperties,
		ImagePath:       imagePath,
	})

	return c.CreateStemcellCID, c.CreateStemcellErr
}

func (c *FakeCloud) CreateVM(
	stemcellCID string,
	cloudProperties map[string]interface{},
	networksSpec map[string]interface{},
	env map[string]interface{},
) (string, error) {
	c.CreateVMInput = CreateVMInput{
		StemcellCID:     stemcellCID,
		CloudProperties: cloudProperties,
		NetworksSpec:    networksSpec,
		Env:             env,
	}

	return c.CreateVMCID, c.CreateVMErr
}
