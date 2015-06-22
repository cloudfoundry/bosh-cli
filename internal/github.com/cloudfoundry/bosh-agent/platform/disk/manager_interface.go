package disk

import (
	boshdevutil "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/platform/deviceutil"
)

type Manager interface {
	GetPartitioner() Partitioner
	GetRootDevicePartitioner() Partitioner
	GetFormatter() Formatter
	GetMounter() Mounter
	GetMountsSearcher() MountsSearcher
	GetDiskUtil(diskPath string) boshdevutil.DeviceUtil
}
