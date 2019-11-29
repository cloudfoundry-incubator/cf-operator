// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"context"
	"sync"

	"code.cloudfoundry.org/cf-operator/pkg/bosh/converter"
	"code.cloudfoundry.org/cf-operator/pkg/bosh/manifest"
)

type FakeDesiredManifest struct {
	DesiredManifestStub        func(context.Context, string, string) (*manifest.Manifest, error)
	desiredManifestMutex       sync.RWMutex
	desiredManifestArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 string
	}
	desiredManifestReturns struct {
		result1 *manifest.Manifest
		result2 error
	}
	desiredManifestReturnsOnCall map[int]struct {
		result1 *manifest.Manifest
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeDesiredManifest) DesiredManifest(arg1 context.Context, arg2 string, arg3 string) (*manifest.Manifest, error) {
	fake.desiredManifestMutex.Lock()
	ret, specificReturn := fake.desiredManifestReturnsOnCall[len(fake.desiredManifestArgsForCall)]
	fake.desiredManifestArgsForCall = append(fake.desiredManifestArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 string
	}{arg1, arg2, arg3})
	fake.recordInvocation("DesiredManifest", []interface{}{arg1, arg2, arg3})
	fake.desiredManifestMutex.Unlock()
	if fake.DesiredManifestStub != nil {
		return fake.DesiredManifestStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.desiredManifestReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeDesiredManifest) DesiredManifestCallCount() int {
	fake.desiredManifestMutex.RLock()
	defer fake.desiredManifestMutex.RUnlock()
	return len(fake.desiredManifestArgsForCall)
}

func (fake *FakeDesiredManifest) DesiredManifestCalls(stub func(context.Context, string, string) (*manifest.Manifest, error)) {
	fake.desiredManifestMutex.Lock()
	defer fake.desiredManifestMutex.Unlock()
	fake.DesiredManifestStub = stub
}

func (fake *FakeDesiredManifest) DesiredManifestArgsForCall(i int) (context.Context, string, string) {
	fake.desiredManifestMutex.RLock()
	defer fake.desiredManifestMutex.RUnlock()
	argsForCall := fake.desiredManifestArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeDesiredManifest) DesiredManifestReturns(result1 *manifest.Manifest, result2 error) {
	fake.desiredManifestMutex.Lock()
	defer fake.desiredManifestMutex.Unlock()
	fake.DesiredManifestStub = nil
	fake.desiredManifestReturns = struct {
		result1 *manifest.Manifest
		result2 error
	}{result1, result2}
}

func (fake *FakeDesiredManifest) DesiredManifestReturnsOnCall(i int, result1 *manifest.Manifest, result2 error) {
	fake.desiredManifestMutex.Lock()
	defer fake.desiredManifestMutex.Unlock()
	fake.DesiredManifestStub = nil
	if fake.desiredManifestReturnsOnCall == nil {
		fake.desiredManifestReturnsOnCall = make(map[int]struct {
			result1 *manifest.Manifest
			result2 error
		})
	}
	fake.desiredManifestReturnsOnCall[i] = struct {
		result1 *manifest.Manifest
		result2 error
	}{result1, result2}
}

func (fake *FakeDesiredManifest) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.desiredManifestMutex.RLock()
	defer fake.desiredManifestMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeDesiredManifest) recordInvocation(key string, args []interface{}) {
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

var _ converter.DesiredManifest = new(FakeDesiredManifest)
