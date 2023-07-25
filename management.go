package shieldoo_lighthouse

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/shieldoo/shieldoo-mesh-lighthouse/dns"
)

var gtelLogin OAuthLoginResponse

func telemetryLogin() error {
	if gtelLogin.ValidTo.UTC().Local().Add(-300 * time.Second).Before(time.Now().UTC()) {
		uri := myconfig.Uri + "api/oauth/authorizelighthouse"
		log.Info("Login  to management server: ", uri)
		timst := time.Now().UTC().Unix()
		keymaterial := strconv.FormatInt(timst, 10) + "|" + myconfig.Secret
		hash := sha256.Sum256([]byte(keymaterial))
		var jsonReq []byte
		req := OAuthLighthouseLoginRequest{
			PublicIp:  myconfig.PublicIP,
			Timestamp: timst,
			Key:       base64.URLEncoding.EncodeToString(hash[:]),
		}
		jsonReq, _ = json.Marshal(req)
		log.Debug("Login message: ", string(jsonReq))
		response, err := http.Post(uri, "application/json; charset=utf-8", bytes.NewBuffer(jsonReq))
		if err != nil {
			log.Error("Login error - post: ", err)
			return err
		}
		log.Debug("Login http status: ", response.Status)
		if response.StatusCode != 200 {
			return errors.New("http error")
		}
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Error("Login error - post/read: ", err)
			return err
		}
		log.Debug("login response bytes: ", string(bodyBytes))
		err = json.Unmarshal(bodyBytes, &gtelLogin)
		if err != nil {
			log.Error("Login error - post/unmarshal: ", err)
			log.Error("Login error - post/unmarshal - body: ", string(bodyBytes))
		}
		return err
	}
	return nil
}

func telemetryProcessChanges(cfg *ManagementResponseConfig) {
	// save configs and certs
	localconf.ConfigHash = cfg.ConfigData.Hash
	localconf.ConfigData = cfg
	localconf.Loaded = true
}

func telemetryCollectLogData() string {
	_log := ""

	// collect telemtry data
	_logreading := true
	// in first step we will try to read any data from channel, if there is nothing we will wait for defined time
	select {
	case _l := <-logdata:
		if _l != "" {
			_log += _l + "\n"
		}
	case <-time.After(time.Duration(myconfig.SendInterval) * 1000 * time.Millisecond):
	}
	// there we will try to read rest of data from channel
	for _logreading {
		select {
		case _l := <-logdata:
			if _l != "" {
				_log += _l + "\n"
			}
		case <-time.After(100 * time.Millisecond):
			_logreading = false
		}
	}
	return _log
}

func telemetrySend() (ret bool) {
	_log := ""
	connected := false

	// exception handling
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			log.Error("telemetry error: ", err)
			ret = false
			// return log data to memory for next time
			logdata <- _log
		}
		prometheusSetConnectionReady(connected)
		if svcProcess != nil && svcProcess.nebula != nil {
			listHostmap := svcProcess.nebula.ListHostmapHosts(false)
			prometheusOpenTunnels.Set(float64(len(listHostmap)))
		}
	}()

	// collect telemtry data
	_log = telemetryCollectLogData()

	ret = false
	// sned telemetry
	if e := telemetryLogin(); e == nil {
		uri := myconfig.Uri + "api/management/messagelighthouse"
		log.Debug("Sending telemetry to: ", uri)
		request := ManagementRequest{
			AccessID:      myconfig.AccessId,
			ConfigHash:    localconf.ConfigHash,
			DnsHash:       dnsconf.DnsHash,
			Timestamp:     time.Now().UTC(),
			LogData:       _log,
			OverWebSocket: false,
			IsConnected:   true,
		}
		jsonReq, _ := json.Marshal(request)
		log.Debug("http req: ", string(jsonReq))

		req, _ := http.NewRequest("POST", uri, bytes.NewBuffer(jsonReq))
		req.Header.Set("Authorization", "Bearer "+gtelLogin.JWTToken)
		req.Header.Add("Accept", "application/json; charset=utf-8")
		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		log.Debug("http resp: ", response.Status)
		if response.StatusCode == 401 {
			gtelLogin.ValidTo = time.Now().UTC().Add(-1000 * time.Hour)
			panic(errors.New("unauthorized call to management API (401)"))
		} else if response.StatusCode != 200 {
			panic(errors.New("status code from management API != 200: " + response.Status))
		}
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			panic(err)
		}
		resp := ManagementResponse{}
		err = json.Unmarshal(bodyBytes, &resp)
		if err != nil {
			panic(err)
		}
		if resp.Dns != nil {
			log.Info("Save new DNS config data")
			dnsconf = *resp.Dns
			// apply new DNS records
			dns.ApplyDNSRecords(dnsconf.DnsRecords)
		}
		if resp.ConfigData != nil {
			log.Info("Save new config data")
			telemetryProcessChanges(resp.ConfigData)
			ret = true
		}
		connected = true
	}
	return
}
