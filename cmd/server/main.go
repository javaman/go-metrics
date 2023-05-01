package main

import (
	"net/http"
	"strconv"
	"strings"
)

type MemStorage struct {
	gauge   float64
	counter int64
}

func (h *MemStorage) ServeHTTP(res http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) <= 2 || (parts[2] != "gauge" && parts[2] != "counter") {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(parts) == 4 && parts[2] == "gauge" {
		if _, err := strconv.ParseFloat(parts[3], 64); err == nil {
			res.WriteHeader(http.StatusNotFound)
		} else {
			res.WriteHeader(http.StatusBadRequest)
		}
		return
	}
	if len(parts) == 4 && parts[2] == "counter" {
		if _, err := strconv.ParseInt(parts[3], 10, 64); err == nil {
			res.WriteHeader(http.StatusNotFound)
		} else {
			res.WriteHeader(http.StatusBadRequest)
		}
		return
	}
	res.Header().Set("Content-Type", "text/plain")
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/update/", new(MemStorage))

	err := http.ListenAndServe("127.0.0.1:8080", mux)

	if err != nil {
		panic(err)
	}
}
