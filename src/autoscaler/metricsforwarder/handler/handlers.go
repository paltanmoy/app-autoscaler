package handler

import (
	"autoscaler/metricsforwarder/forwarder"
	"autoscaler/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type MetricsHandler struct {
	metricForwarder forwarder.MetricForwarder
}

func NewMetricsHandler(metricForwarder forwarder.MetricForwarder) *MetricsHandler {
	return &MetricsHandler{
		metricForwarder: metricForwarder,
	}
}

func (mh *MetricsHandler) PublishMetrics(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Serving POST Metrics")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	var metric models.CustomMetric
	err = json.Unmarshal(body, &metric)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	//metrics = append(metrics, metric)
	mh.metricForwarder.EmitMetric(metric)
	w.WriteHeader(http.StatusCreated)
}
