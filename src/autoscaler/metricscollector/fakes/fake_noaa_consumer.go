// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"autoscaler/metricscollector/noaa"
	"sync"

	"github.com/cloudfoundry/sonde-go/events"
)

type FakeNoaaConsumer struct {
	ContainerEnvelopesStub        func(appGuid string, authToken string) ([]*events.Envelope, error)
	containerEnvelopesMutex       sync.RWMutex
	containerEnvelopesArgsForCall []struct {
		appGuid   string
		authToken string
	}
	containerEnvelopesReturns struct {
		result1 []*events.Envelope
		result2 error
	}
	containerEnvelopesReturnsOnCall map[int]struct {
		result1 []*events.Envelope
		result2 error
	}
	StreamStub        func(appGuid string, authToken string) (outputChan <-chan *events.Envelope, errorChan <-chan error)
	streamMutex       sync.RWMutex
	streamArgsForCall []struct {
		appGuid   string
		authToken string
	}
	streamReturns struct {
		result1 <-chan *events.Envelope
		result2 <-chan error
	}
	streamReturnsOnCall map[int]struct {
		result1 <-chan *events.Envelope
		result2 <-chan error
	}
	FirehoseStub        func(subscriptionId string, authToken string) (outputChan <-chan *events.Envelope, errorChan <-chan error)
	firehoseMutex       sync.RWMutex
	firehoseArgsForCall []struct {
		subscriptionId string
		authToken      string
	}
	firehoseReturns struct {
		result1 <-chan *events.Envelope
		result2 <-chan error
	}
	firehoseReturnsOnCall map[int]struct {
		result1 <-chan *events.Envelope
		result2 <-chan error
	}
	CloseStub        func() error
	closeMutex       sync.RWMutex
	closeArgsForCall []struct{}
	closeReturns     struct {
		result1 error
	}
	closeReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeNoaaConsumer) ContainerEnvelopes(appGuid string, authToken string) ([]*events.Envelope, error) {
	fake.containerEnvelopesMutex.Lock()
	ret, specificReturn := fake.containerEnvelopesReturnsOnCall[len(fake.containerEnvelopesArgsForCall)]
	fake.containerEnvelopesArgsForCall = append(fake.containerEnvelopesArgsForCall, struct {
		appGuid   string
		authToken string
	}{appGuid, authToken})
	fake.recordInvocation("ContainerEnvelopes", []interface{}{appGuid, authToken})
	fake.containerEnvelopesMutex.Unlock()
	if fake.ContainerEnvelopesStub != nil {
		return fake.ContainerEnvelopesStub(appGuid, authToken)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.containerEnvelopesReturns.result1, fake.containerEnvelopesReturns.result2
}

func (fake *FakeNoaaConsumer) ContainerEnvelopesCallCount() int {
	fake.containerEnvelopesMutex.RLock()
	defer fake.containerEnvelopesMutex.RUnlock()
	return len(fake.containerEnvelopesArgsForCall)
}

func (fake *FakeNoaaConsumer) ContainerEnvelopesArgsForCall(i int) (string, string) {
	fake.containerEnvelopesMutex.RLock()
	defer fake.containerEnvelopesMutex.RUnlock()
	return fake.containerEnvelopesArgsForCall[i].appGuid, fake.containerEnvelopesArgsForCall[i].authToken
}

func (fake *FakeNoaaConsumer) ContainerEnvelopesReturns(result1 []*events.Envelope, result2 error) {
	fake.ContainerEnvelopesStub = nil
	fake.containerEnvelopesReturns = struct {
		result1 []*events.Envelope
		result2 error
	}{result1, result2}
}

func (fake *FakeNoaaConsumer) ContainerEnvelopesReturnsOnCall(i int, result1 []*events.Envelope, result2 error) {
	fake.ContainerEnvelopesStub = nil
	if fake.containerEnvelopesReturnsOnCall == nil {
		fake.containerEnvelopesReturnsOnCall = make(map[int]struct {
			result1 []*events.Envelope
			result2 error
		})
	}
	fake.containerEnvelopesReturnsOnCall[i] = struct {
		result1 []*events.Envelope
		result2 error
	}{result1, result2}
}

func (fake *FakeNoaaConsumer) Stream(appGuid string, authToken string) (outputChan <-chan *events.Envelope, errorChan <-chan error) {
	fake.streamMutex.Lock()
	ret, specificReturn := fake.streamReturnsOnCall[len(fake.streamArgsForCall)]
	fake.streamArgsForCall = append(fake.streamArgsForCall, struct {
		appGuid   string
		authToken string
	}{appGuid, authToken})
	fake.recordInvocation("Stream", []interface{}{appGuid, authToken})
	fake.streamMutex.Unlock()
	if fake.StreamStub != nil {
		return fake.StreamStub(appGuid, authToken)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.streamReturns.result1, fake.streamReturns.result2
}

func (fake *FakeNoaaConsumer) StreamCallCount() int {
	fake.streamMutex.RLock()
	defer fake.streamMutex.RUnlock()
	return len(fake.streamArgsForCall)
}

func (fake *FakeNoaaConsumer) StreamArgsForCall(i int) (string, string) {
	fake.streamMutex.RLock()
	defer fake.streamMutex.RUnlock()
	return fake.streamArgsForCall[i].appGuid, fake.streamArgsForCall[i].authToken
}

func (fake *FakeNoaaConsumer) StreamReturns(result1 <-chan *events.Envelope, result2 <-chan error) {
	fake.StreamStub = nil
	fake.streamReturns = struct {
		result1 <-chan *events.Envelope
		result2 <-chan error
	}{result1, result2}
}

func (fake *FakeNoaaConsumer) StreamReturnsOnCall(i int, result1 <-chan *events.Envelope, result2 <-chan error) {
	fake.StreamStub = nil
	if fake.streamReturnsOnCall == nil {
		fake.streamReturnsOnCall = make(map[int]struct {
			result1 <-chan *events.Envelope
			result2 <-chan error
		})
	}
	fake.streamReturnsOnCall[i] = struct {
		result1 <-chan *events.Envelope
		result2 <-chan error
	}{result1, result2}
}

func (fake *FakeNoaaConsumer) Firehose(subscriptionId string, authToken string) (outputChan <-chan *events.Envelope, errorChan <-chan error) {
	fake.firehoseMutex.Lock()
	ret, specificReturn := fake.firehoseReturnsOnCall[len(fake.firehoseArgsForCall)]
	fake.firehoseArgsForCall = append(fake.firehoseArgsForCall, struct {
		subscriptionId string
		authToken      string
	}{subscriptionId, authToken})
	fake.recordInvocation("Firehose", []interface{}{subscriptionId, authToken})
	fake.firehoseMutex.Unlock()
	if fake.FirehoseStub != nil {
		return fake.FirehoseStub(subscriptionId, authToken)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.firehoseReturns.result1, fake.firehoseReturns.result2
}

func (fake *FakeNoaaConsumer) FirehoseCallCount() int {
	fake.firehoseMutex.RLock()
	defer fake.firehoseMutex.RUnlock()
	return len(fake.firehoseArgsForCall)
}

func (fake *FakeNoaaConsumer) FirehoseArgsForCall(i int) (string, string) {
	fake.firehoseMutex.RLock()
	defer fake.firehoseMutex.RUnlock()
	return fake.firehoseArgsForCall[i].subscriptionId, fake.firehoseArgsForCall[i].authToken
}

func (fake *FakeNoaaConsumer) FirehoseReturns(result1 <-chan *events.Envelope, result2 <-chan error) {
	fake.FirehoseStub = nil
	fake.firehoseReturns = struct {
		result1 <-chan *events.Envelope
		result2 <-chan error
	}{result1, result2}
}

func (fake *FakeNoaaConsumer) FirehoseReturnsOnCall(i int, result1 <-chan *events.Envelope, result2 <-chan error) {
	fake.FirehoseStub = nil
	if fake.firehoseReturnsOnCall == nil {
		fake.firehoseReturnsOnCall = make(map[int]struct {
			result1 <-chan *events.Envelope
			result2 <-chan error
		})
	}
	fake.firehoseReturnsOnCall[i] = struct {
		result1 <-chan *events.Envelope
		result2 <-chan error
	}{result1, result2}
}

func (fake *FakeNoaaConsumer) Close() error {
	fake.closeMutex.Lock()
	ret, specificReturn := fake.closeReturnsOnCall[len(fake.closeArgsForCall)]
	fake.closeArgsForCall = append(fake.closeArgsForCall, struct{}{})
	fake.recordInvocation("Close", []interface{}{})
	fake.closeMutex.Unlock()
	if fake.CloseStub != nil {
		return fake.CloseStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.closeReturns.result1
}

func (fake *FakeNoaaConsumer) CloseCallCount() int {
	fake.closeMutex.RLock()
	defer fake.closeMutex.RUnlock()
	return len(fake.closeArgsForCall)
}

func (fake *FakeNoaaConsumer) CloseReturns(result1 error) {
	fake.CloseStub = nil
	fake.closeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeNoaaConsumer) CloseReturnsOnCall(i int, result1 error) {
	fake.CloseStub = nil
	if fake.closeReturnsOnCall == nil {
		fake.closeReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.closeReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeNoaaConsumer) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.containerEnvelopesMutex.RLock()
	defer fake.containerEnvelopesMutex.RUnlock()
	fake.streamMutex.RLock()
	defer fake.streamMutex.RUnlock()
	fake.firehoseMutex.RLock()
	defer fake.firehoseMutex.RUnlock()
	fake.closeMutex.RLock()
	defer fake.closeMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeNoaaConsumer) recordInvocation(key string, args []interface{}) {
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

var _ noaa.NoaaConsumer = new(FakeNoaaConsumer)
