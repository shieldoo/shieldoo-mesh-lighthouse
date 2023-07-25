package dns

import (
	"context"
	"math/rand"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

// global logging
var log *logrus.Logger

var upstreamServer string
var upstreamProtocol string
var localListener string
var localProtocol string

var dnsRecords []string

func ApplyDNSRecords(records []string) {
	dnsRecords = records
	flush <- struct{}{}
}

func Run(logr *logrus.Logger, upSrv string, upProto string, localLst string, localProto string) {
	log = logr

	upstreamServer = upSrv
	upstreamProtocol = upProto
	localListener = localLst
	localProtocol = localProto

	fillCache()
	go flushCache()
	go timeCache()
	go memCache()
	srv := &dns.Server{Addr: localListener, Net: localProtocol}
	srv.Handler = &dnsMistake{}
	log.Debug("starting dns server on ", localListener, " ", localProtocol)
	log.Debug("upstream server ", upstreamServer, " ", upstreamProtocol)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("fatal: failed to set udp|tcp listener %s\n", err.Error())
	}
}

var dataCache = make(map[string]string)
var datamx = &sync.Mutex{}

type dnsMistake struct{}

var flush = make(chan struct{})

// fill cache from dnsRecords
func fillCache() {
	log.Debug("filling memory cache .. ", len(dnsRecords))
	for _, record := range dnsRecords {
		// split by space
		rec := strings.Split(record, " ")
		if len(rec) == 2 {
			dataCache[rec[1]] = rec[0]
		}
	}
}

func flushCache() {
	for {
		<-flush
		log.Debug("dropping memory cache .. ", len(dataCache))
		dataCache = make(map[string]string)
		fillCache()
	}
}

func timeCache() {
	for {
		time.Sleep(time.Second*time.Duration(rand.Intn(120)) + 90)
		log.Debug("time to flush cache .. ")
		flush <- struct{}{}
	}
}

func memCache() {
	for {
		time.Sleep(time.Millisecond * 100)
		if len(dataCache) > 900000 {
			log.Debug("memory cache too big, flushing .. ", len(dataCache))
			flush <- struct{}{}
		}

	}
}

func (*dnsMistake) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	switch r.Question[0].Qtype {
	case dns.TypeA:
		msg.Authoritative = true
		domain := msg.Question[0].Name
		log.Debug("DNS request for ", domain)
		address, ok := haveInCache(domain)
		if ok {
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP(address),
			})
		}
	}
	w.WriteMsg(&msg)
}

func haveInCache(domain string) (string, bool) {
	for k, v := range dataCache {
		if ok, _ := regexp.MatchString(k, domain); ok {
			log.Debug("DNS cache hit for ", domain, " ", v)
			return v, true
		}
	}

	log.Debug("DNS cache miss for", domain)
	if addr, ok := askUpstream(domain); ok {
		log.Debug("DNS upstream hit for ", domain, " ", addr)
		datamx.Lock()
		dataCache[domain] = addr
		datamx.Unlock()
		return addr, true
	}
	return "err", false
}

func askUpstream(s string) (string, bool) {
	log.Debug("DNS upstream request for ", s)
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, upstreamProtocol, upstreamServer)
		},
	}

	ip, err := r.LookupHost(context.Background(), s)
	if err != nil {
		return "err", false
	}
	return ip[0], true
}
