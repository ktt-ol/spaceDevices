package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/ktt-ol/spaceDevices/conf"
	log "github.com/sirupsen/logrus"
)

const CLIENT_ID = "spaceDevices2"

var logger = log.WithField("where", "mqtt")

type MqttHandler struct {
	WifiSessionList []WifiSession
	// more to come, e.g. LanSessions
}

type WifiSession struct {
	Ip   string
	Mac  string
	Vlan string
	AP   int
}

//func init() {
//	mqtt.ERROR.SetOutput(copyOfStdLogger(log.ErrorLevel).Writer())
//	mqtt.CRITICAL.SetOutput(copyOfStdLogger(log.ErrorLevel).Writer())
//	mqtt.WARN.SetOutput(copyOfStdLogger(log.WarnLevel).Writer())
//	mqtt.DEBUG.SetOutput(copyOfStdLogger(log.DebugLevel).Writer())
//}
//func copyOfStdLogger(level log.Level) *log.Logger {
//	logger := log.New()
//	logger.Formatter = log.StandardLogger().Formatter
//	logger.Out = log.StandardLogger().Out
//	logger.SetLevel(level)
//	return logger
//}

func EnableMqttDebugLogging() {
	stdLogWriter := log.StandardLogger().Writer()
	mqtt.ERROR.SetOutput(stdLogWriter)
	mqtt.CRITICAL.SetOutput(stdLogWriter)
	mqtt.WARN.SetOutput(stdLogWriter)
	mqtt.DEBUG.SetOutput(stdLogWriter)
}

func NewMqttHandler(conf conf.MqttConf) *MqttHandler {
	opts := mqtt.NewClientOptions()

	opts.AddBroker(conf.Url)

	if conf.Username != "" {
		opts.SetUsername(conf.Username)
	}
	if conf.Password != "" {
		opts.SetPassword(conf.Password)
	}

	certs := defaultCertPool(conf.CertFile)
	tlsConf := &tls.Config{
		RootCAs: certs,
	}
	opts.SetTLSConfig(tlsConf)

	opts.SetClientID(CLIENT_ID)
	opts.SetAutoReconnect(true)
	opts.SetKeepAlive(10 * time.Second)
	opts.SetMaxReconnectInterval(5 * time.Minute)

	handler := MqttHandler{}
	opts.SetOnConnectHandler(handler.onConnect)
	opts.SetConnectionLostHandler(handler.onConnectionLost)

	client := mqtt.NewClient(opts)
	if tok := client.Connect(); tok.WaitTimeout(5*time.Second) && tok.Error() != nil {
		logger.WithError(tok.Error()).Fatal("Could not connect to mqtt server.")
	}

	return &handler
}

func (h *MqttHandler) onConnect(client mqtt.Client) {
	logger.Info("connected")

	err := subscribe(client, "/net/wlan-sessions",
		func(client mqtt.Client, message mqtt.Message) {
			//logger.Debug("new wifi sessions")
			mock := `{  "38134": {
    "last-auth": 1509211121,
    "vlan": "default",
    "stats": {
      "rx-multicast-pkts": 0,
      "rx-unicast-pkts": 292,
      "tx-unicast-pkts": 654,
      "rx-unicast-bytes": 20510,
      "tx-unicast-bytes": 278565,
      "rx-multicast-bytes": 0
    },
    "ssid": "mainframe",
    "ip": "::1",
    "hostname": "-",
    "last-snr": 47,
    "last-rate-mbits": "6",
    "ap": 1,
    "mac": "10:68:3f:bb:bb:bb",
    "radio": 2,
    "userinfo": {
      "name": "Holger",
      "visibility": "show",
      "ts": 1427737817755
    },
    "session-start": 1509211121,
    "last-rssi-dbm": -48,
    "last-activity": 1509211584
  }}`
			// message.Payload()
			h.WifiSessionList = parseWifiSessions([]byte(mock))
		})
	if err != nil {
		logger.WithError(err).Fatal("Could not subscribe")
	}
}

func (h *MqttHandler) onConnectionLost(client mqtt.Client, err error) {
	logger.WithError(err).Error("Connection lost.")
}

func (h *MqttHandler) GetByIp(ip string) (WifiSession, bool) {
	for _, v := range h.WifiSessionList {
		if v.Ip == ip {
			return v, true
		}
	}

	return WifiSession{}, false
}

func subscribe(client mqtt.Client, topic string, cb mqtt.MessageHandler) error {
	qos := 0
	tok := client.Subscribe(topic, byte(qos), cb)
	tok.WaitTimeout(5 * time.Second)
	return tok.Error()
}

func defaultCertPool(certFile string) *x509.CertPool {
	if certFile == "" {
		log.Debug("No certFile given, using system pool")
		pool, err := x509.SystemCertPool()
		if err != nil {
			log.WithError(err).Fatal("Could not create system cert pool.")
		}
		return pool
	}

	fileData, err := ioutil.ReadFile(certFile)
	if err != nil {
		log.WithError(err).Fatal("Could not read given cert file.")
	}

	certs := x509.NewCertPool()
	if !certs.AppendCertsFromPEM(fileData) {
		log.Fatal("unable to add given certificate to CertPool")
	}

	return certs
}

func parseWifiSessions(rawData []byte) []WifiSession {
	sessionsList := []WifiSession{}

	// we don't use a struct here, because we are interested in a small subset, only.
	var sessionData map[string]interface{}
	if err := json.Unmarshal(rawData, &sessionData); err != nil {
		log.WithFields(log.Fields{
			"rawData": string(rawData),
			"error":   err,
		}).Error("Unable to unmarshal wifi session json.")
		return sessionsList
	}

	for _, v := range sessionData {
		entry, ok := v.(map[string]interface{})
		if !ok {
			log.WithFields(log.Fields{
				"data": v,
			}).Error("Unable to unmarshal wifi session json. Unexpected structure")
			return sessionsList
		}

		vlan, ok := entry["vlan"].(string)
		if !ok {
			logParseError("vlan", v)
			return []WifiSession{}
		}
		ip, ok := entry["ip"].(string)
		if !ok {
			logParseError("ip", v)
			return []WifiSession{}
		}
		ap, ok := entry["ap"].(float64)
		if !ok {
			logParseError("ap", v)
			return []WifiSession{}
		}
		mac, ok := entry["mac"].(string)
		if !ok {
			logParseError("mac", v)
			return []WifiSession{}
		}
		sessionsList = append(sessionsList, WifiSession{ip, mac, vlan, int(ap)})
	}

	return sessionsList
}

func logParseError(field string, data interface{}) {
	logger.WithFields(log.Fields{
		"field": field,
		"data":  data,
	}).Error("Parse error for field.")
}
