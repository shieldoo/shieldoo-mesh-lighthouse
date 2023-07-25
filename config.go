package shieldoo_lighthouse

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type NebulaClientYamlConfig struct {
	AccessId            int    `yaml:"accessid", envconfig:"ACCESSID"`
	PublicIP            string `yaml:"publicip", envconfig:"PUBLICIP"`
	Uri                 string `yaml:"uri", envconfig:"URI"`
	Secret              string `yaml:"secret", envconfig:"SECRET"`
	Debug               bool   `yaml:"debug", envconfig:"DEBUG"`
	SendInterval        int    `yaml:"sendinterval", envconfig:"SENDINTERVAL"`
	WebSocketPort       int    `yaml:"websocketport", envconfig:"WEBSOCKETPORT"`
	UdpPort             int    `yaml:"udpport", envconfig:"UDPPORT"`
	DnsUpstreamServer   string `yaml:"dnsupstreamserver", envconfig:"DNSUPSTREAMSERVER"`
	DnsUpstreamProtocol string `yaml:"dnsupstreamprotocol", envconfig:"DNSUPSTREAMPROTOCOL"`
	DnsLocalListener    string `yaml:"dnslocallistener", envconfig:"DNSLOCALLISTENER"`
	DnsLocalProtocol    string `yaml:"dnslocalprotocol", envconfig:"DNSLOCALPROTOCOL"`
	WssDnsName          string `yaml:"wssdnsname", envconfig:"WSSDNSNAME"`
}

var myconfig *NebulaClientYamlConfig
var localconf NebulaLocalYamlConfig
var dnsconf ManagementResponseDNS

const MYCONFIG_FILENAME = "myconfig.yaml"

var execPath string

func execPathCreate(p string) string {
	return filepath.FromSlash(execPath + "/config/" + p)
}

func InitExecPath() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	execPath = filepath.Dir(ex)
}

func CreateConfigFromBase64(str string) (err error) {
	InitExecPath()

	// create config folder if not exists
	_ = os.MkdirAll(execPathCreate(""), 0700)

	var data []byte
	data, err = base64.StdEncoding.DecodeString(str)
	if err != nil {
		return
	}
	err = saveFile(MYCONFIG_FILENAME, data)
	return
}

func InitConfig() {
	InitExecPath()

	// create config folder if not exists
	_ = os.MkdirAll(execPathCreate(""), 0700)

	log.Debug("Loading configs ..")
	// read myconfig.yaml
	mc, err := readClientConf(MYCONFIG_FILENAME)
	if err != nil {
		log.Debug("cannot find "+execPathCreate(MYCONFIG_FILENAME)+" file or configuration file is corrupted: ", err)
	}
	myconfig = mc

	readEnv(myconfig)

	// sanitize config
	if myconfig.SendInterval <= 0 || myconfig.SendInterval > 3600 {
		myconfig.SendInterval = 60
	}

	if myconfig.WebSocketPort <= 0 || myconfig.WebSocketPort > 65535 {
		myconfig.WebSocketPort = 8080
	}

	if myconfig.UdpPort <= 0 || myconfig.UdpPort > 65535 {
		myconfig.UdpPort = 4242
	}
	if !strings.HasSuffix(myconfig.Uri, "/") {
		myconfig.Uri += "/"
	}

	// dns default
	if myconfig.DnsUpstreamServer == "" {
		myconfig.DnsUpstreamServer = "8.8.8.8:53" // google dns
	}
	if myconfig.DnsUpstreamProtocol == "" {
		myconfig.DnsUpstreamProtocol = "udp" // google dns
	}
	if myconfig.DnsLocalListener == "" {
		myconfig.DnsLocalListener = "0.0.0.0:53" // listen on all interfaces
	}
	if myconfig.DnsLocalProtocol == "" {
		myconfig.DnsLocalProtocol = "udp"
	}
}

func readClientConf(filename string) (*NebulaClientYamlConfig, error) {
	c := &NebulaClientYamlConfig{}
	buf, err := ioutil.ReadFile(execPathCreate(filename))
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return c, fmt.Errorf("in file %q: %v", filename, err)
	}

	return c, nil
}

func saveTextFile(filename string, text string) error {
	return saveFile(filename, []byte(text))
}

func saveFile(filename string, data []byte) error {

	file, err := os.OpenFile(execPathCreate(filename), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)

	return err
}

func readEnv(cfg *NebulaClientYamlConfig) {
	err := envconfig.Process("", cfg)
	if err != nil {
		log.Error("readEnv() error: ", err)
	}
}
