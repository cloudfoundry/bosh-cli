package fakes

import (
	boshdisk "github.com/cloudfoundry/bosh-agent/platform/disk"
)

type FakeDiskManager struct {
	FakePartitioner           *FakePartitioner
	FakeFormatter             *FakeFormatter
	FakeMounter               *FakeMounter
	FakeMountsSearcher        *FakeMountsSearcher
	FakeRootDevicePartitioner *FakePartitioner
}

func NewFakeDiskManager() (manager *FakeDiskManager) {
	manager = &FakeDiskManager{}
	manager.FakePartitioner = NewFakePartitioner()
	manager.FakeFormatter = &FakeFormatter{}
	manager.FakeMounter = &FakeMounter{}
	manager.FakeMountsSearcher = &FakeMountsSearcher{}
	manager.FakeRootDevicePartitioner = NewFakePartitioner()
	return
}

func (m FakeDiskManager) GetPartitioner() boshdisk.Partitioner {
	return m.FakePartitioner
}

func (m FakeDiskManager) GetRootDevicePartitioner() boshdisk.Partitioner {
	return m.FakeRootDevicePartitioner
}

func (m FakeDiskManager) GetFormatter() boshdisk.Formatter {
	return m.FakeFormatter
}

func (m FakeDiskManager) GetMounter() boshdisk.Mounter {
	return m.FakeMounter
}

func (m FakeDiskManager) GetMountsSearcher() boshdisk.MountsSearcher {
	return m.FakeMountsSearcher
}
