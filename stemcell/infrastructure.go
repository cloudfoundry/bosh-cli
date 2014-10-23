package stemcell

type Infrastructure interface {
	CreateStemcell(Manifest) (CID, error)
	//	DeleteStemcell(CID) error
}

type infrastructure struct{}
