// This file was generated by counterfeiter
package fakes

import (
	"sync"

	"github.com/svett/pho"
)

type FakeSocketWriter struct {
	SocketIDStub        func() string
	socketIDMutex       sync.RWMutex
	socketIDArgsForCall []struct{}
	socketIDReturns     struct {
		result1 string
	}
	UserAgentStub        func() string
	userAgentMutex       sync.RWMutex
	userAgentArgsForCall []struct{}
	userAgentReturns     struct {
		result1 string
	}
	RemoteAddrStub        func() string
	remoteAddrMutex       sync.RWMutex
	remoteAddrArgsForCall []struct{}
	remoteAddrReturns     struct {
		result1 string
	}
	MetadataStub        func() pho.Metadata
	metadataMutex       sync.RWMutex
	metadataArgsForCall []struct{}
	metadataReturns     struct {
		result1 pho.Metadata
	}
	WriteStub        func(string, int, []byte) error
	writeMutex       sync.RWMutex
	writeArgsForCall []struct {
		arg1 string
		arg2 int
		arg3 []byte
	}
	writeReturns struct {
		result1 error
	}
	WriteErrorStub        func(err error, code int) error
	writeErrorMutex       sync.RWMutex
	writeErrorArgsForCall []struct {
		err  error
		code int
	}
	writeErrorReturns struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeSocketWriter) SocketID() string {
	fake.socketIDMutex.Lock()
	fake.socketIDArgsForCall = append(fake.socketIDArgsForCall, struct{}{})
	fake.recordInvocation("SocketID", []interface{}{})
	fake.socketIDMutex.Unlock()
	if fake.SocketIDStub != nil {
		return fake.SocketIDStub()
	}
	return fake.socketIDReturns.result1
}

func (fake *FakeSocketWriter) SocketIDCallCount() int {
	fake.socketIDMutex.RLock()
	defer fake.socketIDMutex.RUnlock()
	return len(fake.socketIDArgsForCall)
}

func (fake *FakeSocketWriter) SocketIDReturns(result1 string) {
	fake.SocketIDStub = nil
	fake.socketIDReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeSocketWriter) UserAgent() string {
	fake.userAgentMutex.Lock()
	fake.userAgentArgsForCall = append(fake.userAgentArgsForCall, struct{}{})
	fake.recordInvocation("UserAgent", []interface{}{})
	fake.userAgentMutex.Unlock()
	if fake.UserAgentStub != nil {
		return fake.UserAgentStub()
	}
	return fake.userAgentReturns.result1
}

func (fake *FakeSocketWriter) UserAgentCallCount() int {
	fake.userAgentMutex.RLock()
	defer fake.userAgentMutex.RUnlock()
	return len(fake.userAgentArgsForCall)
}

func (fake *FakeSocketWriter) UserAgentReturns(result1 string) {
	fake.UserAgentStub = nil
	fake.userAgentReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeSocketWriter) RemoteAddr() string {
	fake.remoteAddrMutex.Lock()
	fake.remoteAddrArgsForCall = append(fake.remoteAddrArgsForCall, struct{}{})
	fake.recordInvocation("RemoteAddr", []interface{}{})
	fake.remoteAddrMutex.Unlock()
	if fake.RemoteAddrStub != nil {
		return fake.RemoteAddrStub()
	}
	return fake.remoteAddrReturns.result1
}

func (fake *FakeSocketWriter) RemoteAddrCallCount() int {
	fake.remoteAddrMutex.RLock()
	defer fake.remoteAddrMutex.RUnlock()
	return len(fake.remoteAddrArgsForCall)
}

func (fake *FakeSocketWriter) RemoteAddrReturns(result1 string) {
	fake.RemoteAddrStub = nil
	fake.remoteAddrReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeSocketWriter) Metadata() pho.Metadata {
	fake.metadataMutex.Lock()
	fake.metadataArgsForCall = append(fake.metadataArgsForCall, struct{}{})
	fake.recordInvocation("Metadata", []interface{}{})
	fake.metadataMutex.Unlock()
	if fake.MetadataStub != nil {
		return fake.MetadataStub()
	}
	return fake.metadataReturns.result1
}

