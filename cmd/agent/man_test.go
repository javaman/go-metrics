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
