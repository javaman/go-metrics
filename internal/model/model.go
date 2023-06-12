package model

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (m *Metrics) IsDeltaCounter() bool {
	return m.MType == Counter
}

func (m *Metrics) IsValueGauge() bool {
	return m.MType == Gauge
}
