package disk

type Manager interface {
	GetPartitioner() Partitioner
	GetRootDevicePartitioner() Partitioner
	GetFormatter() Formatter
	GetMounter() Mounter
	GetMountsSearcher() MountsSearcher
}
