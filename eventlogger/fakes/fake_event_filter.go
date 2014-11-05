package fakes

import (
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type FilterDelegate func(*bmeventlog.Event) error

type FakeEventFilter struct {
	filterDelegate FilterDelegate
}

func NewFakeEventFilter() *FakeEventFilter {
	return &FakeEventFilter{}
}

func (f *FakeEventFilter) Filter(event *bmeventlog.Event) error {
	return f.filterDelegate(event)
}

func (f *FakeEventFilter) SetFilterBehavior(delegate FilterDelegate) {
	f.filterDelegate = delegate
}
