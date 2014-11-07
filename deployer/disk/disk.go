package disk

type Disk interface {
	CID() string
}

type disk struct {
	cid string
}

func NewDisk(cid string) Disk {
	return &disk{
		cid: cid,
	}
}

func (d *disk) CID() string {
	return d.cid
}
