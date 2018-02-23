package main

import (
	"autoscaler/metricsforwarder/config"
	"autoscaler/metricsforwarder/forwarder"
	"autoscaler/metricsforwarder/handler"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {

	var path string
	flag.StringVar(&path, "c", "", "config file")
	flag.Parse()
	if path == "" {
		fmt.Fprintln(os.Stderr, "missing config file")
		os.Exit(1)
	}
	configFile, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stdout, "failed to open config file '%s' : %s\n", path, err.Error())
		os.Exit(1)
	}

	var conf *config.Config
	conf, err = config.LoadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stdout, "failed to read config file '%s' : %s\n", path, err.Error())
		os.Exit(1)
	}
	configFile.Close()

	metricForwarder, err := forwarder.NewMetricForwarder(conf)
	if err != nil {
		fmt.Println("failed to create metricforwarder", err)
		os.Exit(1)
	}

	metricHandler := handler.NewMetricsHandler(metricForwarder)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/v1/metrics", metricHandler.PublishMetrics).Methods("POST")
	var port string
	if port = os.Getenv("PORT"); len(port) == 0 {
		port = "3000"
	}
	log.Fatal(http.ListenAndServe(":"+port, router))
}
