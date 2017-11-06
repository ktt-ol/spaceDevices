package main

import (
	"github.com/ktt-ol/spaceDevices/conf"
	"github.com/ktt-ol/spaceDevices/db"
	"github.com/ktt-ol/spaceDevices/mqtt"
	"github.com/ktt-ol/spaceDevices/webService"
	log "github.com/sirupsen/logrus"
)

const CONFIG_FILE = "config.toml"

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{DisableColors: true})

	config := conf.LoadConfig(CONFIG_FILE)

	//spaceDevices.EnableMqttDebugLogging()
	mqttHandler := mqtt.NewMqttHandler(config.Mqtt)
	macDb := db.NewUserMacSettings(config.MacDb)
	webService.StartWebService(config.Server, mqttHandler, macDb)
}
