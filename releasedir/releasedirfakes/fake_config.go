// Code generated by counterfeiter. DO NOT EDIT.
package releasedirfakes

import (
	"sync"

	"github.com/cloudfoundry/bosh-cli/v7/releasedir"
)

type FakeConfig struct {
	BlobstoreStub        func() (string, map[string]interface{}, error)
	blobstoreMutex       sync.RWMutex
	blobstoreArgsForCall []struct {
	}
	blobstoreReturns struct {
		result1 string
		result2 map[string]interface{}
		result3 error
	}
	blobstoreReturnsOnCall map[int]struct {
		result1 string
		result2 map[string]interface{}
		result3 error
	}
	NameStub        func() (string, error)
	nameMutex       sync.RWMutex
	nameArgsForCall []struct {
	}
	nameReturns struct {
		result1 string
		result2 error
	}
	nameReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	SaveNameStub        func(string) error
	saveNameMutex       sync.RWMutex
	saveNameArgsForCall []struct {
		arg1 string
	}
	saveNameReturns struct {
		result1 error
	}
	saveNameReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeConfig) Blobstore() (string, map[string]interface{}, error) {
	fake.blobstoreMutex.Lock()
	ret, specificReturn := fake.blobstoreReturnsOnCall[len(fake.blobstoreArgsForCall)]
	fake.blobstoreArgsForCall = append(fake.blobstoreArgsForCall, struct {
	}{})
	stub := fake.BlobstoreStub
	fakeReturns := fake.blobstoreReturns
	fake.recordInvocation("Blobstore", []interface{}{})
	fake.blobstoreMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakeConfig) BlobstoreCallCount() int {
	fake.blobstoreMutex.RLock()
	defer fake.blobstoreMutex.RUnlock()
	return len(fake.blobstoreArgsForCall)
}

func (fake *FakeConfig) BlobstoreCalls(stub func() (string, map[string]interface{}, error)) {
	fake.blobstoreMutex.Lock()
	defer fake.blobstoreMutex.Unlock()
	fake.BlobstoreStub = stub
}

func (fake *FakeConfig) BlobstoreReturns(result1 string, result2 map[string]interface{}, result3 error) {
	fake.blobstoreMutex.Lock()
	defer fake.blobstoreMutex.Unlock()
	fake.BlobstoreStub = nil
	fake.blobstoreReturns = struct {
		result1 string
		result2 map[string]interface{}
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeConfig) BlobstoreReturnsOnCall(i int, result1 string, result2 map[string]interface{}, result3 error) {
	fake.blobstoreMutex.Lock()
	defer fake.blobstoreMutex.Unlock()
	fake.BlobstoreStub = nil
	if fake.blobstoreReturnsOnCall == nil {
		fake.blobstoreReturnsOnCall = make(map[int]struct {
			result1 string
			result2 map[string]interface{}
			result3 error
		})
	}
	fake.blobstoreReturnsOnCall[i] = struct {
		result1 string
		result2 map[string]interface{}
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeConfig) Name() (string, error) {
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
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeConfig) NameCallCount() int {
	fake.nameMutex.RLock()
	defer fake.nameMutex.RUnlock()
	return len(fake.nameArgsForCall)
}

func (fake *FakeConfig) NameCalls(stub func() (string, error)) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = stub
}

func (fake *FakeConfig) NameReturns(result1 string, result2 error) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = nil
	fake.nameReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeConfig) NameReturnsOnCall(i int, result1 string, result2 error) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = nil
	if fake.nameReturnsOnCall == nil {
		fake.nameReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.nameReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeConfig) SaveName(arg1 string) error {
	fake.saveNameMutex.Lock()
	ret, specificReturn := fake.saveNameReturnsOnCall[len(fake.saveNameArgsForCall)]
	fake.saveNameArgsForCall = append(fake.saveNameArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.SaveNameStub
	fakeReturns := fake.saveNameReturns
	fake.recordInvocation("SaveName", []interface{}{arg1})
	fake.saveNameMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeConfig) SaveNameCallCount() int {
	fake.saveNameMutex.RLock()
	defer fake.saveNameMutex.RUnlock()
	return len(fake.saveNameArgsForCall)
}

func (fake *FakeConfig) SaveNameCalls(stub func(string) error) {
	fake.saveNameMutex.Lock()
	defer fake.saveNameMutex.Unlock()
	fake.SaveNameStub = stub
}

func (fake *FakeConfig) SaveNameArgsForCall(i int) string {
	fake.saveNameMutex.RLock()
	defer fake.saveNameMutex.RUnlock()
	argsForCall := fake.saveNameArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeConfig) SaveNameReturns(result1 error) {
	fake.saveNameMutex.Lock()
	defer fake.saveNameMutex.Unlock()
	fake.SaveNameStub = nil
	fake.saveNameReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeConfig) SaveNameReturnsOnCall(i int, result1 error) {
	fake.saveNameMutex.Lock()
	defer fake.saveNameMutex.Unlock()
	fake.SaveNameStub = nil
	if fake.saveNameReturnsOnCall == nil {
		fake.saveNameReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.saveNameReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeConfig) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.blobstoreMutex.RLock()
	defer fake.blobstoreMutex.RUnlock()
	fake.nameMutex.RLock()
	defer fake.nameMutex.RUnlock()
	fake.saveNameMutex.RLock()
	defer fake.saveNameMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeConfig) recordInvocation(key string, args []interface{}) {
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

var _ releasedir.Config = new(FakeConfig)
