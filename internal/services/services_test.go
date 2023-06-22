package services

import (
	"testing"

	"github.com/javaman/go-metrics/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) SaveGauge(name string, v float64) {
	m.Called(name, v)
}

func (m *mockStorage) GetGauge(name string) (float64, bool) {
	args := m.Called(name)
	return args.Get(0).(float64), args.Bool(1)
}

func (m *mockStorage) AllGauges(f func(string, float64)) {
	m.Called(f)
}

func (m *mockStorage) SaveCounter(name string, v int64) {
	m.Called(name, v)
}

func (m *mockStorage) GetCounter(name string) (int64, bool) {
	args := m.Called(name)
	return args.Get(0).(int64), args.Bool(1)
}

func (m *mockStorage) AllCounters(f func(string, int64)) {
	m.Called(f)
}

func (m *mockStorage) WriteToFile(fname string) {
	m.Called(fname)
}

func (m *mockStorage) Lock() repository.LockedStorage {
	m.Called()
	return nil
}

func TestSaveGauge(t *testing.T) {
	theMock := &mockStorage{}
	theMock.On("SaveGauge", "one", 3.14)
	ms := NewMetricsService(theMock, func() error { return nil })
	ms.SaveGauge("one", 3.14)
	theMock.AssertCalled(t, "SaveGauge", "one", 3.14)
	theMock.AssertExpectations(t)
	mock.AssertExpectationsForObjects(t, theMock)
}

func TestGetGauge(t *testing.T) {
	theMock := &mockStorage{}
	theMock.On("GetGauge", "one").Return(3.14, true)
	ms := NewMetricsService(theMock, func() error { return nil })
	if v, ok := ms.GetGauge("one"); ok {
		assert.Equal(t, v, 3.14, "That should not happen")
	} else {
		assert.Fail(t, "Ooops")
	}
	theMock.AssertCalled(t, "GetGauge", "one")
	theMock.AssertExpectations(t)
	mock.AssertExpectationsForObjects(t, theMock)
}

func TestAllGauges(t *testing.T) {
	theMock := &mockStorage{}
	f := func(k string, v float64) {}
	theMock.On("AllGauges", mock.Anything)
	ms := NewMetricsService(theMock, func() error { return nil })
	ms.AllGauges(f)
	theMock.AssertCalled(t, "AllGauges", mock.Anything)
	theMock.AssertExpectations(t)
	mock.AssertExpectationsForObjects(t, theMock)
}

func TestSaveCounter(t *testing.T) {
	theMock := &mockStorage{}
	theMock.On("GetCounter", "one").Return(int64(0), false)
	theMock.On("SaveCounter", "one", int64(1))
	ms := NewMetricsService(theMock, func() error { return nil })
	ms.SaveCounter("one", 1)
	theMock.AssertCalled(t, "GetCounter", "one")
	theMock.AssertCalled(t, "SaveCounter", "one", int64(1))
	theMock.AssertExpectations(t)
	mock.AssertExpectationsForObjects(t, theMock)
}

func TestSaveCounterUpdate(t *testing.T) {
	theMock := &mockStorage{}
	theMock.On("GetCounter", "one").Return(int64(3), true)
	theMock.On("SaveCounter", "one", int64(4))
	ms := NewMetricsService(theMock, func() error { return nil })
	ms.SaveCounter("one", 1)
	theMock.AssertCalled(t, "GetCounter", "one")
	theMock.AssertCalled(t, "SaveCounter", "one", int64(4))
	theMock.AssertExpectations(t)
	mock.AssertExpectationsForObjects(t, theMock)
}

func TestGetCounter(t *testing.T) {
	theMock := &mockStorage{}
	theMock.On("GetCounter", "one").Return(int64(42), true)
	ms := NewMetricsService(theMock, func() error { return nil })
	if v, ok := ms.GetCounter("one"); ok {
		assert.Equal(t, v, int64(42), "That should not happen")
	} else {
		assert.Fail(t, "Ooops")
	}
	theMock.AssertCalled(t, "GetCounter", "one")
	theMock.AssertExpectations(t)
	mock.AssertExpectationsForObjects(t, theMock)
}

func TestAllCounters(t *testing.T) {
	theMock := &mockStorage{}
	f := func(k string, v int64) {}
	theMock.On("AllCounters", mock.Anything)
	ms := NewMetricsService(theMock, func() error { return nil })
	ms.AllCounters(f)
	theMock.AssertCalled(t, "AllCounters", mock.Anything)
	theMock.AssertExpectations(t)
	mock.AssertExpectationsForObjects(t, theMock)
}
