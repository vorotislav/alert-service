package gauge

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
)

type Storage interface {
	UpdateGauge(name string, value float64) error
}

type Handler struct {
	s Storage
}

func NewHandler(storage Storage) *Handler {
	return &Handler{
		s: storage,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	re := regexp.MustCompile(`^/update/gauge/(?P<name>\w+)/(?P<value>[^/]+)$`)
	values := re.FindStringSubmatch(r.RequestURI)

	if len(values) != 3 {
		http.Error(w, "the path is specified incorrectly", http.StatusNotFound)

		return
	}

	metricName := values[1]
	metricValue, err := strconv.ParseFloat(values[2], 64)

	if err != nil {
		http.Error(w, fmt.Errorf("cannot convert metric value: %w", err).Error(), http.StatusBadRequest)

		return
	}

	if err := h.s.UpdateGauge(metricName, metricValue); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")

	w.WriteHeader(http.StatusOK)
}
