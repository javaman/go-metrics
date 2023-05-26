package main

import (
	"encoding/json"
	"math/rand"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/javaman/go-metrics/internal/config"
	"github.com/javaman/go-metrics/internal/model"
)

type MeasureDestination interface {
	saveCounter(m Measure, value int64)
	saveGauge(m Measure, value float64)
}

type Measure interface {
	save(to MeasureDestination)
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

func (m GaugeMeasure) save(to MeasureDestination) {
	to.saveGauge(m, m.value)
}

func (m CounterMeasure) name() string {
	return m.mname
}

func (m CounterMeasure) save(to MeasureDestination) {
	to.saveCounter(m, m.value)
}

type Measured interface {
	captureMetrics(to MeasureDestination)
}

type measuresBuffer struct {
	buffer []Measure
}

func (mb *measuresBuffer) saveCounter(m Measure, value int64) {
	mb.buffer = append(mb.buffer, m)
}

func (mb *measuresBuffer) saveGauge(m Measure, value float64) {
	mb.buffer = append(mb.buffer, m)
}

type defaultMeasured struct {
	pollCount int64
}

func (d *defaultMeasured) captureMetrics(destination MeasureDestination) {
	memStats := new(runtime.MemStats)
	runtime.ReadMemStats(memStats)

	GaugeMeasure{float64(memStats.Alloc), "Alloc"}.save(destination)
	GaugeMeasure{float64(memStats.BuckHashSys), "BuckHashSys"}.save(destination)
	GaugeMeasure{float64(memStats.Frees), "Frees"}.save(destination)
	GaugeMeasure{memStats.GCCPUFraction, "GCCPUFraction"}.save(destination)
	GaugeMeasure{float64(memStats.GCSys), "GCSys"}.save(destination)
	GaugeMeasure{float64(memStats.HeapAlloc), "HeapAlloc"}.save(destination)
	GaugeMeasure{float64(memStats.HeapIdle), "HeapIdle"}.save(destination)
	GaugeMeasure{float64(memStats.HeapInuse), "HeapInuse"}.save(destination)
	GaugeMeasure{float64(memStats.HeapObjects), "HeapObjects"}.save(destination)
	GaugeMeasure{float64(memStats.HeapReleased), "HeapReleased"}.save(destination)
	GaugeMeasure{float64(memStats.HeapSys), "HeapSys"}.save(destination)
	GaugeMeasure{float64(memStats.LastGC), "LastGC"}.save(destination)
	GaugeMeasure{float64(memStats.Lookups), "Lookups"}.save(destination)
	GaugeMeasure{float64(memStats.MCacheInuse), "MCacheInuse"}.save(destination)
	GaugeMeasure{float64(memStats.MSpanSys), "MSpanSys"}.save(destination)
	GaugeMeasure{float64(memStats.Mallocs), "Mallocs"}.save(destination)
	GaugeMeasure{float64(memStats.NextGC), "NextGC"}.save(destination)
	GaugeMeasure{float64(memStats.OtherSys), "OtherSys"}.save(destination)
	GaugeMeasure{float64(memStats.PauseTotalNs), "PauseTotalNs"}.save(destination)
	GaugeMeasure{float64(memStats.StackInuse), "StackInuse"}.save(destination)
	GaugeMeasure{float64(memStats.StackSys), "StackSys"}.save(destination)
	GaugeMeasure{float64(memStats.Sys), "Sys"}.save(destination)
	GaugeMeasure{float64(memStats.TotalAlloc), "TotalAlloc"}.save(destination)

	CounterMeasure{d.pollCount, "PollCount"}.save(destination)
	d.pollCount += 1

	GaugeMeasure{rand.Float64(), "RandomValue"}.save(destination)

}

type measuresServer struct {
	*resty.Client
}

func (s *measuresServer) saveCounter(m Measure, v int64) {
	j := &model.Metrics{}
	j.ID = m.name()
	j.MType = "counter"
	j.Delta = &v
	encoded, _ := json.Marshal(*j)
	s.R().
		SetHeader("Content-Type", "application/json").
		SetBody(string(encoded[:])).
		Post("/")
}

func (s *measuresServer) saveGauge(m Measure, v float64) {
	j := &model.Metrics{}
	j.ID = m.name()
	j.MType = "gauge"
	j.Value = &v
	encoded, _ := json.Marshal(j)
	s.R().
		SetHeader("Content-Type", "application/json").
		SetBody(string(encoded[:])).
		Post("/")
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}

	return a
}

func send(measures []Measure, destination MeasureDestination) {
	for _, m := range measures {
		m.save(destination)
	}
}

type Worker interface {
	run()
}

type defaultWorker struct {
	pollInterval   int
	reportInterval int
	pollFunction   func()
	reportFunction func()
	advance        func() bool
}

func (w *defaultWorker) run() {
	intervalsGcd := gcd(w.pollInterval, w.reportInterval)
	var timeSpent int
	for w.advance() {
		time.Sleep(time.Duration(intervalsGcd) * time.Second)
		timeSpent += intervalsGcd
		if timeSpent%w.pollInterval == 0 {
			w.pollFunction()
		}
		if timeSpent%w.reportInterval == 0 {
			w.reportFunction()
		}
		if timeSpent%w.reportInterval == 0 && timeSpent%w.pollInterval == 0 {
			timeSpent = 0
		}
	}
}

func main() {
	conf := config.ConfigureAgent()

	defaultMeasured := &defaultMeasured{}
	measuresBuffer := &measuresBuffer{}
	measuresServer := &measuresServer{
		resty.New(),
	}
	measuresServer.SetBaseURL("http://" + conf.Address + "/update")
	measuresServer.SetDebug(true)

	dw := &defaultWorker{
		conf.PollInterval,
		conf.ReportInterval,
		func() { defaultMeasured.captureMetrics(measuresBuffer) },
		func() {
			metricsToSend := make([]Measure, len(measuresBuffer.buffer))
			copy(metricsToSend, measuresBuffer.buffer)

			send(metricsToSend, measuresServer)
			measuresBuffer.buffer = measuresBuffer.buffer[:0]
		},
		func() bool {
			return true
		},
	}

	dw.run()

}
