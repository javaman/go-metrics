package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStorageAddGauge(t *testing.T) {
	ms := NewInMemoryStorage()

	const gaugeName = "g1"
	const gaugeValue = float64(3.14)

	ms.SaveGauge(gaugeName, gaugeValue)

	assert.Equal(t, gaugeValue, ms.gauges[gaugeName], "Test add gauge Gauge with same name must be equal")
}

func TestMemStorageGetGauge(t *testing.T) {
	ms := NewInMemoryStorage()

	const gaugeName = "g1"
	const gaugeValue = float64(3.14)

	ms.gauges[gaugeName] = gaugeValue

	if value, ok := ms.GetGauge(gaugeName); ok {
		assert.Equal(t, gaugeValue, value, "Test get gauge with same name must be equal")
	} else {
		assert.Fail(t, "Gague Value Must Be There")
	}

}

func TestMemStorageGetGaugeNotExist(t *testing.T) {
	ms := NewInMemoryStorage()

	const gaugeName = "g1"

	if _, ok := ms.GetGauge(gaugeName); ok {
		assert.Fail(t, "Unexpected gauge")
	}

}

func TestMemStorageAllGauges(t *testing.T) {
	ms := NewInMemoryStorage()
	testData := map[string]float64{
		"pi": 3.14,
		"e":  2.72,
	}
	for k, v := range testData {
		ms.SaveGauge(k, v)
	}
	ms.AllGauges(func(k string, v float64) {
		if value, ok := testData[k]; ok {
			assert.Equal(t, v, value, "Must be equals")
			delete(testData, k)
		} else {
			assert.Fail(t, "Unknow gauge!")
		}
	})
	assert.Empty(t, testData, "all gauges must be enumerated")
}

func TestMemStorageAddCounter(t *testing.T) {
	ms := NewInMemoryStorage()

	const counterName = "c1"
	const counterValue = int64(42)

	ms.SaveCounter(counterName, counterValue)

	assert.Equal(t, counterValue, ms.counters[counterName], "Test add counter with same name must be equal")
}

func TestMemStorageGetCounter(t *testing.T) {
	ms := NewInMemoryStorage()

	const counterName = "g1"
	const counterValue = int64(42)

	ms.counters[counterName] = counterValue

	if value, ok := ms.GetCounter(counterName); ok {
		assert.Equal(t, counterValue, value, "Test GetCounter with same name must be equal")
	} else {
		assert.Fail(t, "Counter Value Must Be There")
	}

}

func TestMemStorageGetCounterNotExist(t *testing.T) {
	ms := NewInMemoryStorage()

	const counterName = "c1"

	if _, ok := ms.GetCounter(counterName); ok {
		assert.Fail(t, "Unexpected counter")
	}

}

func TestMemStorageAllCounters(t *testing.T) {
	ms := NewInMemoryStorage()
	testData := map[string]int64{
		"42":   42,
		"2023": 2023,
	}
	for k, v := range testData {
		ms.SaveCounter(k, v)
	}
	ms.AllCounters(func(k string, v int64) {
		if value, ok := testData[k]; ok {
			assert.Equal(t, v, value, "Must be equals")
			delete(testData, k)
		} else {
			assert.Fail(t, "Unknow counter!")
		}
	})
	assert.Empty(t, testData, "all counters must be enumerated")
}
