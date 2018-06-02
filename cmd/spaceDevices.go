package main

import (
	"fmt"
	"os"

	"github.com/ktt-ol/spaceDevices/conf"
	"github.com/ktt-ol/spaceDevices/db"
	"github.com/ktt-ol/spaceDevices/mqtt"
	"github.com/ktt-ol/spaceDevices/webService"
	"github.com/sirupsen/logrus"
)

const CONFIG_FILE = "config.toml"

func main() {
	config := conf.LoadConfig(CONFIG_FILE)

	setupLogging(config.Misc)

	logrus.WithFields(logrus.Fields{
		"session": config.Mqtt.SessionTopic,
		"devices": config.Mqtt.DevicesTopic,
		"mqttUser": config.Mqtt.Username,
		"master": config.MacDb.MasterFile,
		"user": config.MacDb.UserFile,
	}).Info("SpaceDevices starting...")

	//mqtt.EnableMqttDebugLogging()

	userDb := db.NewUserDb(config.MacDb)
	masterDb := db.NewMasterDb(config.MacDb)

	mqttHandler := mqtt.NewMqttHandler(config.Mqtt)
	data := mqtt.NewDeviceData(config.Locations, mqttHandler, masterDb, userDb)

	webService.StartWebService(config.Server, data, userDb)
}

type StdErrLogHook struct {
}

func (h *StdErrLogHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
}
func (h *StdErrLogHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read entry, %v", err)
		return err
	}
	fmt.Fprintf(os.Stderr, line)
	return nil
}

func setupLogging(config conf.MiscConf) {
	logrus.SetFormatter(&logrus.TextFormatter{DisableColors: true})
	if config.DebugLogging {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	if config.Logfile == "" {
		logrus.SetOutput(os.Stdout)
	} else {
		// https://github.com/sirupsen/logrus/issues/227
		file, err := os.OpenFile(config.Logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err == nil {
			logrus.SetOutput(file)
		} else {
			logrus.Warnf("Failed to log to file '%s', using default stderr.", config.Logfile)
		}
		logrus.AddHook(&StdErrLogHook{})
	}
}
