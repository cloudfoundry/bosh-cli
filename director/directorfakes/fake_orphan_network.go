// Code generated by counterfeiter. DO NOT EDIT.
package directorfakes

import (
	"sync"
	"time"

	"github.com/cloudfoundry/bosh-cli/v7/director"
)

type FakeOrphanNetwork struct {
	CreatedAtStub        func() time.Time
	createdAtMutex       sync.RWMutex
	createdAtArgsForCall []struct {
	}
	createdAtReturns struct {
		result1 time.Time
	}
	createdAtReturnsOnCall map[int]struct {
		result1 time.Time
	}
	DeleteStub        func() error
	deleteMutex       sync.RWMutex
	deleteArgsForCall []struct {
	}
	deleteReturns struct {
		result1 error
	}
	deleteReturnsOnCall map[int]struct {
		result1 error
	}
	NameStub        func() string
	nameMutex       sync.RWMutex
	nameArgsForCall []struct {
	}
	nameReturns struct {
		result1 string
	}
	nameReturnsOnCall map[int]struct {
		result1 string
	}
	OrphanedAtStub        func() time.Time
	orphanedAtMutex       sync.RWMutex
	orphanedAtArgsForCall []struct {
	}
	orphanedAtReturns struct {
		result1 time.Time
	}
	orphanedAtReturnsOnCall map[int]struct {
		result1 time.Time
	}
	TypeStub        func() string
	typeMutex       sync.RWMutex
	typeArgsForCall []struct {
	}
	typeReturns struct {
		result1 string
	}
	typeReturnsOnCall map[int]struct {
		result1 string
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeOrphanNetwork) CreatedAt() time.Time {
	fake.createdAtMutex.Lock()
	ret, specificReturn := fake.createdAtReturnsOnCall[len(fake.createdAtArgsForCall)]
	fake.createdAtArgsForCall = append(fake.createdAtArgsForCall, struct {
	}{})
	stub := fake.CreatedAtStub
	fakeReturns := fake.createdAtReturns
	fake.recordInvocation("CreatedAt", []interface{}{})
	fake.createdAtMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeOrphanNetwork) CreatedAtCallCount() int {
	fake.createdAtMutex.RLock()
	defer fake.createdAtMutex.RUnlock()
	return len(fake.createdAtArgsForCall)
}

func (fake *FakeOrphanNetwork) CreatedAtCalls(stub func() time.Time) {
	fake.createdAtMutex.Lock()
	defer fake.createdAtMutex.Unlock()
	fake.CreatedAtStub = stub
}

func (fake *FakeOrphanNetwork) CreatedAtReturns(result1 time.Time) {
	fake.createdAtMutex.Lock()
	defer fake.createdAtMutex.Unlock()
	fake.CreatedAtStub = nil
	fake.createdAtReturns = struct {
		result1 time.Time
	}{result1}
}

func (fake *FakeOrphanNetwork) CreatedAtReturnsOnCall(i int, result1 time.Time) {
	fake.createdAtMutex.Lock()
	defer fake.createdAtMutex.Unlock()
	fake.CreatedAtStub = nil
	if fake.createdAtReturnsOnCall == nil {
		fake.createdAtReturnsOnCall = make(map[int]struct {
			result1 time.Time
		})
	}
	fake.createdAtReturnsOnCall[i] = struct {
		result1 time.Time
	}{result1}
}

func (fake *FakeOrphanNetwork) Delete() error {
	fake.deleteMutex.Lock()
	ret, specificReturn := fake.deleteReturnsOnCall[len(fake.deleteArgsForCall)]
	fake.deleteArgsForCall = append(fake.deleteArgsForCall, struct {
	}{})
	stub := fake.DeleteStub
	fakeReturns := fake.deleteReturns
	fake.recordInvocation("Delete", []interface{}{})
	fake.deleteMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeOrphanNetwork) DeleteCallCount() int {
	fake.deleteMutex.RLock()
	defer fake.deleteMutex.RUnlock()
	return len(fake.deleteArgsForCall)
}

func (fake *FakeOrphanNetwork) DeleteCalls(stub func() error) {
	fake.deleteMutex.Lock()
	defer fake.deleteMutex.Unlock()
	fake.DeleteStub = stub
}

func (fake *FakeOrphanNetwork) DeleteReturns(result1 error) {
	fake.deleteMutex.Lock()
	defer fake.deleteMutex.Unlock()
	fake.DeleteStub = nil
	fake.deleteReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeOrphanNetwork) DeleteReturnsOnCall(i int, result1 error) {
	fake.deleteMutex.Lock()
	defer fake.deleteMutex.Unlock()
	fake.DeleteStub = nil
	if fake.deleteReturnsOnCall == nil {
		fake.deleteReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.deleteReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeOrphanNetwork) Name() string {
	fake.nameMutex.Lock()
	ret, specificReturn := fake.nameReturnsOnCall[len(fake.nameArgsForCall)]
	fake.nameArgsForCall = append(fake.nameArgsForCall, struct {
	}{})
	stub := fake.NameStub
	fakeReturns := fake.nameReturns
	fake.recordInvocation("Name", []interface{}{})
	fake.nameMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeOrphanNetwork) NameCallCount() int {
	fake.nameMutex.RLock()
	defer fake.nameMutex.RUnlock()
	return len(fake.nameArgsForCall)
}

func (fake *FakeOrphanNetwork) NameCalls(stub func() string) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = stub
}

func (fake *FakeOrphanNetwork) NameReturns(result1 string) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = nil
	fake.nameReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeOrphanNetwork) NameReturnsOnCall(i int, result1 string) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = nil
	if fake.nameReturnsOnCall == nil {
		fake.nameReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.nameReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *FakeOrphanNetwork) OrphanedAt() time.Time {
	fake.orphanedAtMutex.Lock()
	ret, specificReturn := fake.orphanedAtReturnsOnCall[len(fake.orphanedAtArgsForCall)]
	fake.orphanedAtArgsForCall = append(fake.orphanedAtArgsForCall, struct {
	}{})
	stub := fake.OrphanedAtStub
	fakeReturns := fake.orphanedAtReturns
	fake.recordInvocation("OrphanedAt", []interface{}{})
	fake.orphanedAtMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeOrphanNetwork) OrphanedAtCallCount() int {
	fake.orphanedAtMutex.RLock()
	defer fake.orphanedAtMutex.RUnlock()
	return len(fake.orphanedAtArgsForCall)
}

