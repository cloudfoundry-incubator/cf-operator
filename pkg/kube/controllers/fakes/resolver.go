// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"context"
	"sync"

	"code.cloudfoundry.org/quarks-operator/pkg/kube/controllers/boshdeployment"
)

type FakeInterpolateSecrets struct {
	InterpolateVariableFromSecretsStub        func(context.Context, []byte, string, string) ([]byte, error)
	interpolateVariableFromSecretsMutex       sync.RWMutex
	interpolateVariableFromSecretsArgsForCall []struct {
		arg1 context.Context
		arg2 []byte
		arg3 string
		arg4 string
	}
	interpolateVariableFromSecretsReturns struct {
		result1 []byte
		result2 error
	}
	interpolateVariableFromSecretsReturnsOnCall map[int]struct {
		result1 []byte
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeInterpolateSecrets) InterpolateVariableFromSecrets(arg1 context.Context, arg2 []byte, arg3 string, arg4 string) ([]byte, error) {
	var arg2Copy []byte
	if arg2 != nil {
		arg2Copy = make([]byte, len(arg2))
		copy(arg2Copy, arg2)
	}
	fake.interpolateVariableFromSecretsMutex.Lock()
	ret, specificReturn := fake.interpolateVariableFromSecretsReturnsOnCall[len(fake.interpolateVariableFromSecretsArgsForCall)]
	fake.interpolateVariableFromSecretsArgsForCall = append(fake.interpolateVariableFromSecretsArgsForCall, struct {
		arg1 context.Context
		arg2 []byte
		arg3 string
		arg4 string
	}{arg1, arg2Copy, arg3, arg4})
	fake.recordInvocation("InterpolateVariableFromSecrets", []interface{}{arg1, arg2Copy, arg3, arg4})
	fake.interpolateVariableFromSecretsMutex.Unlock()
	if fake.InterpolateVariableFromSecretsStub != nil {
		return fake.InterpolateVariableFromSecretsStub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.interpolateVariableFromSecretsReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeInterpolateSecrets) InterpolateVariableFromSecretsCallCount() int {
	fake.interpolateVariableFromSecretsMutex.RLock()
	defer fake.interpolateVariableFromSecretsMutex.RUnlock()
	return len(fake.interpolateVariableFromSecretsArgsForCall)
}

func (fake *FakeInterpolateSecrets) InterpolateVariableFromSecretsCalls(stub func(context.Context, []byte, string, string) ([]byte, error)) {
	fake.interpolateVariableFromSecretsMutex.Lock()
	defer fake.interpolateVariableFromSecretsMutex.Unlock()
	fake.InterpolateVariableFromSecretsStub = stub
}

func (fake *FakeInterpolateSecrets) InterpolateVariableFromSecretsArgsForCall(i int) (context.Context, []byte, string, string) {
	fake.interpolateVariableFromSecretsMutex.RLock()
	defer fake.interpolateVariableFromSecretsMutex.RUnlock()
	argsForCall := fake.interpolateVariableFromSecretsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeInterpolateSecrets) InterpolateVariableFromSecretsReturns(result1 []byte, result2 error) {
	fake.interpolateVariableFromSecretsMutex.Lock()
	defer fake.interpolateVariableFromSecretsMutex.Unlock()
	fake.InterpolateVariableFromSecretsStub = nil
	fake.interpolateVariableFromSecretsReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *FakeInterpolateSecrets) InterpolateVariableFromSecretsReturnsOnCall(i int, result1 []byte, result2 error) {
	fake.interpolateVariableFromSecretsMutex.Lock()
	defer fake.interpolateVariableFromSecretsMutex.Unlock()
	fake.InterpolateVariableFromSecretsStub = nil
	if fake.interpolateVariableFromSecretsReturnsOnCall == nil {
		fake.interpolateVariableFromSecretsReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 error
		})
	}
	fake.interpolateVariableFromSecretsReturnsOnCall[i] = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *FakeInterpolateSecrets) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.interpolateVariableFromSecretsMutex.RLock()
	defer fake.interpolateVariableFromSecretsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeInterpolateSecrets) recordInvocation(key string, args []interface{}) {
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

var _ boshdeployment.InterpolateSecrets = new(FakeInterpolateSecrets)