func (fake *FakeSocketWriter) MetadataCallCount() int {
	fake.metadataMutex.RLock()
	defer fake.metadataMutex.RUnlock()
	return len(fake.metadataArgsForCall)
}

func (fake *FakeSocketWriter) MetadataReturns(result1 pho.Metadata) {
	fake.MetadataStub = nil
	fake.metadataReturns = struct {
		result1 pho.Metadata
	}{result1}
}

func (fake *FakeSocketWriter) Write(arg1 string, arg2 int, arg3 []byte) error {
	var arg3Copy []byte
	if arg3 != nil {
		arg3Copy = make([]byte, len(arg3))
		copy(arg3Copy, arg3)
	}
	fake.writeMutex.Lock()
	fake.writeArgsForCall = append(fake.writeArgsForCall, struct {
		arg1 string
		arg2 int
		arg3 []byte
	}{arg1, arg2, arg3Copy})
	fake.recordInvocation("Write", []interface{}{arg1, arg2, arg3Copy})
	fake.writeMutex.Unlock()
	if fake.WriteStub != nil {
		return fake.WriteStub(arg1, arg2, arg3)
	}
	return fake.writeReturns.result1
}

func (fake *FakeSocketWriter) WriteCallCount() int {
	fake.writeMutex.RLock()
	defer fake.writeMutex.RUnlock()
	return len(fake.writeArgsForCall)
}

func (fake *FakeSocketWriter) WriteArgsForCall(i int) (string, int, []byte) {
	fake.writeMutex.RLock()
	defer fake.writeMutex.RUnlock()
	return fake.writeArgsForCall[i].arg1, fake.writeArgsForCall[i].arg2, fake.writeArgsForCall[i].arg3
}

func (fake *FakeSocketWriter) WriteReturns(result1 error) {
	fake.WriteStub = nil
	fake.writeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeSocketWriter) WriteError(err error, code int) error {
	fake.writeErrorMutex.Lock()
	fake.writeErrorArgsForCall = append(fake.writeErrorArgsForCall, struct {
		err  error
		code int
	}{err, code})
	fake.recordInvocation("WriteError", []interface{}{err, code})
	fake.writeErrorMutex.Unlock()
	if fake.WriteErrorStub != nil {
		return fake.WriteErrorStub(err, code)
	}
	return fake.writeErrorReturns.result1
}

func (fake *FakeSocketWriter) WriteErrorCallCount() int {
	fake.writeErrorMutex.RLock()
	defer fake.writeErrorMutex.RUnlock()
	return len(fake.writeErrorArgsForCall)
}

func (fake *FakeSocketWriter) WriteErrorArgsForCall(i int) (error, int) {
	fake.writeErrorMutex.RLock()
	defer fake.writeErrorMutex.RUnlock()
	return fake.writeErrorArgsForCall[i].err, fake.writeErrorArgsForCall[i].code
}

func (fake *FakeSocketWriter) WriteErrorReturns(result1 error) {
	fake.WriteErrorStub = nil
	fake.writeErrorReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeSocketWriter) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.socketIDMutex.RLock()
	defer fake.socketIDMutex.RUnlock()
	fake.userAgentMutex.RLock()
	defer fake.userAgentMutex.RUnlock()
	fake.remoteAddrMutex.RLock()
	defer fake.remoteAddrMutex.RUnlock()
	fake.metadataMutex.RLock()
	defer fake.metadataMutex.RUnlock()
	fake.writeMutex.RLock()
	defer fake.writeMutex.RUnlock()
	fake.writeErrorMutex.RLock()
	defer fake.writeErrorMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeSocketWriter) recordInvocation(key string, args []interface{}) {
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

var _ pho.SocketWriter = new(FakeSocketWriter)
