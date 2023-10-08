package shieldoo_lighthouse

import (
	"gopkg.in/yaml.v3"
)

type NebulaYamlConfigFW struct {
	Port   string   `yaml:"port"`
	Proto  string   `yaml:"proto"`
	Host   string   `yaml:"host,omitempty"`
	Groups []string `yaml:"groups,omitempty"`
}

type NebulaYamlConfigUnsafeRoutes struct {
	Route string `yaml:"route"`
	Via   string `yaml:"via"`
}

type NebulaYamlConfig struct {
	Pki struct {
		Ca        string   `yaml:"ca"`
		Cert      string   `yaml:"cert"`
		Key       string   `yaml:"key"`
		Blocklist []string `yaml:"blocklist"`
	} `yaml:"pki"`
	StaticHostMap map[string][]string `yaml:"static_host_map"`
	Lighthouse    struct {
		AmLighthouse bool     `yaml:"am_lighthouse"`
		Interval     int      `yaml:"interval"`
		Hosts        []string `yaml:"hosts"`
	} `yaml:"lighthouse"`
	Listen struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"listen"`
	Punchy struct {
		Punch   bool `yaml:"punch"`
		Respond bool `yaml:"respond"`
	} `yaml:"punchy"`
	Relay struct {
		Relays    []string `yaml:"relays"`
		AmRelay   bool     `yaml:"am_relay"`
		UseRelays bool     `yaml:"use_relays"`
	} `yaml:"relay"`
	Tun struct {
		Disabled           bool                           `yaml:"disabled"`
		Dev                string                         `yaml:"dev"`
		DropLocalBroadcast bool                           `yaml:"drop_local_broadcast"`
		DropMulticast      bool                           `yaml:"drop_multicast"`
		TxQueue            int                            `yaml:"tx_queue"`
		Mtu                int                            `yaml:"mtu"`
		Routes             interface{}                    `yaml:"routes"`
		UnsafeRoutes       []NebulaYamlConfigUnsafeRoutes `yaml:"unsafe_routes"`
	} `yaml:"tun"`
	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`
	Firewall struct {
		Conntrack struct {
			TCPTimeout     string `yaml:"tcp_timeout"`
			UDPTimeout     string `yaml:"udp_timeout"`
			DefaultTimeout string `yaml:"default_timeout"`
			MaxConnections int    `yaml:"max_connections"`
		} `yaml:"conntrack"`
		Outbound []NebulaYamlConfigFW `yaml:"outbound"`
		Inbound  []NebulaYamlConfigFW `yaml:"inbound"`
	} `yaml:"firewall"`
}

func NebulaConfigCreate(configdata string) (string, error) {
	log.Debug("save nebula config standard")
	c := &NebulaYamlConfig{}
	var err error
	buf := []byte(configdata)
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		log.Debug("Error deserialize nebula config: ", err)
		return "", err
	}
	c.Punchy.Respond = false
	// configure firewall rules
	c.Firewall.Outbound = []NebulaYamlConfigFW{
		{
			Port:  "any",
			Proto: "any",
			Host:  "any",
		},
	}
	c.Firewall.Inbound = []NebulaYamlConfigFW{
		{
			Port:  "any",
			Proto: "any",
			Host:  "any",
		},
	}
	buf, err = yaml.Marshal(&c)
	return string(buf), err
}
