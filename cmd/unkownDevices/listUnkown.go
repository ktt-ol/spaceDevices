package main

import (
	"bufio"
	"fmt"
	"github.com/ktt-ol/spaceDevices/internal/conf"
	"github.com/ktt-ol/spaceDevices/internal/db"
	"github.com/ktt-ol/spaceDevices/internal/mqtt"
	"os"
	"strings"
)

const CONFIG_FILE = "config.toml"

const (
	InfoColor    = "\033[1;34m%s\033[0m"
	NoticeColor  = "\033[1;36m%s\033[0m"
	WarningColor = "\033[1;33m%s\033[0m"
	ErrorColor   = "\033[1;31m%s\033[0m"
	DebugColor   = "\033[0;36m%s\033[0m"
	PrintColor = "\033[38;5;%dm%s\033[39;49m\n"
)

func main() {
	config := conf.LoadConfig(CONFIG_FILE)

	userDb := db.NewUserDb(config.MacDb)
	masterDb := db.NewMasterDb(config.MacDb)

	mqttHandler := mqtt.NewMqttHandler(config.Mqtt, true)
	data := mqtt.NewDeviceData(config.Locations, mqttHandler, masterDb, userDb)
	unknownSession := data.GetOneEntry()

	macDb := loadMacDb()
	for _, s := range unknownSession {
		// 5c:51:4f
		firstHexBytes := strings.ToUpper(strings.ReplaceAll(s.Mac[0:8], ":", ""))
		name, ok := macDb[firstHexBytes]
		if !ok {
			name = "Unknown"
		}
		fmt.Printf("%s %s\n", fmt.Sprintf(InfoColor, s.Mac), name)
		fmt.Printf("-> %s // %s\n", s.Ipv4, s.Ipv6)
		jsonEntry := fmt.Sprintf(`"%s":{"name": "%s", "device-type": "", "visibility": "ignore"},`, s.Mac, name)
		fmt.Println(jsonEntry)
		fmt.Println("")
	}
}

func loadMacDb() map[string]string {
	file, err := os.Open("macVendorDb.csv")
	check(err)
	defer file.Close()

	macDb := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		assignment := line[0:6]
		name := line[7:]
		macDb[assignment] = name
	}

	check(scanner.Err())

	return macDb
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
