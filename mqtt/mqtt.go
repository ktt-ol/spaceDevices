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

var mqttLogger = log.WithField("where", "mqtt")

type MqttHandler struct {
	client       mqtt.Client
	newDataChan  chan []byte
	sessionTopic string
	devicesTopic string
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
	opts.SetWill(conf.DevicesTopic, emptyPeopleAndDevices(), 0, true)

	handler := MqttHandler{newDataChan: make(chan []byte), devicesTopic: conf.DevicesTopic, sessionTopic: conf.SessionTopic}
	opts.SetOnConnectHandler(handler.onConnect)
	opts.SetConnectionLostHandler(handler.onConnectionLost)

	handler.client = mqtt.NewClient(opts)
	if tok := handler.client.Connect(); tok.WaitTimeout(5*time.Second) && tok.Error() != nil {
		mqttLogger.WithError(tok.Error()).Fatal("Could not connect to mqtt server.")
	}

	return &handler
}

func (h *MqttHandler) GetNewDataChannel() chan []byte {
	return h.newDataChan
}

func (h *MqttHandler) SendPeopleAndDevices(data PeopleAndDevices) {
	bytes, err := json.Marshal(data)
	if err != nil {
		mqttLogger.Errorln("Invalid people json", err)
		return
	}

	mqttLogger.Infof("Sending PeopleAndDevices: %d, %d, %d, %d",
		data.PeopleCount, data.DeviceCount, data.UnknownDevicesCount, len(data.People))

	token := h.client.Publish(h.devicesTopic, 0, true, string(bytes))
	ok := token.WaitTimeout(time.Duration(time.Second * 10))
	if !ok {
		mqttLogger.Warn("Error sending devices to:", h.devicesTopic)
		return
	}
}

func (h *MqttHandler) onConnect(client mqtt.Client) {
	mqttLogger.Info("connected")

	err := subscribe(client, h.sessionTopic,
		func(client mqtt.Client, message mqtt.Message) {
			mqttLogger.Debug("new wifi sessions")
			/*
							mock := []byte(`{  "38134": {
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
				    "mac": "d4:38:9c:01:dd:03",
				    "radio": 2,
				    "userinfo": {
				      "name": "Holger",
				      "visibility": "show",
				      "ts": 1427737817755
				    },
				    "session-start": 1509211121,
				    "last-rssi-dbm": -48,
				    "last-activity": 1509211584
				  }}`)
			*/
			select {
			//case h.newDataChan <- mock:
			case h.newDataChan <- message.Payload():
				break
			default:
				mqttLogger.Println("No one receives the message.")
			}

		})
	if err != nil {
		mqttLogger.WithError(err).Fatal("Could not subscribe")
	}
}

func (h *MqttHandler) onConnectionLost(client mqtt.Client, err error) {
	mqttLogger.WithError(err).Error("Connection lost.")
}

func subscribe(client mqtt.Client, topic string, cb mqtt.MessageHandler) error {
	qos := 0
	tok := client.Subscribe(topic, byte(qos), cb)
	tok.WaitTimeout(5 * time.Second)
	return tok.Error()
}

func defaultCertPool(certFile string) *x509.CertPool {
	if certFile == "" {
		mqttLogger.Debug("No certFile given, using system pool")
		pool, err := x509.SystemCertPool()
		if err != nil {
			mqttLogger.WithError(err).Fatal("Could not create system cert pool.")
		}
		return pool
	}

	fileData, err := ioutil.ReadFile(certFile)
	if err != nil {
		mqttLogger.WithError(err).Fatal("Could not read given cert file.")
	}

	certs := x509.NewCertPool()
	if !certs.AppendCertsFromPEM(fileData) {
		mqttLogger.Fatal("unable to add given certificate to CertPool")
	}

	return certs
}

func emptyPeopleAndDevices() string {
	pad := PeopleAndDevices{People: []Person{}}
	bytes, err := json.Marshal(pad)
	if err != nil {
		mqttLogger.WithError(err).Panic()
	}
	return string(bytes)
}
