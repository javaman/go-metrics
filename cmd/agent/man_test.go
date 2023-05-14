package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveGauge(t *testing.T) {
	mb := &measuresBuffer{}

	gm := GaugeMeasure{float64(2.72), "Test"}
	gm.save(mb)

	assert.Equal(t, 1, len(mb.buffer))
	assert.Equal(t, gm, mb.buffer[0])
}

func TestSaveCounter(t *testing.T) {
	mb := &measuresBuffer{}

	cm := CounterMeasure{42, "Test"}
	cm.save(mb)

	assert.Equal(t, 1, len(mb.buffer))
	assert.Equal(t, cm, mb.buffer[0])
}

func TestDefaultMeasured(t *testing.T) {
	expectedMetrics := map[string]bool{
		"Alloc":         true,
		"BuckHashSys":   true,
		"Frees":         true,
		"GCCPUFraction": true,
		"GCSys":         true,
		"HeapAlloc":     true,
		"HeapIdle":      true,
		"HeapInuse":     true,
		"HeapObjects":   true,
		"HeapReleased":  true,
		"HeapSys":       true,
		"LastGC":        true,
		"Lookups":       true,
		"MCacheInuse":   true,
		"MSpanSys":      true,
		"Mallocs":       true,
		"NextGC":        true,
		"OtherSys":      true,
		"PauseTotalNs":  true,
		"StackInuse":    true,
		"StackSys":      true,
		"Sys":           true,
		"TotalAlloc":    true,
		"PollCount":     true,
		"RandomValue":   true,
	}

	mb := &measuresBuffer{}
	dm := defaultMeasured{}

	dm.captureMetrics(mb)

	for _, v := range mb.buffer {
		delete(expectedMetrics, v.name())
	}

	assert.Empty(t, expectedMetrics)
}

func TestDefaultWorker(t *testing.T) {
	callsCount := map[string]int{
		"one": 2,
		"two": 1,
	}
	f1 := func() { callsCount["one"] = callsCount["one"] - 1 }
	f2 := func() { callsCount["two"] = callsCount["two"] - 1 }
	p := func() bool { return callsCount["two"] > 0 }
	dw := &defaultWorker{
		1,
		2,
		f1,
		f2,
		p,
	}
	dw.run()
	assert.Equal(t, 0, callsCount["one"])
	assert.Equal(t, 0, callsCount["two"])

}
