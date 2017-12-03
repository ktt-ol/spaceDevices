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

	userDb := db.NewUserDb(config.MacDb)
	masterDb := db.NewMasterDb(config.MacDb)

	mqttHandler := mqtt.NewMqttHandler(config.Mqtt)
	data := mqtt.NewDeviceData(mqttHandler, masterDb, userDb)

	webService.StartWebService(config.Server, data, userDb)
}
