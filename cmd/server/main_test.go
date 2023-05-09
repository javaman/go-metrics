package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatuses(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{
			name:           "Bad Request wrong metric type",
			url:            "/update/xxx/name/123",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Not Found empty metric name",
			url:            "/update/counter//123",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Bad Request wrong metric value",
			url:            "/update/counter/name/xxx",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Good request",
			url:            "/update/counter/name/123",
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.url, nil)
			w := httptest.NewRecorder()
			memStorage := &MemStorage{}
			memStorage.ServeHTTP(w, request)

			result := w.Result()

			assert.Equal(t, test.expectedStatus, result.StatusCode)
		})
	}

}
