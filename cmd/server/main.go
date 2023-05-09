package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

type MemStorage struct {
	gauge   float64
	counter int64
}

func validate(r *http.Request) (string, bool, float64, int64, int) {
	if r.Method != http.MethodPost {
		return "", false, 0, 0, http.StatusMethodNotAllowed
	} else {
		parts := strings.Split(r.URL.Path, "/")
		metricType := ""
		metricName := ""
		metricValue := ""
		if len(parts) >= 3 {
			metricType = strings.TrimSpace(parts[2])
		}
		if len(parts) >= 4 {
			metricName = strings.TrimSpace(parts[3])
		}
		if len(parts) >= 5 {
			metricValue = strings.TrimSpace(parts[4])
		}
		return validateParams(metricType, metricName, metricValue)
	}
}

func validateParams(metricType string, metricName string, metricValue string) (name string, isGauge bool, gaugeValue float64, counter int64, status int) {
	name = metricName
	isGauge = false
	gaugeValue = 0
	counter = 0
	status = http.StatusOK

	switch metricType {
	case "gauge":
		isGauge = true
	case "counter":
		isGauge = false
	default:
		status = http.StatusBadRequest
		return
	}

	if metricName == "" {
		status = http.StatusNotFound
		return
	}

	if isGauge {
		if value, err := strconv.ParseFloat(metricValue, 64); err == nil {
			gaugeValue = value
		} else {
			status = http.StatusBadRequest
			return
		}
	} else {
		if value, err := strconv.ParseInt(metricValue, 10, 64); err == nil {
			counter = value
		} else {
			status = http.StatusBadRequest
			return
		}
	}

	return
}

func (h *MemStorage) ServeHTTP(res http.ResponseWriter, r *http.Request) {
	_, _, _, _, status := validate(r)
	log.Print("Got Request!")
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(status)
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/update/", new(MemStorage))

	err := http.ListenAndServe("127.0.0.1:8080", mux)

	if err != nil {
		panic(err)
	}
}