func (fake *FakeOrphanNetwork) OrphanedAtCalls(stub func() time.Time) {
	fake.orphanedAtMutex.Lock()
	defer fake.orphanedAtMutex.Unlock()
	fake.OrphanedAtStub = stub
}

func (fake *FakeOrphanNetwork) OrphanedAtReturns(result1 time.Time) {
	fake.orphanedAtMutex.Lock()
	defer fake.orphanedAtMutex.Unlock()
	fake.OrphanedAtStub = nil
	fake.orphanedAtReturns = struct {
		result1 time.Time
	}{result1}
}

func (fake *FakeOrphanNetwork) OrphanedAtReturnsOnCall(i int, result1 time.Time) {
	fake.orphanedAtMutex.Lock()
	defer fake.orphanedAtMutex.Unlock()
	fake.OrphanedAtStub = nil
	if fake.orphanedAtReturnsOnCall == nil {
		fake.orphanedAtReturnsOnCall = make(map[int]struct {
			result1 time.Time
		})
	}
	fake.orphanedAtReturnsOnCall[i] = struct {
		result1 time.Time
	}{result1}
}

func (fake *FakeOrphanNetwork) Type() string {
	fake.typeMutex.Lock()
	ret, specificReturn := fake.typeReturnsOnCall[len(fake.typeArgsForCall)]
	fake.typeArgsForCall = append(fake.typeArgsForCall, struct {
	}{})
	stub := fake.TypeStub
	fakeReturns := fake.typeReturns
	fake.recordInvocation("Type", []interface{}{})
	fake.typeMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeOrphanNetwork) TypeCallCount() int {
	fake.typeMutex.RLock()
	defer fake.typeMutex.RUnlock()
	return len(fake.typeArgsForCall)
}

func (fake *FakeOrphanNetwork) TypeCalls(stub func() string) {
	fake.typeMutex.Lock()
	defer fake.typeMutex.Unlock()
	fake.TypeStub = stub
}

func (fake *FakeOrphanNetwork) TypeReturns(result1 string) {
	fake.typeMutex.Lock()
	defer fake.typeMutex.Unlock()
	fake.TypeStub = nil
	fake.typeReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeOrphanNetwork) TypeReturnsOnCall(i int, result1 string) {
	fake.typeMutex.Lock()
	defer fake.typeMutex.Unlock()
	fake.TypeStub = nil
	if fake.typeReturnsOnCall == nil {
		fake.typeReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.typeReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *FakeOrphanNetwork) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.createdAtMutex.RLock()
	defer fake.createdAtMutex.RUnlock()
	fake.deleteMutex.RLock()
	defer fake.deleteMutex.RUnlock()
	fake.nameMutex.RLock()
	defer fake.nameMutex.RUnlock()
	fake.orphanedAtMutex.RLock()
	defer fake.orphanedAtMutex.RUnlock()
	fake.typeMutex.RLock()
	defer fake.typeMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeOrphanNetwork) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ director.OrphanNetwork = new(FakeOrphanNetwork)
