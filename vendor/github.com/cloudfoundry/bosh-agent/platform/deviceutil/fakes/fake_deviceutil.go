package fakes

type FakeDeviceUtil struct {
	GetFilesContentsFileNames []string
	GetFilesContentsError     error
	GetFilesContentsContents  [][]byte
}

func NewFakeDeviceUtil() (util *FakeDeviceUtil) {
	util = &FakeDeviceUtil{}
	return
}

func (util *FakeDeviceUtil) GetFilesContents(fileNames []string) ([][]byte, error) {
	util.GetFilesContentsFileNames = fileNames
	return util.GetFilesContentsContents, util.GetFilesContentsError
}
