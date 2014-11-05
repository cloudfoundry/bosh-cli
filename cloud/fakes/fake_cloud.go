package fakes

type FakeCloud struct {
	CreateStemcellInputs []CreateStemcellInput
	CreateStemcellCID    string
	CreateStemcellErr    error

	CreateVMInput CreateVMInput
	CreateVMCID   string
	CreateVMErr   error

	CreateDiskInput CreateDiskInput
	CreateDiskCID   string
	CreateDiskErr   error
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

type CreateDiskInput struct {
	Size            int
	CloudProperties map[string]interface{}
	InstanceID      string
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

func (c *FakeCloud) CreateDisk(
	size int,
	cloudProperties map[string]interface{},
	instanceID string,
) (string, error) {
	c.CreateDiskInput = CreateDiskInput{
		Size:            size,
		CloudProperties: cloudProperties,
		InstanceID:      instanceID,
	}

	return c.CreateDiskCID, c.CreateDiskErr
}
