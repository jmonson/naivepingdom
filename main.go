package main

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//Contains list of sites to probe
var sites Sites

//Contains Prometheus metrics exporter
var exporter *Exporter

//Sites contains the list of sites to probe
type Sites []struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

//Site contains the definition for the site to probe
type Site struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

//Exporter holds the Prometheus metrics to export
type Exporter struct {
	SiteMetrics map[string]SiteMetric
}

//SiteMetric contains the metric setup
type SiteMetric struct {
	PromMetric *prometheus.Desc
	Address    string
}

//newExporter initializes descriptors and returns a pointer to the exporter
func newExporter() *Exporter {
	siteMetrics := make(map[string]SiteMetric)
	for _, site := range sites {
		metric := new(SiteMetric)
		metric.PromMetric = prometheus.NewDesc(
			site.Name+"_http_response_duration",
			"The response time of the HTTP request",
			[]string{"site", "status_code", "content_type"}, nil,
		)
		metric.Address = site.Address
		siteMetrics[site.Name] = *metric
	}
	return &Exporter{siteMetrics}
}

//Describe writes all descriptors to the Prometheus desc channel
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.SiteMetrics {
		ch <- m.PromMetric
	}
}

//Collect implements collect function and sets the metric value to return to Prometheus
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	for site, metric := range e.SiteMetrics {
		start := time.Now()
		resp, err := http.Get(metric.Address)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		elapsed := time.Since(start).Seconds()
		statusCode := resp.StatusCode
		contentType := resp.Header.Get("Content-Type")
		ch <- prometheus.MustNewConstMetric(metric.PromMetric, prometheus.CounterValue, elapsed, site, strconv.Itoa(statusCode), contentType)
	}
}

//siteAPI is restful endpoint for listing and adding new sites
func siteAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		output := "{"
		for _, site := range sites {
			output += fmt.Sprintf("\"%v\":\"%v\",", site.Name, site.Address)
		}
		output = strings.TrimRight(output, ",") + "}"
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(output))
	case "POST":
		decoder := json.NewDecoder(r.Body)
		var newSite Site
		err := decoder.Decode(&newSite)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		sites = append(sites, newSite)
		saveSites()
		resetExporter()
		w.WriteHeader(http.StatusNoContent)
		w.Header().Set("Content-Type", "application/json")
	case "DELETE":
		decoder := json.NewDecoder(r.Body)
		var removeSite Site
		err := decoder.Decode(&removeSite)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		var removeIndex int
		for i, site := range sites {
			if site.Name == removeSite.Name {
				removeIndex = i
				break
			}
		}
		sites = sites[:removeIndex+copy(sites[removeIndex:], sites[removeIndex+1:])]
		saveSites()
		resetExporter()
		w.WriteHeader(http.StatusNoContent)
		w.Header().Set("Content-Type", "application/json")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"message\":\"Method not supported\"}"))
	}
}

//saveSites persists Site information to flat file datastore
func saveSites() {
	file, err := os.Create("config/probe/sites.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.Encode(sites)
}

//readSites reads site data from flat file datastore
func readSites() {
	sites = make(Sites, 0)
	file, err := os.Open("config/probe/sites.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(&sites); err != nil {
		panic(err)
	}
}

//resetExporter resets exporter to refresh Prometheus metrics
func resetExporter() {
	prometheus.Unregister(exporter)
	newExporter := newExporter()
	prometheus.MustRegister(newExporter)
	exporter = newExporter
}

func main() {
	readSites()

	exporter = newExporter()
	prometheus.MustRegister(exporter)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/api/site", siteAPI)
	fmt.Println("Beginning to serve on port :8080")
	http.ListenAndServe(":8080", nil)
}
