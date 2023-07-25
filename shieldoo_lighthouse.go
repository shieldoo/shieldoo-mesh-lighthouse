package shieldoo_lighthouse

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shieldoo/shieldoo-mesh-lighthouse/dns"
	"github.com/sirupsen/logrus"
)

// global log data which are send to server during telemtry collection
var logdata chan string

// instance IP address
var myipaddress string

// global logging
var log *logrus.Logger

var APPVERSION = "0.0.0"

func Init() {
	// initialize logrus
	log = logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.Info("shieldoo-mesh version: ", APPVERSION)

	logdata = make(chan string, 10000)

	InitConfig()

	if myconfig.Debug {
		log.SetLevel(logrus.DebugLevel)
	}
}

func Run() {
	go WSTunnelRun()

	// start webserver
	log.Info("Starting webserver on port 80")
	prometheus.Register(prometheusConnectionReady)
	prometheus.Register(prometheusOpenTunnels)
	prometheus.Register(prometheusWSSStatus)
	prometheusSetConnectionReady(false)
	prometheusOpenTunnels.Set(0)
	go runWeb()

	// start DNS server
	log.Info("Starting DNS server on port ", myconfig.DnsLocalListener)
	go dns.Run(
		log,
		myconfig.DnsUpstreamServer,
		myconfig.DnsLocalProtocol,
		myconfig.DnsLocalListener,
		myconfig.DnsLocalProtocol)

	// start service
	SvcConnectionStart()
}
