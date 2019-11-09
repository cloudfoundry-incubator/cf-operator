// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"code.cloudfoundry.org/cf-operator/pkg/bosh/converter"
	"code.cloudfoundry.org/cf-operator/pkg/bosh/disk"
	"code.cloudfoundry.org/cf-operator/pkg/bosh/manifest"
	v1 "k8s.io/api/core/v1"
)

type FakeContainerFactory struct {
	JobsToContainersStub        func([]manifest.Job, []v1.VolumeMount, disk.BPMResourceDisks) ([]v1.Container, error)
	jobsToContainersMutex       sync.RWMutex
	jobsToContainersArgsForCall []struct {
		arg1 []manifest.Job
		arg2 []v1.VolumeMount
		arg3 disk.BPMResourceDisks
	}
	jobsToContainersReturns struct {
		result1 []v1.Container
		result2 error
	}
	jobsToContainersReturnsOnCall map[int]struct {
		result1 []v1.Container
		result2 error
	}
	JobsToInitContainersStub        func([]manifest.Job, []v1.VolumeMount, disk.BPMResourceDisks, *string) ([]v1.Container, error)
	jobsToInitContainersMutex       sync.RWMutex
	jobsToInitContainersArgsForCall []struct {
		arg1 []manifest.Job
		arg2 []v1.VolumeMount
		arg3 disk.BPMResourceDisks
		arg4 *string
	}
	jobsToInitContainersReturns struct {
		result1 []v1.Container
		result2 error
	}
	jobsToInitContainersReturnsOnCall map[int]struct {
		result1 []v1.Container
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeContainerFactory) JobsToContainers(arg1 []manifest.Job, arg2 []v1.VolumeMount, arg3 disk.BPMResourceDisks) ([]v1.Container, error) {
	var arg1Copy []manifest.Job
	if arg1 != nil {
		arg1Copy = make([]manifest.Job, len(arg1))
		copy(arg1Copy, arg1)
	}
	var arg2Copy []v1.VolumeMount
	if arg2 != nil {
		arg2Copy = make([]v1.VolumeMount, len(arg2))
		copy(arg2Copy, arg2)
	}
	fake.jobsToContainersMutex.Lock()
	ret, specificReturn := fake.jobsToContainersReturnsOnCall[len(fake.jobsToContainersArgsForCall)]
	fake.jobsToContainersArgsForCall = append(fake.jobsToContainersArgsForCall, struct {
		arg1 []manifest.Job
		arg2 []v1.VolumeMount
		arg3 disk.BPMResourceDisks
	}{arg1Copy, arg2Copy, arg3})
	fake.recordInvocation("JobsToContainers", []interface{}{arg1Copy, arg2Copy, arg3})
	fake.jobsToContainersMutex.Unlock()
	if fake.JobsToContainersStub != nil {
		return fake.JobsToContainersStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.jobsToContainersReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeContainerFactory) JobsToContainersCallCount() int {
	fake.jobsToContainersMutex.RLock()
	defer fake.jobsToContainersMutex.RUnlock()
	return len(fake.jobsToContainersArgsForCall)
}

func (fake *FakeContainerFactory) JobsToContainersCalls(stub func([]manifest.Job, []v1.VolumeMount, disk.BPMResourceDisks) ([]v1.Container, error)) {
	fake.jobsToContainersMutex.Lock()
	defer fake.jobsToContainersMutex.Unlock()
	fake.JobsToContainersStub = stub
}

func (fake *FakeContainerFactory) JobsToContainersArgsForCall(i int) ([]manifest.Job, []v1.VolumeMount, disk.BPMResourceDisks) {
	fake.jobsToContainersMutex.RLock()
	defer fake.jobsToContainersMutex.RUnlock()
	argsForCall := fake.jobsToContainersArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeContainerFactory) JobsToContainersReturns(result1 []v1.Container, result2 error) {
	fake.jobsToContainersMutex.Lock()
	defer fake.jobsToContainersMutex.Unlock()
	fake.JobsToContainersStub = nil
	fake.jobsToContainersReturns = struct {
		result1 []v1.Container
		result2 error
	}{result1, result2}
}

func (fake *FakeContainerFactory) JobsToContainersReturnsOnCall(i int, result1 []v1.Container, result2 error) {
	fake.jobsToContainersMutex.Lock()
	defer fake.jobsToContainersMutex.Unlock()
	fake.JobsToContainersStub = nil
	if fake.jobsToContainersReturnsOnCall == nil {
		fake.jobsToContainersReturnsOnCall = make(map[int]struct {
			result1 []v1.Container
			result2 error
		})
	}
	fake.jobsToContainersReturnsOnCall[i] = struct {
		result1 []v1.Container
		result2 error
	}{result1, result2}
}

func (fake *FakeContainerFactory) JobsToInitContainers(arg1 []manifest.Job, arg2 []v1.VolumeMount, arg3 disk.BPMResourceDisks, arg4 *string) ([]v1.Container, error) {
	var arg1Copy []manifest.Job
	if arg1 != nil {
		arg1Copy = make([]manifest.Job, len(arg1))
		copy(arg1Copy, arg1)
	}
	var arg2Copy []v1.VolumeMount
	if arg2 != nil {
		arg2Copy = make([]v1.VolumeMount, len(arg2))
		copy(arg2Copy, arg2)
	}
	fake.jobsToInitContainersMutex.Lock()
	ret, specificReturn := fake.jobsToInitContainersReturnsOnCall[len(fake.jobsToInitContainersArgsForCall)]
	fake.jobsToInitContainersArgsForCall = append(fake.jobsToInitContainersArgsForCall, struct {
		arg1 []manifest.Job
		arg2 []v1.VolumeMount
		arg3 disk.BPMResourceDisks
		arg4 *string
	}{arg1Copy, arg2Copy, arg3, arg4})
	fake.recordInvocation("JobsToInitContainers", []interface{}{arg1Copy, arg2Copy, arg3, arg4})
	fake.jobsToInitContainersMutex.Unlock()
	if fake.JobsToInitContainersStub != nil {
		return fake.JobsToInitContainersStub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.jobsToInitContainersReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeContainerFactory) JobsToInitContainersCallCount() int {
	fake.jobsToInitContainersMutex.RLock()
	defer fake.jobsToInitContainersMutex.RUnlock()
	return len(fake.jobsToInitContainersArgsForCall)
}

func (fake *FakeContainerFactory) JobsToInitContainersCalls(stub func([]manifest.Job, []v1.VolumeMount, disk.BPMResourceDisks, *string) ([]v1.Container, error)) {
	fake.jobsToInitContainersMutex.Lock()
	defer fake.jobsToInitContainersMutex.Unlock()
	fake.JobsToInitContainersStub = stub
}

func (fake *FakeContainerFactory) JobsToInitContainersArgsForCall(i int) ([]manifest.Job, []v1.VolumeMount, disk.BPMResourceDisks, *string) {
	fake.jobsToInitContainersMutex.RLock()
	defer fake.jobsToInitContainersMutex.RUnlock()
	argsForCall := fake.jobsToInitContainersArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeContainerFactory) JobsToInitContainersReturns(result1 []v1.Container, result2 error) {
	fake.jobsToInitContainersMutex.Lock()
	defer fake.jobsToInitContainersMutex.Unlock()
	fake.JobsToInitContainersStub = nil
	fake.jobsToInitContainersReturns = struct {
		result1 []v1.Container
		result2 error
	}{result1, result2}
}

func (fake *FakeContainerFactory) JobsToInitContainersReturnsOnCall(i int, result1 []v1.Container, result2 error) {
	fake.jobsToInitContainersMutex.Lock()
	defer fake.jobsToInitContainersMutex.Unlock()
	fake.JobsToInitContainersStub = nil
	if fake.jobsToInitContainersReturnsOnCall == nil {
		fake.jobsToInitContainersReturnsOnCall = make(map[int]struct {
			result1 []v1.Container
			result2 error
		})
	}
	fake.jobsToInitContainersReturnsOnCall[i] = struct {
		result1 []v1.Container
		result2 error
	}{result1, result2}
}

func (fake *FakeContainerFactory) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.jobsToContainersMutex.RLock()
	defer fake.jobsToContainersMutex.RUnlock()
	fake.jobsToInitContainersMutex.RLock()
	defer fake.jobsToInitContainersMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeContainerFactory) recordInvocation(key string, args []interface{}) {
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

var _ converter.ContainerFactory = new(FakeContainerFactory)
