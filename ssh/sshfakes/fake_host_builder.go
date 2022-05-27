// Code generated by counterfeiter. DO NOT EDIT.
package sshfakes

import (
	"sync"

	"github.com/cloudfoundry/bosh-cli/v7/director"
	"github.com/cloudfoundry/bosh-cli/v7/ssh"
)

type FakeHostBuilder struct {
	BuildHostStub        func(director.AllOrInstanceGroupOrInstanceSlug, string, ssh.DeploymentFetcher) (director.Host, error)
	buildHostMutex       sync.RWMutex
	buildHostArgsForCall []struct {
		arg1 director.AllOrInstanceGroupOrInstanceSlug
		arg2 string
		arg3 ssh.DeploymentFetcher
	}
	buildHostReturns struct {
		result1 director.Host
		result2 error
	}
	buildHostReturnsOnCall map[int]struct {
		result1 director.Host
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeHostBuilder) BuildHost(arg1 director.AllOrInstanceGroupOrInstanceSlug, arg2 string, arg3 ssh.DeploymentFetcher) (director.Host, error) {
	fake.buildHostMutex.Lock()
	ret, specificReturn := fake.buildHostReturnsOnCall[len(fake.buildHostArgsForCall)]
	fake.buildHostArgsForCall = append(fake.buildHostArgsForCall, struct {
		arg1 director.AllOrInstanceGroupOrInstanceSlug
		arg2 string
		arg3 ssh.DeploymentFetcher
	}{arg1, arg2, arg3})
	stub := fake.BuildHostStub
	fakeReturns := fake.buildHostReturns
	fake.recordInvocation("BuildHost", []interface{}{arg1, arg2, arg3})
	fake.buildHostMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeHostBuilder) BuildHostCallCount() int {
	fake.buildHostMutex.RLock()
	defer fake.buildHostMutex.RUnlock()
	return len(fake.buildHostArgsForCall)
}

func (fake *FakeHostBuilder) BuildHostCalls(stub func(director.AllOrInstanceGroupOrInstanceSlug, string, ssh.DeploymentFetcher) (director.Host, error)) {
	fake.buildHostMutex.Lock()
	defer fake.buildHostMutex.Unlock()
	fake.BuildHostStub = stub
}

func (fake *FakeHostBuilder) BuildHostArgsForCall(i int) (director.AllOrInstanceGroupOrInstanceSlug, string, ssh.DeploymentFetcher) {
	fake.buildHostMutex.RLock()
	defer fake.buildHostMutex.RUnlock()
	argsForCall := fake.buildHostArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeHostBuilder) BuildHostReturns(result1 director.Host, result2 error) {
	fake.buildHostMutex.Lock()
	defer fake.buildHostMutex.Unlock()
	fake.BuildHostStub = nil
	fake.buildHostReturns = struct {
		result1 director.Host
		result2 error
	}{result1, result2}
}

func (fake *FakeHostBuilder) BuildHostReturnsOnCall(i int, result1 director.Host, result2 error) {
	fake.buildHostMutex.Lock()
	defer fake.buildHostMutex.Unlock()
	fake.BuildHostStub = nil
	if fake.buildHostReturnsOnCall == nil {
		fake.buildHostReturnsOnCall = make(map[int]struct {
			result1 director.Host
			result2 error
		})
	}
	fake.buildHostReturnsOnCall[i] = struct {
		result1 director.Host
		result2 error
	}{result1, result2}
}

func (fake *FakeHostBuilder) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.buildHostMutex.RLock()
	defer fake.buildHostMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeHostBuilder) recordInvocation(key string, args []interface{}) {
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

var _ ssh.HostBuilder = new(FakeHostBuilder)
