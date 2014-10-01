package stemcell

type Infrastructure interface {
	CreateStemcell(Stemcell) (CID, error)
	//	DeleteStemcell(CID) error
}

type infrastructure struct{}
