package shieldoo_lighthouse

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/slackhq/nebula"
	"github.com/slackhq/nebula/config"
)

type ChannelWriter struct {
	canwrite bool
}

func (p *ChannelWriter) Write(data []byte) (n int, err error) {
	if p.canwrite {
		s := string(data)
		// ignore timeout lighthouse messages
		if (strings.Contains(s, `"msg":"Handshake timed out"`)) ||
			strings.Contains(s, `"msg":"Host roamed to new udp ip/port."`) ||
			strings.Contains(s, `"msg":"Failed to write to tun"`) {
			// ignored messages
			fmt.Printf("NEBULA-: %s", data)
		} else {
			// collected messages
			fmt.Printf("NEBULA+: %s", data)
			logdata <- string(data)
		}
	} else {
		fmt.Printf("NEBULA#: %s", data)
	}
	return len(data), nil
}

type SvcNetworkCard struct {
	AccessID   int
	ConfigHash string
	IPAddress  string
	nebula     *nebula.Control
	ncfg       *config.C
	log        ChannelWriter
	nl         *logrus.Logger
}

func (r *SvcNetworkCard) Stop() {
	// get tun/tap name
	ttname := r.ncfg.GetString("tun.dev", "")

	log.Debug("stopping nebula ..")
	if r.nebula != nil {
		r.nebula.Stop()
	}
	r.nebula = nil
	runtime.GC()

	// wait for a while to deallocate TUN/TAP
	time.Sleep(1000 * time.Millisecond)
	if ttname != "" {
		if runtime.GOOS == "linux" {
			for i := 1; i < 300; i++ {
				time.Sleep(1000 * time.Millisecond)

				// check if adapter exists - if yes than we have to wait to disapear
				if _, err := os.Stat("/sys/class/net/" + ttname + "/mtu"); err != nil {
					break
				}
				log.Debug("waiting for tun/tap disappear: ", ttname)

				// trying to delete interface
				if i%10 == 0 {
					cmd := exec.Command("ip", "link", "delete", ttname)
					log.Info("deleting tun/tap: ", ttname)
					err := cmd.Run()
					if err != nil {
						log.Error("cannot execute ip link delete: ", err)
					}
				}
			}
		}
	}

	log.Debug("nebula stopped")
	log.Debug("stoped nebula with ip ", r.IPAddress)
}

var svcProcess *SvcNetworkCard = nil

func svcCleanupProcesses(cfg *NebulaLocalYamlConfig) {
	if svcProcess != nil {
		if svcProcess.AccessID != cfg.ConfigData.AccessID /* accessID changed */ ||
			svcProcess.IPAddress != cfg.ConfigData.ConfigData.IPAddress /* IP address of tun/tap changed */ {
			// there is change in config which will recreate network adapter
			svcStopProcess()
		}
	}
}

func svcStopProcess() {
	log.Debug("stopping service: ", svcProcess.IPAddress)
	// stop standard nebula layer
	if svcProcess != nil {
		svcProcess.Stop()
		svcProcess = nil
		runtime.GC()
	}
}

func svcNewProcess(c *ManagementResponseConfig) (SvcNetworkCard, error) {
	ret := SvcNetworkCard{
		AccessID:   c.AccessID,
		ConfigHash: c.ConfigData.Hash,
		IPAddress:  c.ConfigData.IPAddress,
	}

	log.Debug("create service: ", c.ConfigData.IPAddress)

	// ### start process

	cfgtext, err := NebulaConfigCreate(c.ConfigData.Data)
	if err != nil {
		log.Error("cannot create nebula config: ", err)
		return ret, err
	}

	ret.log.canwrite = false
	ret.nl = logrus.New()
	ret.nl.Out = &ret.log
	ret.ncfg = config.NewC(ret.nl)
	err = ret.ncfg.LoadString(cfgtext)

	if err != nil {
		log.Error("failed to load config: ", err)
		return ret, err
	}

	for i := 1; i <= 256; i++ {
		ctrl, err := nebula.Main(ret.ncfg, false, APPVERSION, ret.nl, nil)
		if err == nil {
			ret.nebula = ctrl
			break
		}
		if err != nil && i == 256 {
			log.Error("failed to start nebula: ", err)
			return ret, err
		}
		ctrl = nil
		log.Error("repeating start of nebula: ", err)
		time.Sleep(1000 * time.Millisecond * time.Duration(i))
	}
	ret.log.canwrite = true
	log.Debug("start nebula with ip ", ret.IPAddress)
	ret.nebula.Start()

	// wait for a while to create TUN/TAP
	time.Sleep(500 * time.Millisecond)
	return ret, nil
}

func svcUpdateProcesses(cfg *NebulaLocalYamlConfig) bool {
	log.Debug("updating services ..")
	if cfg.ConfigData != nil {
		// create new nebula process if needed
		if svcProcess == nil {
			svcProcess = nil
			// standard proccess
			newp, err := svcNewProcess(cfg.ConfigData)
			if err != nil {
				newp.Stop()
				return false
			}
			svcProcess = &newp
		} else {
			// update properties of running nebula
			if svcProcess.ConfigHash != cfg.ConfigHash {
				// create config files
				cfgtext, err := NebulaConfigCreate(cfg.ConfigData.ConfigData.Data)
				if err != nil {
					log.Error("cannot create nebula config: ", err)
					return false
				}
				log.Debug("updating services ..")
				err = svcProcess.ncfg.ReloadConfigString(cfgtext)
				if err != nil {
					log.Error("failed to reload config: ", err)
					svcStopProcess()
					return false
				}
				log.Debug("reload config for nebula with ip ", svcProcess.AccessID)
				svcProcess.ConfigHash = cfg.ConfigHash
			}
		}
	}
	return true
}

func configureServices() bool {
	log.Debug("create service..")
	myipaddress = localconf.ConfigData.ConfigData.IPAddress
	var ret bool = true
	// cleanup not existing network configs or changed ..
	svcCleanupProcesses(&localconf)
	// create new processes if needed and update workers..
	if !svcUpdateProcesses(&localconf) {
		ret = false
	}
	return ret
}

var svcconnCancel bool = false
var svcconnIsRunning bool = false
var svcconnStopped chan bool

func SvcConnectionStart() {
	log.Debug("svcconnection starting ..")
	if svcconnIsRunning {
		return
	}
	log.Debug("svcconnection starting ....")
	svcconnCancel = false
	var isinititalized bool = false
	svcconnStopped = make(chan bool)
	// insert into log channel empty string to initialize immediate sending after startup
	logdata <- ""
	svcconnIsRunning = true
	for {
		// run telemetry and config
		log.Debug("waiting for next telemetry send ..")
		if telemetrySend() ||
			!isinititalized {
			if localconf.Loaded {
				// need restart or its first time
				isinititalized = configureServices()
			}
		}
		if svcconnCancel {
			// stop services
			svcStopProcess()
			// send stop signal
			svcconnStopped <- true
			break
		}
	}
	svcconnIsRunning = false
}

func SvcConnectionStop() {
	log.Debug("svcconnection stopping ..")
	if svcconnCancel || !svcconnIsRunning {
		return
	}
	log.Debug("svcconnection stopping ....")

	svcconnCancel = true
	// invoke break of waiting loop in telemtrySend
	logdata <- ""

	// wait for stop nebula connections
	if svcconnStopped != nil {
		<-svcconnStopped
	}
	svcconnCancel = false
}
