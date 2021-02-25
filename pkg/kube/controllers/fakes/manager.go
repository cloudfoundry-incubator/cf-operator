// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"context"
	"net/http"
	"sync"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

type FakeManager struct {
	AddStub        func(manager.Runnable) error
	addMutex       sync.RWMutex
	addArgsForCall []struct {
		arg1 manager.Runnable
	}
	addReturns struct {
		result1 error
	}
	addReturnsOnCall map[int]struct {
		result1 error
	}
	AddHealthzCheckStub        func(string, healthz.Checker) error
	addHealthzCheckMutex       sync.RWMutex
	addHealthzCheckArgsForCall []struct {
		arg1 string
		arg2 healthz.Checker
	}
	addHealthzCheckReturns struct {
		result1 error
	}
	addHealthzCheckReturnsOnCall map[int]struct {
		result1 error
	}
	AddMetricsExtraHandlerStub        func(string, http.Handler) error
	addMetricsExtraHandlerMutex       sync.RWMutex
	addMetricsExtraHandlerArgsForCall []struct {
		arg1 string
		arg2 http.Handler
	}
	addMetricsExtraHandlerReturns struct {
		result1 error
	}
	addMetricsExtraHandlerReturnsOnCall map[int]struct {
		result1 error
	}
	AddReadyzCheckStub        func(string, healthz.Checker) error
	addReadyzCheckMutex       sync.RWMutex
	addReadyzCheckArgsForCall []struct {
		arg1 string
		arg2 healthz.Checker
	}
	addReadyzCheckReturns struct {
		result1 error
	}
	addReadyzCheckReturnsOnCall map[int]struct {
		result1 error
	}
	ElectedStub        func() <-chan struct{}
	electedMutex       sync.RWMutex
	electedArgsForCall []struct {
	}
	electedReturns struct {
		result1 <-chan struct{}
	}
	electedReturnsOnCall map[int]struct {
		result1 <-chan struct{}
	}
	GetAPIReaderStub        func() client.Reader
	getAPIReaderMutex       sync.RWMutex
	getAPIReaderArgsForCall []struct {
	}
	getAPIReaderReturns struct {
		result1 client.Reader
	}
	getAPIReaderReturnsOnCall map[int]struct {
		result1 client.Reader
	}
	GetCacheStub        func() cache.Cache
	getCacheMutex       sync.RWMutex
	getCacheArgsForCall []struct {
	}
	getCacheReturns struct {
		result1 cache.Cache
	}
	getCacheReturnsOnCall map[int]struct {
		result1 cache.Cache
	}
	GetClientStub        func() client.Client
	getClientMutex       sync.RWMutex
	getClientArgsForCall []struct {
	}
	getClientReturns struct {
		result1 client.Client
	}
	getClientReturnsOnCall map[int]struct {
		result1 client.Client
	}
	GetConfigStub        func() *rest.Config
	getConfigMutex       sync.RWMutex
	getConfigArgsForCall []struct {
	}
	getConfigReturns struct {
		result1 *rest.Config
	}
	getConfigReturnsOnCall map[int]struct {
		result1 *rest.Config
	}
	GetEventRecorderForStub        func(string) record.EventRecorder
	getEventRecorderForMutex       sync.RWMutex
	getEventRecorderForArgsForCall []struct {
		arg1 string
	}
	getEventRecorderForReturns struct {
		result1 record.EventRecorder
	}
	getEventRecorderForReturnsOnCall map[int]struct {
		result1 record.EventRecorder
	}
	GetFieldIndexerStub        func() client.FieldIndexer
	getFieldIndexerMutex       sync.RWMutex
	getFieldIndexerArgsForCall []struct {
	}
	getFieldIndexerReturns struct {
		result1 client.FieldIndexer
	}
	getFieldIndexerReturnsOnCall map[int]struct {
		result1 client.FieldIndexer
	}
	GetLoggerStub        func() logr.Logger
	getLoggerMutex       sync.RWMutex
	getLoggerArgsForCall []struct {
	}
	getLoggerReturns struct {
		result1 logr.Logger
	}
	getLoggerReturnsOnCall map[int]struct {
		result1 logr.Logger
	}
	GetRESTMapperStub        func() meta.RESTMapper
	getRESTMapperMutex       sync.RWMutex
	getRESTMapperArgsForCall []struct {
	}
	getRESTMapperReturns struct {
		result1 meta.RESTMapper
	}
	getRESTMapperReturnsOnCall map[int]struct {
		result1 meta.RESTMapper
	}
	GetSchemeStub        func() *runtime.Scheme
	getSchemeMutex       sync.RWMutex
	getSchemeArgsForCall []struct {
	}
	getSchemeReturns struct {
		result1 *runtime.Scheme
	}
	getSchemeReturnsOnCall map[int]struct {
		result1 *runtime.Scheme
	}
	GetWebhookServerStub        func() *webhook.Server
	getWebhookServerMutex       sync.RWMutex
	getWebhookServerArgsForCall []struct {
	}
	getWebhookServerReturns struct {
		result1 *webhook.Server
	}
	getWebhookServerReturnsOnCall map[int]struct {
		result1 *webhook.Server
	}
	SetFieldsStub        func(interface{}) error
	setFieldsMutex       sync.RWMutex
	setFieldsArgsForCall []struct {
		arg1 interface{}
	}
	setFieldsReturns struct {
		result1 error
	}
	setFieldsReturnsOnCall map[int]struct {
		result1 error
	}
	StartStub        func(context.Context) error
	startMutex       sync.RWMutex
	startArgsForCall []struct {
		arg1 context.Context
	}
	startReturns struct {
		result1 error
	}
	startReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeManager) Add(arg1 manager.Runnable) error {
	fake.addMutex.Lock()
	ret, specificReturn := fake.addReturnsOnCall[len(fake.addArgsForCall)]
	fake.addArgsForCall = append(fake.addArgsForCall, struct {
		arg1 manager.Runnable
	}{arg1})
	fake.recordInvocation("Add", []interface{}{arg1})
	fake.addMutex.Unlock()
	if fake.AddStub != nil {
		return fake.AddStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.addReturns
	return fakeReturns.result1
}

func (fake *FakeManager) AddCallCount() int {
	fake.addMutex.RLock()
	defer fake.addMutex.RUnlock()
	return len(fake.addArgsForCall)
}

func (fake *FakeManager) AddCalls(stub func(manager.Runnable) error) {
	fake.addMutex.Lock()
	defer fake.addMutex.Unlock()
	fake.AddStub = stub
}

func (fake *FakeManager) AddArgsForCall(i int) manager.Runnable {
	fake.addMutex.RLock()
	defer fake.addMutex.RUnlock()
	argsForCall := fake.addArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeManager) AddReturns(result1 error) {
	fake.addMutex.Lock()
	defer fake.addMutex.Unlock()
	fake.AddStub = nil
	fake.addReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeManager) AddReturnsOnCall(i int, result1 error) {
	fake.addMutex.Lock()
	defer fake.addMutex.Unlock()
	fake.AddStub = nil
	if fake.addReturnsOnCall == nil {
		fake.addReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.addReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeManager) AddHealthzCheck(arg1 string, arg2 healthz.Checker) error {
	fake.addHealthzCheckMutex.Lock()
	ret, specificReturn := fake.addHealthzCheckReturnsOnCall[len(fake.addHealthzCheckArgsForCall)]
	fake.addHealthzCheckArgsForCall = append(fake.addHealthzCheckArgsForCall, struct {
		arg1 string
		arg2 healthz.Checker
	}{arg1, arg2})
	fake.recordInvocation("AddHealthzCheck", []interface{}{arg1, arg2})
	fake.addHealthzCheckMutex.Unlock()
	if fake.AddHealthzCheckStub != nil {
		return fake.AddHealthzCheckStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.addHealthzCheckReturns
	return fakeReturns.result1
}

func (fake *FakeManager) AddHealthzCheckCallCount() int {
	fake.addHealthzCheckMutex.RLock()
	defer fake.addHealthzCheckMutex.RUnlock()
	return len(fake.addHealthzCheckArgsForCall)
}

func (fake *FakeManager) AddHealthzCheckCalls(stub func(string, healthz.Checker) error) {
	fake.addHealthzCheckMutex.Lock()
	defer fake.addHealthzCheckMutex.Unlock()
	fake.AddHealthzCheckStub = stub
}

func (fake *FakeManager) AddHealthzCheckArgsForCall(i int) (string, healthz.Checker) {
	fake.addHealthzCheckMutex.RLock()
	defer fake.addHealthzCheckMutex.RUnlock()
	argsForCall := fake.addHealthzCheckArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeManager) AddHealthzCheckReturns(result1 error) {
	fake.addHealthzCheckMutex.Lock()
	defer fake.addHealthzCheckMutex.Unlock()
	fake.AddHealthzCheckStub = nil
	fake.addHealthzCheckReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeManager) AddHealthzCheckReturnsOnCall(i int, result1 error) {
	fake.addHealthzCheckMutex.Lock()
	defer fake.addHealthzCheckMutex.Unlock()
	fake.AddHealthzCheckStub = nil
	if fake.addHealthzCheckReturnsOnCall == nil {
		fake.addHealthzCheckReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.addHealthzCheckReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeManager) AddMetricsExtraHandler(arg1 string, arg2 http.Handler) error {
	fake.addMetricsExtraHandlerMutex.Lock()
	ret, specificReturn := fake.addMetricsExtraHandlerReturnsOnCall[len(fake.addMetricsExtraHandlerArgsForCall)]
	fake.addMetricsExtraHandlerArgsForCall = append(fake.addMetricsExtraHandlerArgsForCall, struct {
		arg1 string
		arg2 http.Handler
	}{arg1, arg2})
	fake.recordInvocation("AddMetricsExtraHandler", []interface{}{arg1, arg2})
	fake.addMetricsExtraHandlerMutex.Unlock()
	if fake.AddMetricsExtraHandlerStub != nil {
		return fake.AddMetricsExtraHandlerStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.addMetricsExtraHandlerReturns
	return fakeReturns.result1
}

func (fake *FakeManager) AddMetricsExtraHandlerCallCount() int {
	fake.addMetricsExtraHandlerMutex.RLock()
	defer fake.addMetricsExtraHandlerMutex.RUnlock()
	return len(fake.addMetricsExtraHandlerArgsForCall)
}

func (fake *FakeManager) AddMetricsExtraHandlerCalls(stub func(string, http.Handler) error) {
	fake.addMetricsExtraHandlerMutex.Lock()
	defer fake.addMetricsExtraHandlerMutex.Unlock()
	fake.AddMetricsExtraHandlerStub = stub
}

func (fake *FakeManager) AddMetricsExtraHandlerArgsForCall(i int) (string, http.Handler) {
	fake.addMetricsExtraHandlerMutex.RLock()
	defer fake.addMetricsExtraHandlerMutex.RUnlock()
	argsForCall := fake.addMetricsExtraHandlerArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeManager) AddMetricsExtraHandlerReturns(result1 error) {
	fake.addMetricsExtraHandlerMutex.Lock()
	defer fake.addMetricsExtraHandlerMutex.Unlock()
	fake.AddMetricsExtraHandlerStub = nil
	fake.addMetricsExtraHandlerReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeManager) AddMetricsExtraHandlerReturnsOnCall(i int, result1 error) {
	fake.addMetricsExtraHandlerMutex.Lock()
	defer fake.addMetricsExtraHandlerMutex.Unlock()
	fake.AddMetricsExtraHandlerStub = nil
	if fake.addMetricsExtraHandlerReturnsOnCall == nil {
		fake.addMetricsExtraHandlerReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.addMetricsExtraHandlerReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeManager) AddReadyzCheck(arg1 string, arg2 healthz.Checker) error {
	fake.addReadyzCheckMutex.Lock()
	ret, specificReturn := fake.addReadyzCheckReturnsOnCall[len(fake.addReadyzCheckArgsForCall)]
	fake.addReadyzCheckArgsForCall = append(fake.addReadyzCheckArgsForCall, struct {
		arg1 string
		arg2 healthz.Checker
	}{arg1, arg2})
	fake.recordInvocation("AddReadyzCheck", []interface{}{arg1, arg2})
	fake.addReadyzCheckMutex.Unlock()
	if fake.AddReadyzCheckStub != nil {
		return fake.AddReadyzCheckStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.addReadyzCheckReturns
	return fakeReturns.result1
}

func (fake *FakeManager) AddReadyzCheckCallCount() int {
	fake.addReadyzCheckMutex.RLock()
	defer fake.addReadyzCheckMutex.RUnlock()
	return len(fake.addReadyzCheckArgsForCall)
}

func (fake *FakeManager) AddReadyzCheckCalls(stub func(string, healthz.Checker) error) {
	fake.addReadyzCheckMutex.Lock()
	defer fake.addReadyzCheckMutex.Unlock()
	fake.AddReadyzCheckStub = stub
}

func (fake *FakeManager) AddReadyzCheckArgsForCall(i int) (string, healthz.Checker) {
	fake.addReadyzCheckMutex.RLock()
	defer fake.addReadyzCheckMutex.RUnlock()
	argsForCall := fake.addReadyzCheckArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeManager) AddReadyzCheckReturns(result1 error) {
	fake.addReadyzCheckMutex.Lock()
	defer fake.addReadyzCheckMutex.Unlock()
	fake.AddReadyzCheckStub = nil
	fake.addReadyzCheckReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeManager) AddReadyzCheckReturnsOnCall(i int, result1 error) {
	fake.addReadyzCheckMutex.Lock()
	defer fake.addReadyzCheckMutex.Unlock()
	fake.AddReadyzCheckStub = nil
	if fake.addReadyzCheckReturnsOnCall == nil {
		fake.addReadyzCheckReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.addReadyzCheckReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeManager) Elected() <-chan struct{} {
	fake.electedMutex.Lock()
	ret, specificReturn := fake.electedReturnsOnCall[len(fake.electedArgsForCall)]
	fake.electedArgsForCall = append(fake.electedArgsForCall, struct {
	}{})
	fake.recordInvocation("Elected", []interface{}{})
	fake.electedMutex.Unlock()
	if fake.ElectedStub != nil {
		return fake.ElectedStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.electedReturns
	return fakeReturns.result1
}

func (fake *FakeManager) ElectedCallCount() int {
	fake.electedMutex.RLock()
	defer fake.electedMutex.RUnlock()
	return len(fake.electedArgsForCall)
}

func (fake *FakeManager) ElectedCalls(stub func() <-chan struct{}) {
	fake.electedMutex.Lock()
	defer fake.electedMutex.Unlock()
	fake.ElectedStub = stub
}

func (fake *FakeManager) ElectedReturns(result1 <-chan struct{}) {
	fake.electedMutex.Lock()
	defer fake.electedMutex.Unlock()
	fake.ElectedStub = nil
	fake.electedReturns = struct {
		result1 <-chan struct{}
	}{result1}
}

func (fake *FakeManager) ElectedReturnsOnCall(i int, result1 <-chan struct{}) {
	fake.electedMutex.Lock()
	defer fake.electedMutex.Unlock()
	fake.ElectedStub = nil
	if fake.electedReturnsOnCall == nil {
		fake.electedReturnsOnCall = make(map[int]struct {
			result1 <-chan struct{}
		})
	}
	fake.electedReturnsOnCall[i] = struct {
		result1 <-chan struct{}
	}{result1}
}

func (fake *FakeManager) GetAPIReader() client.Reader {
	fake.getAPIReaderMutex.Lock()
	ret, specificReturn := fake.getAPIReaderReturnsOnCall[len(fake.getAPIReaderArgsForCall)]
	fake.getAPIReaderArgsForCall = append(fake.getAPIReaderArgsForCall, struct {
	}{})
	fake.recordInvocation("GetAPIReader", []interface{}{})
	fake.getAPIReaderMutex.Unlock()
	if fake.GetAPIReaderStub != nil {
		return fake.GetAPIReaderStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getAPIReaderReturns
	return fakeReturns.result1
}

func (fake *FakeManager) GetAPIReaderCallCount() int {
	fake.getAPIReaderMutex.RLock()
	defer fake.getAPIReaderMutex.RUnlock()
	return len(fake.getAPIReaderArgsForCall)
}

func (fake *FakeManager) GetAPIReaderCalls(stub func() client.Reader) {
	fake.getAPIReaderMutex.Lock()
	defer fake.getAPIReaderMutex.Unlock()
	fake.GetAPIReaderStub = stub
}

func (fake *FakeManager) GetAPIReaderReturns(result1 client.Reader) {
	fake.getAPIReaderMutex.Lock()
	defer fake.getAPIReaderMutex.Unlock()
	fake.GetAPIReaderStub = nil
	fake.getAPIReaderReturns = struct {
		result1 client.Reader
	}{result1}
}

func (fake *FakeManager) GetAPIReaderReturnsOnCall(i int, result1 client.Reader) {
	fake.getAPIReaderMutex.Lock()
	defer fake.getAPIReaderMutex.Unlock()
	fake.GetAPIReaderStub = nil
	if fake.getAPIReaderReturnsOnCall == nil {
		fake.getAPIReaderReturnsOnCall = make(map[int]struct {
			result1 client.Reader
		})
	}
	fake.getAPIReaderReturnsOnCall[i] = struct {
		result1 client.Reader
	}{result1}
}

func (fake *FakeManager) GetCache() cache.Cache {
	fake.getCacheMutex.Lock()
	ret, specificReturn := fake.getCacheReturnsOnCall[len(fake.getCacheArgsForCall)]
	fake.getCacheArgsForCall = append(fake.getCacheArgsForCall, struct {
	}{})
	fake.recordInvocation("GetCache", []interface{}{})
	fake.getCacheMutex.Unlock()
	if fake.GetCacheStub != nil {
		return fake.GetCacheStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getCacheReturns
	return fakeReturns.result1
}

func (fake *FakeManager) GetCacheCallCount() int {
	fake.getCacheMutex.RLock()
	defer fake.getCacheMutex.RUnlock()
	return len(fake.getCacheArgsForCall)
}

func (fake *FakeManager) GetCacheCalls(stub func() cache.Cache) {
	fake.getCacheMutex.Lock()
	defer fake.getCacheMutex.Unlock()
	fake.GetCacheStub = stub
}

func (fake *FakeManager) GetCacheReturns(result1 cache.Cache) {
	fake.getCacheMutex.Lock()
	defer fake.getCacheMutex.Unlock()
	fake.GetCacheStub = nil
	fake.getCacheReturns = struct {
		result1 cache.Cache
	}{result1}
}

func (fake *FakeManager) GetCacheReturnsOnCall(i int, result1 cache.Cache) {
	fake.getCacheMutex.Lock()
	defer fake.getCacheMutex.Unlock()
	fake.GetCacheStub = nil
	if fake.getCacheReturnsOnCall == nil {
		fake.getCacheReturnsOnCall = make(map[int]struct {
			result1 cache.Cache
		})
	}
	fake.getCacheReturnsOnCall[i] = struct {
		result1 cache.Cache
	}{result1}
}

func (fake *FakeManager) GetClient() client.Client {
	fake.getClientMutex.Lock()
	ret, specificReturn := fake.getClientReturnsOnCall[len(fake.getClientArgsForCall)]
	fake.getClientArgsForCall = append(fake.getClientArgsForCall, struct {
	}{})
	fake.recordInvocation("GetClient", []interface{}{})
	fake.getClientMutex.Unlock()
	if fake.GetClientStub != nil {
		return fake.GetClientStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getClientReturns
	return fakeReturns.result1
}

func (fake *FakeManager) GetClientCallCount() int {
	fake.getClientMutex.RLock()
	defer fake.getClientMutex.RUnlock()
	return len(fake.getClientArgsForCall)
}

func (fake *FakeManager) GetClientCalls(stub func() client.Client) {
	fake.getClientMutex.Lock()
	defer fake.getClientMutex.Unlock()
	fake.GetClientStub = stub
}

func (fake *FakeManager) GetClientReturns(result1 client.Client) {
	fake.getClientMutex.Lock()
	defer fake.getClientMutex.Unlock()
	fake.GetClientStub = nil
	fake.getClientReturns = struct {
		result1 client.Client
	}{result1}
}

func (fake *FakeManager) GetClientReturnsOnCall(i int, result1 client.Client) {
	fake.getClientMutex.Lock()
	defer fake.getClientMutex.Unlock()
	fake.GetClientStub = nil
	if fake.getClientReturnsOnCall == nil {
		fake.getClientReturnsOnCall = make(map[int]struct {
			result1 client.Client
		})
	}
	fake.getClientReturnsOnCall[i] = struct {
		result1 client.Client
	}{result1}
}

func (fake *FakeManager) GetConfig() *rest.Config {
	fake.getConfigMutex.Lock()
	ret, specificReturn := fake.getConfigReturnsOnCall[len(fake.getConfigArgsForCall)]
	fake.getConfigArgsForCall = append(fake.getConfigArgsForCall, struct {
	}{})
	fake.recordInvocation("GetConfig", []interface{}{})
	fake.getConfigMutex.Unlock()
	if fake.GetConfigStub != nil {
		return fake.GetConfigStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getConfigReturns
	return fakeReturns.result1
}

func (fake *FakeManager) GetConfigCallCount() int {
	fake.getConfigMutex.RLock()
	defer fake.getConfigMutex.RUnlock()
	return len(fake.getConfigArgsForCall)
}

func (fake *FakeManager) GetConfigCalls(stub func() *rest.Config) {
	fake.getConfigMutex.Lock()
	defer fake.getConfigMutex.Unlock()
	fake.GetConfigStub = stub
}

func (fake *FakeManager) GetConfigReturns(result1 *rest.Config) {
	fake.getConfigMutex.Lock()
	defer fake.getConfigMutex.Unlock()
	fake.GetConfigStub = nil
	fake.getConfigReturns = struct {
		result1 *rest.Config
	}{result1}
}

func (fake *FakeManager) GetConfigReturnsOnCall(i int, result1 *rest.Config) {
	fake.getConfigMutex.Lock()
	defer fake.getConfigMutex.Unlock()
	fake.GetConfigStub = nil
	if fake.getConfigReturnsOnCall == nil {
		fake.getConfigReturnsOnCall = make(map[int]struct {
			result1 *rest.Config
		})
	}
	fake.getConfigReturnsOnCall[i] = struct {
		result1 *rest.Config
	}{result1}
}

func (fake *FakeManager) GetEventRecorderFor(arg1 string) record.EventRecorder {
	fake.getEventRecorderForMutex.Lock()
	ret, specificReturn := fake.getEventRecorderForReturnsOnCall[len(fake.getEventRecorderForArgsForCall)]
	fake.getEventRecorderForArgsForCall = append(fake.getEventRecorderForArgsForCall, struct {
		arg1 string
	}{arg1})
	fake.recordInvocation("GetEventRecorderFor", []interface{}{arg1})
	fake.getEventRecorderForMutex.Unlock()
	if fake.GetEventRecorderForStub != nil {
		return fake.GetEventRecorderForStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getEventRecorderForReturns
	return fakeReturns.result1
}

func (fake *FakeManager) GetEventRecorderForCallCount() int {
	fake.getEventRecorderForMutex.RLock()
	defer fake.getEventRecorderForMutex.RUnlock()
	return len(fake.getEventRecorderForArgsForCall)
}

func (fake *FakeManager) GetEventRecorderForCalls(stub func(string) record.EventRecorder) {
	fake.getEventRecorderForMutex.Lock()
	defer fake.getEventRecorderForMutex.Unlock()
	fake.GetEventRecorderForStub = stub
}

func (fake *FakeManager) GetEventRecorderForArgsForCall(i int) string {
	fake.getEventRecorderForMutex.RLock()
	defer fake.getEventRecorderForMutex.RUnlock()
	argsForCall := fake.getEventRecorderForArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeManager) GetEventRecorderForReturns(result1 record.EventRecorder) {
	fake.getEventRecorderForMutex.Lock()
	defer fake.getEventRecorderForMutex.Unlock()
	fake.GetEventRecorderForStub = nil
	fake.getEventRecorderForReturns = struct {
		result1 record.EventRecorder
	}{result1}
}

func (fake *FakeManager) GetEventRecorderForReturnsOnCall(i int, result1 record.EventRecorder) {
	fake.getEventRecorderForMutex.Lock()
	defer fake.getEventRecorderForMutex.Unlock()
	fake.GetEventRecorderForStub = nil
	if fake.getEventRecorderForReturnsOnCall == nil {
		fake.getEventRecorderForReturnsOnCall = make(map[int]struct {
			result1 record.EventRecorder
		})
	}
	fake.getEventRecorderForReturnsOnCall[i] = struct {
		result1 record.EventRecorder
	}{result1}
}

func (fake *FakeManager) GetFieldIndexer() client.FieldIndexer {
	fake.getFieldIndexerMutex.Lock()
	ret, specificReturn := fake.getFieldIndexerReturnsOnCall[len(fake.getFieldIndexerArgsForCall)]
	fake.getFieldIndexerArgsForCall = append(fake.getFieldIndexerArgsForCall, struct {
	}{})
	fake.recordInvocation("GetFieldIndexer", []interface{}{})
	fake.getFieldIndexerMutex.Unlock()
	if fake.GetFieldIndexerStub != nil {
		return fake.GetFieldIndexerStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getFieldIndexerReturns
	return fakeReturns.result1
}

func (fake *FakeManager) GetFieldIndexerCallCount() int {
	fake.getFieldIndexerMutex.RLock()
	defer fake.getFieldIndexerMutex.RUnlock()
	return len(fake.getFieldIndexerArgsForCall)
}

func (fake *FakeManager) GetFieldIndexerCalls(stub func() client.FieldIndexer) {
	fake.getFieldIndexerMutex.Lock()
	defer fake.getFieldIndexerMutex.Unlock()
	fake.GetFieldIndexerStub = stub
}

func (fake *FakeManager) GetFieldIndexerReturns(result1 client.FieldIndexer) {
	fake.getFieldIndexerMutex.Lock()
	defer fake.getFieldIndexerMutex.Unlock()
	fake.GetFieldIndexerStub = nil
	fake.getFieldIndexerReturns = struct {
		result1 client.FieldIndexer
	}{result1}
}

func (fake *FakeManager) GetFieldIndexerReturnsOnCall(i int, result1 client.FieldIndexer) {
	fake.getFieldIndexerMutex.Lock()
	defer fake.getFieldIndexerMutex.Unlock()
	fake.GetFieldIndexerStub = nil
	if fake.getFieldIndexerReturnsOnCall == nil {
		fake.getFieldIndexerReturnsOnCall = make(map[int]struct {
			result1 client.FieldIndexer
		})
	}
	fake.getFieldIndexerReturnsOnCall[i] = struct {
		result1 client.FieldIndexer
	}{result1}
}

func (fake *FakeManager) GetLogger() logr.Logger {
	fake.getLoggerMutex.Lock()
	ret, specificReturn := fake.getLoggerReturnsOnCall[len(fake.getLoggerArgsForCall)]
	fake.getLoggerArgsForCall = append(fake.getLoggerArgsForCall, struct {
	}{})
	fake.recordInvocation("GetLogger", []interface{}{})
	fake.getLoggerMutex.Unlock()
	if fake.GetLoggerStub != nil {
		return fake.GetLoggerStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getLoggerReturns
	return fakeReturns.result1
}

func (fake *FakeManager) GetLoggerCallCount() int {
	fake.getLoggerMutex.RLock()
	defer fake.getLoggerMutex.RUnlock()
	return len(fake.getLoggerArgsForCall)
}

func (fake *FakeManager) GetLoggerCalls(stub func() logr.Logger) {
	fake.getLoggerMutex.Lock()
	defer fake.getLoggerMutex.Unlock()
	fake.GetLoggerStub = stub
}

func (fake *FakeManager) GetLoggerReturns(result1 logr.Logger) {
	fake.getLoggerMutex.Lock()
	defer fake.getLoggerMutex.Unlock()
	fake.GetLoggerStub = nil
	fake.getLoggerReturns = struct {
		result1 logr.Logger
	}{result1}
}

func (fake *FakeManager) GetLoggerReturnsOnCall(i int, result1 logr.Logger) {
	fake.getLoggerMutex.Lock()
	defer fake.getLoggerMutex.Unlock()
	fake.GetLoggerStub = nil
	if fake.getLoggerReturnsOnCall == nil {
		fake.getLoggerReturnsOnCall = make(map[int]struct {
			result1 logr.Logger
		})
	}
	fake.getLoggerReturnsOnCall[i] = struct {
		result1 logr.Logger
	}{result1}
}

func (fake *FakeManager) GetRESTMapper() meta.RESTMapper {
	fake.getRESTMapperMutex.Lock()
	ret, specificReturn := fake.getRESTMapperReturnsOnCall[len(fake.getRESTMapperArgsForCall)]
	fake.getRESTMapperArgsForCall = append(fake.getRESTMapperArgsForCall, struct {
	}{})
	fake.recordInvocation("GetRESTMapper", []interface{}{})
	fake.getRESTMapperMutex.Unlock()
	if fake.GetRESTMapperStub != nil {
		return fake.GetRESTMapperStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getRESTMapperReturns
	return fakeReturns.result1
}

func (fake *FakeManager) GetRESTMapperCallCount() int {
	fake.getRESTMapperMutex.RLock()
	defer fake.getRESTMapperMutex.RUnlock()
	return len(fake.getRESTMapperArgsForCall)
}

func (fake *FakeManager) GetRESTMapperCalls(stub func() meta.RESTMapper) {
	fake.getRESTMapperMutex.Lock()
	defer fake.getRESTMapperMutex.Unlock()
	fake.GetRESTMapperStub = stub
}

func (fake *FakeManager) GetRESTMapperReturns(result1 meta.RESTMapper) {
	fake.getRESTMapperMutex.Lock()
	defer fake.getRESTMapperMutex.Unlock()
	fake.GetRESTMapperStub = nil
	fake.getRESTMapperReturns = struct {
		result1 meta.RESTMapper
	}{result1}
}

func (fake *FakeManager) GetRESTMapperReturnsOnCall(i int, result1 meta.RESTMapper) {
	fake.getRESTMapperMutex.Lock()
	defer fake.getRESTMapperMutex.Unlock()
	fake.GetRESTMapperStub = nil
	if fake.getRESTMapperReturnsOnCall == nil {
		fake.getRESTMapperReturnsOnCall = make(map[int]struct {
			result1 meta.RESTMapper
		})
	}
	fake.getRESTMapperReturnsOnCall[i] = struct {
		result1 meta.RESTMapper
	}{result1}
}

func (fake *FakeManager) GetScheme() *runtime.Scheme {
	fake.getSchemeMutex.Lock()
	ret, specificReturn := fake.getSchemeReturnsOnCall[len(fake.getSchemeArgsForCall)]
	fake.getSchemeArgsForCall = append(fake.getSchemeArgsForCall, struct {
	}{})
	fake.recordInvocation("GetScheme", []interface{}{})
	fake.getSchemeMutex.Unlock()
	if fake.GetSchemeStub != nil {
		return fake.GetSchemeStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getSchemeReturns
	return fakeReturns.result1
}

func (fake *FakeManager) GetSchemeCallCount() int {
	fake.getSchemeMutex.RLock()
	defer fake.getSchemeMutex.RUnlock()
	return len(fake.getSchemeArgsForCall)
}

func (fake *FakeManager) GetSchemeCalls(stub func() *runtime.Scheme) {
	fake.getSchemeMutex.Lock()
	defer fake.getSchemeMutex.Unlock()
	fake.GetSchemeStub = stub
}

func (fake *FakeManager) GetSchemeReturns(result1 *runtime.Scheme) {
	fake.getSchemeMutex.Lock()
	defer fake.getSchemeMutex.Unlock()
	fake.GetSchemeStub = nil
	fake.getSchemeReturns = struct {
		result1 *runtime.Scheme
	}{result1}
}

func (fake *FakeManager) GetSchemeReturnsOnCall(i int, result1 *runtime.Scheme) {
	fake.getSchemeMutex.Lock()
	defer fake.getSchemeMutex.Unlock()
	fake.GetSchemeStub = nil
	if fake.getSchemeReturnsOnCall == nil {
		fake.getSchemeReturnsOnCall = make(map[int]struct {
			result1 *runtime.Scheme
		})
	}
	fake.getSchemeReturnsOnCall[i] = struct {
		result1 *runtime.Scheme
	}{result1}
}

func (fake *FakeManager) GetWebhookServer() *webhook.Server {
	fake.getWebhookServerMutex.Lock()
	ret, specificReturn := fake.getWebhookServerReturnsOnCall[len(fake.getWebhookServerArgsForCall)]
	fake.getWebhookServerArgsForCall = append(fake.getWebhookServerArgsForCall, struct {
	}{})
	fake.recordInvocation("GetWebhookServer", []interface{}{})
	fake.getWebhookServerMutex.Unlock()
	if fake.GetWebhookServerStub != nil {
		return fake.GetWebhookServerStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getWebhookServerReturns
	return fakeReturns.result1
}

func (fake *FakeManager) GetWebhookServerCallCount() int {
	fake.getWebhookServerMutex.RLock()
	defer fake.getWebhookServerMutex.RUnlock()
	return len(fake.getWebhookServerArgsForCall)
}

func (fake *FakeManager) GetWebhookServerCalls(stub func() *webhook.Server) {
	fake.getWebhookServerMutex.Lock()
	defer fake.getWebhookServerMutex.Unlock()
	fake.GetWebhookServerStub = stub
}

func (fake *FakeManager) GetWebhookServerReturns(result1 *webhook.Server) {
	fake.getWebhookServerMutex.Lock()
	defer fake.getWebhookServerMutex.Unlock()
	fake.GetWebhookServerStub = nil
	fake.getWebhookServerReturns = struct {
		result1 *webhook.Server
	}{result1}
}

func (fake *FakeManager) GetWebhookServerReturnsOnCall(i int, result1 *webhook.Server) {
	fake.getWebhookServerMutex.Lock()
	defer fake.getWebhookServerMutex.Unlock()
	fake.GetWebhookServerStub = nil
	if fake.getWebhookServerReturnsOnCall == nil {
		fake.getWebhookServerReturnsOnCall = make(map[int]struct {
			result1 *webhook.Server
		})
	}
	fake.getWebhookServerReturnsOnCall[i] = struct {
		result1 *webhook.Server
	}{result1}
}

func (fake *FakeManager) SetFields(arg1 interface{}) error {
	fake.setFieldsMutex.Lock()
	ret, specificReturn := fake.setFieldsReturnsOnCall[len(fake.setFieldsArgsForCall)]
	fake.setFieldsArgsForCall = append(fake.setFieldsArgsForCall, struct {
		arg1 interface{}
	}{arg1})
	fake.recordInvocation("SetFields", []interface{}{arg1})
	fake.setFieldsMutex.Unlock()
	if fake.SetFieldsStub != nil {
		return fake.SetFieldsStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.setFieldsReturns
	return fakeReturns.result1
}

func (fake *FakeManager) SetFieldsCallCount() int {
	fake.setFieldsMutex.RLock()
	defer fake.setFieldsMutex.RUnlock()
	return len(fake.setFieldsArgsForCall)
}

func (fake *FakeManager) SetFieldsCalls(stub func(interface{}) error) {
	fake.setFieldsMutex.Lock()
	defer fake.setFieldsMutex.Unlock()
	fake.SetFieldsStub = stub
}

func (fake *FakeManager) SetFieldsArgsForCall(i int) interface{} {
	fake.setFieldsMutex.RLock()
	defer fake.setFieldsMutex.RUnlock()
	argsForCall := fake.setFieldsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeManager) SetFieldsReturns(result1 error) {
	fake.setFieldsMutex.Lock()
	defer fake.setFieldsMutex.Unlock()
	fake.SetFieldsStub = nil
	fake.setFieldsReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeManager) SetFieldsReturnsOnCall(i int, result1 error) {
	fake.setFieldsMutex.Lock()
	defer fake.setFieldsMutex.Unlock()
	fake.SetFieldsStub = nil
	if fake.setFieldsReturnsOnCall == nil {
		fake.setFieldsReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.setFieldsReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeManager) Start(arg1 context.Context) error {
	fake.startMutex.Lock()
	ret, specificReturn := fake.startReturnsOnCall[len(fake.startArgsForCall)]
	fake.startArgsForCall = append(fake.startArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	fake.recordInvocation("Start", []interface{}{arg1})
	fake.startMutex.Unlock()
	if fake.StartStub != nil {
		return fake.StartStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.startReturns
	return fakeReturns.result1
}

func (fake *FakeManager) StartCallCount() int {
	fake.startMutex.RLock()
	defer fake.startMutex.RUnlock()
	return len(fake.startArgsForCall)
}

func (fake *FakeManager) StartCalls(stub func(context.Context) error) {
	fake.startMutex.Lock()
	defer fake.startMutex.Unlock()
	fake.StartStub = stub
}

func (fake *FakeManager) StartArgsForCall(i int) context.Context {
	fake.startMutex.RLock()
	defer fake.startMutex.RUnlock()
	argsForCall := fake.startArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeManager) StartReturns(result1 error) {
	fake.startMutex.Lock()
	defer fake.startMutex.Unlock()
	fake.StartStub = nil
	fake.startReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeManager) StartReturnsOnCall(i int, result1 error) {
	fake.startMutex.Lock()
	defer fake.startMutex.Unlock()
	fake.StartStub = nil
	if fake.startReturnsOnCall == nil {
		fake.startReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.startReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeManager) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.addMutex.RLock()
	defer fake.addMutex.RUnlock()
	fake.addHealthzCheckMutex.RLock()
	defer fake.addHealthzCheckMutex.RUnlock()
	fake.addMetricsExtraHandlerMutex.RLock()
	defer fake.addMetricsExtraHandlerMutex.RUnlock()
	fake.addReadyzCheckMutex.RLock()
	defer fake.addReadyzCheckMutex.RUnlock()
	fake.electedMutex.RLock()
	defer fake.electedMutex.RUnlock()
	fake.getAPIReaderMutex.RLock()
	defer fake.getAPIReaderMutex.RUnlock()
	fake.getCacheMutex.RLock()
	defer fake.getCacheMutex.RUnlock()
	fake.getClientMutex.RLock()
	defer fake.getClientMutex.RUnlock()
	fake.getConfigMutex.RLock()
	defer fake.getConfigMutex.RUnlock()
	fake.getEventRecorderForMutex.RLock()
	defer fake.getEventRecorderForMutex.RUnlock()
	fake.getFieldIndexerMutex.RLock()
	defer fake.getFieldIndexerMutex.RUnlock()
	fake.getLoggerMutex.RLock()
	defer fake.getLoggerMutex.RUnlock()
	fake.getRESTMapperMutex.RLock()
	defer fake.getRESTMapperMutex.RUnlock()
	fake.getSchemeMutex.RLock()
	defer fake.getSchemeMutex.RUnlock()
	fake.getWebhookServerMutex.RLock()
	defer fake.getWebhookServerMutex.RUnlock()
	fake.setFieldsMutex.RLock()
	defer fake.setFieldsMutex.RUnlock()
	fake.startMutex.RLock()
	defer fake.startMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeManager) recordInvocation(key string, args []interface{}) {
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

var _ manager.Manager = new(FakeManager)
