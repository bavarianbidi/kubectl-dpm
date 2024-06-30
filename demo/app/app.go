// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	prometheus.MustRegister(requestsCounter)
}

var version string = os.Getenv("SERVER_VERSION")
var message string = os.Getenv("SERVER_MESSAGE")
var title string = os.Getenv("SERVER_TITLE")

var listenAddr = flag.String("listen", ":8080", "The address to listen on for web requests")
var metricAddr = flag.String("metric", ":9090", "The address to listen on for metric pulls.")
var templatePath = flag.String("template", "/conf/index.tmpl", "The path to the template file")

var requestsCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "A counter for received requests",
		ConstLabels: map[string]string{
			"version": version,
		},
	},
	[]string{"code", "method"})

func serveHTTP(s *http.Server) {
	log.Printf("Server started at %s", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Printf("Starting server failed")
	}
}

func serveMetrics(addr string) {
	log.Printf("Serving metrics on port %s", addr)
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("Starting Prometheus listener failed")
	}
}

type Todo struct {
	Title string
	Done  bool
}

type pageTemplateData struct {
	Title         string
	Message       string
	Hostname      string
	Version       string
	RemoteAddress string
}

func httpHandler(w http.ResponseWriter, req *http.Request) {

	tmpl := template.Must(template.ParseFiles(*templatePath))

	hostname, err := os.Hostname()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	remoteAddress, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := pageTemplateData{
		Title:         title,
		Message:       message,
		Hostname:      hostname,
		Version:       version,
		RemoteAddress: remoteAddress,
	}

	if err = tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func probeHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "ok")
}

func metricsHandler(handler http.Handler) http.Handler {
	return promhttp.InstrumentHandlerCounter(requestsCounter, handler)
}

func main() {

	// flags
	flag.Parse()

	// logs
	log.Printf("Web App Version: %s\n", version)

	// mux
	mux := http.NewServeMux()
	mux.HandleFunc("/", httpHandler)
	mux.HandleFunc("/ready", probeHandler)
	mux.HandleFunc("/live", probeHandler)

	srv := &http.Server{
		Addr:    *listenAddr,
		Handler: metricsHandler(mux),
	}

	go serveHTTP(srv)
	go serveMetrics(*metricAddr)

	// graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	log.Println("Graceful shutdown complete.")
}
