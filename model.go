package shieldoo_lighthouse

import (
	"time"
)

type NebulaLocalYamlConfig struct {
	ConfigHash string                    `json:"config_hash"`
	ConfigData *ManagementResponseConfig `json:"config_data"`
	Loaded     bool                      `json:"-"`
}

type OAuthLoginRequest struct {
	AccessID  int    `json:"access_id"`
	Timestamp int64  `json:"timestamp"`
	Key       string `json:"key"`
}

type OAuthLighthouseLoginRequest struct {
	PublicIp  string `json:"publicip"`
	Timestamp int64  `json:"timestamp"`
	Key       string `json:"key"`
}

type OAuthLoginResponse struct {
	JWTToken string    `json:"jwt"`
	ValidTo  time.Time `json:"valid_to"`
}

type ManagementRequest struct {
	AccessID      int       `json:"access_id"`
	ConfigHash    string    `json:"confighash"`
	DnsHash       string    `json:"dnshash"`
	Timestamp     time.Time `json:"timestamp"`
	LogData       string    `json:"log_data"`
	IsConnected   bool      `json:"is_connected"`
	OverWebSocket bool      `json:"over_websocket"`
}

type ManagementResponseConfigLocalLighthouse struct {
	Port      int    `json:"port"`
	IPAddress string `json:"ipaddress"`
}

type ManagementResponseConfigData struct {
	Data      string `json:"config"`
	Hash      string `json:"hash"`
	IPAddress string `json:"ipaddress"`
}

type ManagementResponseConfig struct {
	AccessID                  int                                     `json:"accessid"`
	Name                      string                                  `json:"name"`
	ConfigData                ManagementResponseConfigData            `json:"config"`
	UnderlayConfigData        ManagementResponseConfigData            `json:"underlayconfig"`
	LocalLighthouse           ManagementResponseConfigLocalLighthouse `json:"locallighthouse"`
	NebulaPunchBack           bool                                    `json:"nebulapunchback"`
	NebulaRestrictiveNetwork  bool                                    `json:"nebularestrictivenetwork"`
	WebSocketUrl              string                                  `json:"websocketurl"`
	WebSocketIPs              []string                                `json:"websocketips"`
	WebSocketUsernamePassword string                                  `json:"websocketusernamepassword"`
	ApplianceListeners        []ManagementResponseListener            `json:"listeners"`
}

type ManagementResponseListener struct {
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	ForwardPort int    `json:"forwardport"`
	ForwardHost string `json:"forwardhost"`
}

type ManagementResponse struct {
	Status     string                    `json:"status"`
	ConfigData *ManagementResponseConfig `json:"config_data"`
	Dns        *ManagementResponseDNS    `json:"dns"`
}

type ManagementResponseDNS struct {
	DnsRecords []string `json:"dnsrecords"`
	DnsHash    string   `json:"dnshash"`
}
