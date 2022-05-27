// Code generated by counterfeiter. DO NOT EDIT.
package cmdfakes

import (
	"sync"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/config"
	"github.com/cloudfoundry/bosh-cli/v7/director"
	"github.com/cloudfoundry/bosh-cli/v7/uaa"
)

type FakeSession struct {
	AnonymousDirectorStub        func() (director.Director, error)
	anonymousDirectorMutex       sync.RWMutex
	anonymousDirectorArgsForCall []struct {
	}
	anonymousDirectorReturns struct {
		result1 director.Director
		result2 error
	}
	anonymousDirectorReturnsOnCall map[int]struct {
		result1 director.Director
		result2 error
	}
	CredentialsStub        func() config.Creds
	credentialsMutex       sync.RWMutex
	credentialsArgsForCall []struct {
	}
	credentialsReturns struct {
		result1 config.Creds
	}
	credentialsReturnsOnCall map[int]struct {
		result1 config.Creds
	}
	DeploymentStub        func() (director.Deployment, error)
	deploymentMutex       sync.RWMutex
	deploymentArgsForCall []struct {
	}
	deploymentReturns struct {
		result1 director.Deployment
		result2 error
	}
	deploymentReturnsOnCall map[int]struct {
		result1 director.Deployment
		result2 error
	}
	DirectorStub        func() (director.Director, error)
	directorMutex       sync.RWMutex
	directorArgsForCall []struct {
	}
	directorReturns struct {
		result1 director.Director
		result2 error
	}
	directorReturnsOnCall map[int]struct {
		result1 director.Director
		result2 error
	}
	EnvironmentStub        func() string
	environmentMutex       sync.RWMutex
	environmentArgsForCall []struct {
	}
	environmentReturns struct {
		result1 string
	}
	environmentReturnsOnCall map[int]struct {
		result1 string
	}
	UAAStub        func() (uaa.UAA, error)
	uAAMutex       sync.RWMutex
	uAAArgsForCall []struct {
	}
	uAAReturns struct {
		result1 uaa.UAA
		result2 error
	}
	uAAReturnsOnCall map[int]struct {
		result1 uaa.UAA
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeSession) AnonymousDirector() (director.Director, error) {
	fake.anonymousDirectorMutex.Lock()
	ret, specificReturn := fake.anonymousDirectorReturnsOnCall[len(fake.anonymousDirectorArgsForCall)]
	fake.anonymousDirectorArgsForCall = append(fake.anonymousDirectorArgsForCall, struct {
	}{})
	stub := fake.AnonymousDirectorStub
	fakeReturns := fake.anonymousDirectorReturns
	fake.recordInvocation("AnonymousDirector", []interface{}{})
	fake.anonymousDirectorMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeSession) AnonymousDirectorCallCount() int {
	fake.anonymousDirectorMutex.RLock()
	defer fake.anonymousDirectorMutex.RUnlock()
	return len(fake.anonymousDirectorArgsForCall)
}

func (fake *FakeSession) AnonymousDirectorCalls(stub func() (director.Director, error)) {
	fake.anonymousDirectorMutex.Lock()
	defer fake.anonymousDirectorMutex.Unlock()
	fake.AnonymousDirectorStub = stub
}

func (fake *FakeSession) AnonymousDirectorReturns(result1 director.Director, result2 error) {
	fake.anonymousDirectorMutex.Lock()
	defer fake.anonymousDirectorMutex.Unlock()
	fake.AnonymousDirectorStub = nil
	fake.anonymousDirectorReturns = struct {
		result1 director.Director
		result2 error
	}{result1, result2}
}

func (fake *FakeSession) AnonymousDirectorReturnsOnCall(i int, result1 director.Director, result2 error) {
	fake.anonymousDirectorMutex.Lock()
	defer fake.anonymousDirectorMutex.Unlock()
	fake.AnonymousDirectorStub = nil
	if fake.anonymousDirectorReturnsOnCall == nil {
		fake.anonymousDirectorReturnsOnCall = make(map[int]struct {
			result1 director.Director
			result2 error
		})
	}
	fake.anonymousDirectorReturnsOnCall[i] = struct {
		result1 director.Director
		result2 error
	}{result1, result2}
}

func (fake *FakeSession) Credentials() config.Creds {
	fake.credentialsMutex.Lock()
	ret, specificReturn := fake.credentialsReturnsOnCall[len(fake.credentialsArgsForCall)]
	fake.credentialsArgsForCall = append(fake.credentialsArgsForCall, struct {
	}{})
	stub := fake.CredentialsStub
	fakeReturns := fake.credentialsReturns
	fake.recordInvocation("Credentials", []interface{}{})
	fake.credentialsMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeSession) CredentialsCallCount() int {
	fake.credentialsMutex.RLock()
	defer fake.credentialsMutex.RUnlock()
	return len(fake.credentialsArgsForCall)
}

func (fake *FakeSession) CredentialsCalls(stub func() config.Creds) {
	fake.credentialsMutex.Lock()
	defer fake.credentialsMutex.Unlock()
	fake.CredentialsStub = stub
}

func (fake *FakeSession) CredentialsReturns(result1 config.Creds) {
	fake.credentialsMutex.Lock()
	defer fake.credentialsMutex.Unlock()
	fake.CredentialsStub = nil
	fake.credentialsReturns = struct {
		result1 config.Creds
	}{result1}
}

func (fake *FakeSession) CredentialsReturnsOnCall(i int, result1 config.Creds) {
	fake.credentialsMutex.Lock()
	defer fake.credentialsMutex.Unlock()
	fake.CredentialsStub = nil
	if fake.credentialsReturnsOnCall == nil {
		fake.credentialsReturnsOnCall = make(map[int]struct {
			result1 config.Creds
		})
	}
	fake.credentialsReturnsOnCall[i] = struct {
		result1 config.Creds
	}{result1}
}

func (fake *FakeSession) Deployment() (director.Deployment, error) {
	fake.deploymentMutex.Lock()
	ret, specificReturn := fake.deploymentReturnsOnCall[len(fake.deploymentArgsForCall)]
	fake.deploymentArgsForCall = append(fake.deploymentArgsForCall, struct {
	}{})
	stub := fake.DeploymentStub
	fakeReturns := fake.deploymentReturns
	fake.recordInvocation("Deployment", []interface{}{})
	fake.deploymentMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeSession) DeploymentCallCount() int {
	fake.deploymentMutex.RLock()
	defer fake.deploymentMutex.RUnlock()
	return len(fake.deploymentArgsForCall)
}

func (fake *FakeSession) DeploymentCalls(stub func() (director.Deployment, error)) {
	fake.deploymentMutex.Lock()
	defer fake.deploymentMutex.Unlock()
	fake.DeploymentStub = stub
}

func (fake *FakeSession) DeploymentReturns(result1 director.Deployment, result2 error) {
	fake.deploymentMutex.Lock()
	defer fake.deploymentMutex.Unlock()
	fake.DeploymentStub = nil
	fake.deploymentReturns = struct {
		result1 director.Deployment
		result2 error
	}{result1, result2}
}

func (fake *FakeSession) DeploymentReturnsOnCall(i int, result1 director.Deployment, result2 error) {
	fake.deploymentMutex.Lock()
	defer fake.deploymentMutex.Unlock()
	fake.DeploymentStub = nil
	if fake.deploymentReturnsOnCall == nil {
		fake.deploymentReturnsOnCall = make(map[int]struct {
			result1 director.Deployment
			result2 error
		})
	}
	fake.deploymentReturnsOnCall[i] = struct {
		result1 director.Deployment
		result2 error
	}{result1, result2}
}

func (fake *FakeSession) Director() (director.Director, error) {
	fake.directorMutex.Lock()
	ret, specificReturn := fake.directorReturnsOnCall[len(fake.directorArgsForCall)]
	fake.directorArgsForCall = append(fake.directorArgsForCall, struct {
	}{})
	stub := fake.DirectorStub
	fakeReturns := fake.directorReturns
	fake.recordInvocation("Director", []interface{}{})
	fake.directorMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeSession) DirectorCallCount() int {
	fake.directorMutex.RLock()
	defer fake.directorMutex.RUnlock()
	return len(fake.directorArgsForCall)
}

func (fake *FakeSession) DirectorCalls(stub func() (director.Director, error)) {
	fake.directorMutex.Lock()
	defer fake.directorMutex.Unlock()
	fake.DirectorStub = stub
}

func (fake *FakeSession) DirectorReturns(result1 director.Director, result2 error) {
	fake.directorMutex.Lock()
	defer fake.directorMutex.Unlock()
	fake.DirectorStub = nil
	fake.directorReturns = struct {
		result1 director.Director
		result2 error
	}{result1, result2}
}

func (fake *FakeSession) DirectorReturnsOnCall(i int, result1 director.Director, result2 error) {
	fake.directorMutex.Lock()
	defer fake.directorMutex.Unlock()
	fake.DirectorStub = nil
	if fake.directorReturnsOnCall == nil {
		fake.directorReturnsOnCall = make(map[int]struct {
			result1 director.Director
			result2 error
		})
	}
	fake.directorReturnsOnCall[i] = struct {
		result1 director.Director
		result2 error
	}{result1, result2}
}

func (fake *FakeSession) Environment() string {
	fake.environmentMutex.Lock()
	ret, specificReturn := fake.environmentReturnsOnCall[len(fake.environmentArgsForCall)]
	fake.environmentArgsForCall = append(fake.environmentArgsForCall, struct {
	}{})
	stub := fake.EnvironmentStub
	fakeReturns := fake.environmentReturns
	fake.recordInvocation("Environment", []interface{}{})
	fake.environmentMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeSession) EnvironmentCallCount() int {
	fake.environmentMutex.RLock()
	defer fake.environmentMutex.RUnlock()
	return len(fake.environmentArgsForCall)
}

func (fake *FakeSession) EnvironmentCalls(stub func() string) {
	fake.environmentMutex.Lock()
	defer fake.environmentMutex.Unlock()
	fake.EnvironmentStub = stub
}

func (fake *FakeSession) EnvironmentReturns(result1 string) {
	fake.environmentMutex.Lock()
	defer fake.environmentMutex.Unlock()
	fake.EnvironmentStub = nil
	fake.environmentReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeSession) EnvironmentReturnsOnCall(i int, result1 string) {
	fake.environmentMutex.Lock()
	defer fake.environmentMutex.Unlock()
	fake.EnvironmentStub = nil
	if fake.environmentReturnsOnCall == nil {
		fake.environmentReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.environmentReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *FakeSession) UAA() (uaa.UAA, error) {
	fake.uAAMutex.Lock()
	ret, specificReturn := fake.uAAReturnsOnCall[len(fake.uAAArgsForCall)]
	fake.uAAArgsForCall = append(fake.uAAArgsForCall, struct {
	}{})
	stub := fake.UAAStub
	fakeReturns := fake.uAAReturns
	fake.recordInvocation("UAA", []interface{}{})
	fake.uAAMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeSession) UAACallCount() int {
	fake.uAAMutex.RLock()
	defer fake.uAAMutex.RUnlock()
	return len(fake.uAAArgsForCall)
}

func (fake *FakeSession) UAACalls(stub func() (uaa.UAA, error)) {
	fake.uAAMutex.Lock()
	defer fake.uAAMutex.Unlock()
	fake.UAAStub = stub
}

func (fake *FakeSession) UAAReturns(result1 uaa.UAA, result2 error) {
	fake.uAAMutex.Lock()
	defer fake.uAAMutex.Unlock()
	fake.UAAStub = nil
	fake.uAAReturns = struct {
		result1 uaa.UAA
		result2 error
	}{result1, result2}
}

func (fake *FakeSession) UAAReturnsOnCall(i int, result1 uaa.UAA, result2 error) {
	fake.uAAMutex.Lock()
	defer fake.uAAMutex.Unlock()
	fake.UAAStub = nil
	if fake.uAAReturnsOnCall == nil {
		fake.uAAReturnsOnCall = make(map[int]struct {
			result1 uaa.UAA
			result2 error
		})
	}
	fake.uAAReturnsOnCall[i] = struct {
		result1 uaa.UAA
		result2 error
	}{result1, result2}
}

func (fake *FakeSession) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.anonymousDirectorMutex.RLock()
	defer fake.anonymousDirectorMutex.RUnlock()
	fake.credentialsMutex.RLock()
	defer fake.credentialsMutex.RUnlock()
	fake.deploymentMutex.RLock()
	defer fake.deploymentMutex.RUnlock()
	fake.directorMutex.RLock()
	defer fake.directorMutex.RUnlock()
	fake.environmentMutex.RLock()
	defer fake.environmentMutex.RUnlock()
	fake.uAAMutex.RLock()
	defer fake.uAAMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeSession) recordInvocation(key string, args []interface{}) {
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

var _ cmd.Session = new(FakeSession)
