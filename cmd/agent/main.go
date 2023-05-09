package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
)

type MeasureDestination interface {
	saveCounter(m Measure, value int64)
	saveGauge(m Measure, value float64)
}

type Measure interface {
	send(to MeasureDestination)
	name() string
}

type GaugeMeasure struct {
	value float64
	mname string
}

type CounterMeasure struct {
	value int64
	mname string
}

func (m GaugeMeasure) name() string {
	return m.mname
}

func (m GaugeMeasure) send(to MeasureDestination) {
	to.saveGauge(m, m.value)
}

func (m CounterMeasure) name() string {
	return m.mname
}

func (m CounterMeasure) send(to MeasureDestination) {
	to.saveCounter(m, m.value)
}

type Measured interface {
	captureMetrics(to MeasureDestination)
}

type measuresBuffer struct {
	buffer []Measure
}

func (mb *measuresBuffer) save(m Measure) {
	mb.buffer = append(mb.buffer, m)
}

func (mb *measuresBuffer) saveCounter(m Measure, value int64) {
	mb.save(m)
}

func (mb *measuresBuffer) saveGauge(m Measure, value float64) {
	mb.save(m)
}

type defaultMeasured struct {
	pollCount int64
}

func (d *defaultMeasured) captureMetrics(destination MeasureDestination) {
	memStats := new(runtime.MemStats)
	runtime.ReadMemStats(memStats)

	GaugeMeasure{float64(memStats.Alloc), "Alloc"}.send(destination)
	GaugeMeasure{float64(memStats.BuckHashSys), "BuckHashSys"}.send(destination)
	GaugeMeasure{float64(memStats.Frees), "Frees"}.send(destination)
	GaugeMeasure{memStats.GCCPUFraction, "GCCPUFraction"}.send(destination)
	GaugeMeasure{float64(memStats.GCSys), "GCSys"}.send(destination)
	GaugeMeasure{float64(memStats.HeapAlloc), "HeapAlloc"}.send(destination)
	GaugeMeasure{float64(memStats.HeapIdle), "HeapIdle"}.send(destination)
	GaugeMeasure{float64(memStats.HeapInuse), "HeapInuse"}.send(destination)
	GaugeMeasure{float64(memStats.HeapObjects), "HeapObjects"}.send(destination)
	GaugeMeasure{float64(memStats.HeapReleased), "HeapReleased"}.send(destination)
	GaugeMeasure{float64(memStats.HeapSys), "HeapSys"}.send(destination)
	GaugeMeasure{float64(memStats.LastGC), "LastGC"}.send(destination)
	GaugeMeasure{float64(memStats.Lookups), "Lookups"}.send(destination)
	GaugeMeasure{float64(memStats.MCacheInuse), "MCacheInuse"}.send(destination)
	GaugeMeasure{float64(memStats.MSpanSys), "MSpanSys"}.send(destination)
	GaugeMeasure{float64(memStats.Mallocs), "Mallocs"}.send(destination)
	GaugeMeasure{float64(memStats.NextGC), "NextGC"}.send(destination)
	GaugeMeasure{float64(memStats.OtherSys), "OtherSys"}.send(destination)
	GaugeMeasure{float64(memStats.PauseTotalNs), "PauseTotalNs"}.send(destination)
	GaugeMeasure{float64(memStats.StackInuse), "StackInuse"}.send(destination)
	GaugeMeasure{float64(memStats.StackSys), "StackSys"}.send(destination)
	GaugeMeasure{float64(memStats.Sys), "Sys"}.send(destination)
	GaugeMeasure{float64(memStats.TotalAlloc), "TotalAlloc"}.send(destination)

	CounterMeasure{d.pollCount, "PollCount"}.send(destination)
	d.pollCount += 1

	GaugeMeasure{rand.Float64(), "RandomeValue"}.send(destination)

}

type measuresServer struct {
	*resty.Client
}

func (s *measuresServer) saveCounter(m Measure, v int64) {
	s.R().Post("/counter/" + m.name() + "/" + fmt.Sprintf("%d", v))
}

func (s *measuresServer) saveGauge(m Measure, v float64) {
	s.R().Post("/gauge/" + m.name() + "/" + fmt.Sprintf("%f", v))
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}

	return a
}

func send(measures []Measure, destination MeasureDestination) {
	for _, m := range measures {
		m.send(destination)
	}
}

func main() {

	defaultMeasured := &defaultMeasured{}
	measuresBuffer := &measuresBuffer{}
	measuresServer := &measuresServer{
		resty.New(),
	}

	measuresServer.SetBaseURL("http://localhost:8080/update")
	measuresServer.SetDebug(true)

	const pollInterval = 2
	const reportInterval = 10

	intervalsGcd := gcd(pollInterval, reportInterval)
	timeSpent := 0

	for true {
		time.Sleep(time.Duration(intervalsGcd) * time.Second)
		timeSpent += intervalsGcd

		if timeSpent%pollInterval == 0 {
			defaultMeasured.captureMetrics(measuresBuffer)
		}

		if timeSpent%reportInterval == 0 {

			metricsToSend := make([]Measure, len(measuresBuffer.buffer))
			copy(metricsToSend, measuresBuffer.buffer)

			go send(metricsToSend, measuresServer)
			measuresBuffer.buffer = measuresBuffer.buffer[:0]
		}

		if timeSpent%reportInterval == 0 && timeSpent%pollInterval == 0 {
			timeSpent = 0
		}

		defaultMeasured.captureMetrics(measuresBuffer)
	}
}
