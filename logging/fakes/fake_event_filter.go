package fakes

import (
	bmlog "github.com/cloudfoundry/bosh-micro-cli/logging"
)

type FilterDelegate func(*bmlog.Event) error

type FakeEventFilter struct {
	filterDelegate FilterDelegate
}

func NewFakeEventFilter() *FakeEventFilter {
	return &FakeEventFilter{}
}

func (f *FakeEventFilter) Filter(event *bmlog.Event) error {
	return f.filterDelegate(event)
}

func (f *FakeEventFilter) SetFilterBehavior(delegate FilterDelegate) {
	f.filterDelegate = delegate
}
