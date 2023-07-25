package shieldoo_lighthouse

import (
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func prometheusSetConnectionReady(status bool) {
	if status {
		prometheusConnectionReady.Set(1)
	} else {
		prometheusConnectionReady.Set(0)
	}
}

// once per 15 seconds check status of WSS connection
func prometheusCheckWSSStatus() {
	for {
		uri := "https://" + myconfig.WssDnsName
		resp, err := http.Get(uri)
		status := -1
		if err != nil {
			log.Errorf("unable to get WSS status: %v", err)
		} else {
			log.Debug("WSS status: ", resp.StatusCode)
			status = resp.StatusCode
		}
		prometheusWSSStatus.Set(float64(status))
		time.Sleep(15 * time.Second)
	}
}

var prometheusConnectionReady = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "shieldoo_lighthouse_connection_ready",
		Help: "Shieldoo Lighthouse connection ready",
	},
)

var prometheusOpenTunnels = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "shieldoo_lighthouse_open_tunnels",
		Help: "Shieldoo Lighthouse open tunnels",
	},
)

var prometheusWSSStatus = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "shieldoo_lighthouse_wss_status",
		Help: "Shieldoo Lighthouse WSS status",
	},
)

func IndexHandler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadFile("shieldoo.html")
	if err != nil {
		log.Errorf("unable to read file: %v", err)
	}
	bodys := strings.ReplaceAll(string(body), "{IPADDRESS}", myipaddress)
	io.WriteString(w, bodys)
}

func HealthHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "OK")
}

func runWeb() {
	r := mux.NewRouter()
	go prometheusCheckWSSStatus()
	r.HandleFunc("/", IndexHandler).Methods("GET")
	r.HandleFunc("/health", HealthHandler).Methods("GET")
	r.Handle("/metrics", promhttp.Handler())
	http.Handle("/", r)
	log.Error(http.ListenAndServe(":80", nil))
}
