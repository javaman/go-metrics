package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/javaman/go-metrics/internal/config"
	"github.com/javaman/go-metrics/internal/domain"
	"github.com/javaman/go-metrics/internal/tools"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type MeasureDestination interface {
	saveCounter(m Measure, value int64)
	saveGauge(m Measure, value float64)
	finishBatch()
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

func (mb *measuresBuffer) finishBatch() {
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
	GaugeMeasure{float64(memStats.NumForcedGC), "NumForcedGC"}.save(destination)
	GaugeMeasure{float64(memStats.MCacheSys), "MCacheSys"}.save(destination)
	GaugeMeasure{float64(memStats.MSpanInuse), "MSpanInuse"}.save(destination)
	GaugeMeasure{float64(memStats.NumGC), "NumGC"}.save(destination)

	CounterMeasure{d.pollCount, "PollCount"}.save(destination)
	d.pollCount += 1

	GaugeMeasure{rand.Float64(), "RandomValue"}.save(destination)

	v, _ := mem.VirtualMemory()
	GaugeMeasure{float64(v.Total), "TotalMemory"}.save(destination)
	GaugeMeasure{float64(v.Free), "FreeMemory"}.save(destination)

	m, _ := cpu.Percent(0, true)

	for i, cpuUtilization := range m {
		GaugeMeasure{cpuUtilization, fmt.Sprintf("CPUUtilization%d", i)}.save(destination)
	}
}

type batchedMeasuresServer struct {
	*resty.Client
	measures []domain.Metric
	key      string
}

func (s *batchedMeasuresServer) saveCounter(m Measure, v int64) {
	s.measures = append(s.measures, domain.Metric{ID: m.name(), MType: "counter", Delta: &v})
}

func (s *batchedMeasuresServer) saveGauge(m Measure, v float64) {
	s.measures = append(s.measures, domain.Metric{ID: m.name(), MType: "gauge", Value: &v})
}

func (s *batchedMeasuresServer) finishBatch() {
	delays := [...]int{1, 3, 5}
	encoded, _ := json.Marshal(s.measures)
	for _, delay := range delays {
		request := s.R()
		if len(s.key) > 0 {
			request.SetHeader("HashSHA256", tools.ComputeSign(encoded, s.key))
		}
		_, err := request.
			SetHeader("Content-Type", "application/json").
			SetBody(string(encoded[:])).
			Post("/")
		if err == nil {
			break
		}
		var opError *net.OpError
		if !errors.As(err, &opError) {
			break
		}
		time.Sleep(time.Duration(delay) * time.Second)
	}
}

func worker(jobs <-chan []Measure, destination MeasureDestination) {
	for measures := range jobs {
		for _, measure := range measures {
			measure.save(destination)
		}
		destination.finishBatch()
	}
}

func main() {
	conf := config.ConfigureAgent()

	defaultMeasured := &defaultMeasured{}
	measuresBuffer := &measuresBuffer{}
	measuresServer := &batchedMeasuresServer{
		resty.New(),
		make([]domain.Metric, 1),
		conf.Key,
	}
	measuresServer.SetBaseURL("http://" + conf.Address + "/updates")

	jobs := make(chan []Measure, conf.RateLimit)

	defer func() {
		close(jobs)
	}()

	for i := 0; i < conf.RateLimit; i++ {
		go worker(jobs, measuresServer)
	}

	go func() {
		for {
			time.Sleep(time.Duration(conf.PollInterval) * time.Second)
			defaultMeasured.captureMetrics(measuresBuffer)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Duration(conf.ReportInterval) * time.Second)
			metricsToSend := make([]Measure, len(measuresBuffer.buffer))
			copy(metricsToSend, measuresBuffer.buffer)
			jobs <- metricsToSend
		}
	}()

	ctx := context.Background()
	<-ctx.Done()
}
