package mqtt

import (
	"encoding/json"
	"fmt"

	"github.com/ktt-ol/spaceDevices/conf"
	"github.com/ktt-ol/spaceDevices/db"

	"github.com/sirupsen/logrus"
)

var ignoredVisibility = [...]db.Visibility{db.VisibilityCriticalInfrastructure, db.VisibilityImportantInfrastructure,
	db.VisibilityInfrastructure, db.VisibilityDeprecatedInfrastructure, db.VisibilityUserInfrastructure}

var ddLogger = logrus.WithField("where", "deviceData")

type devicesEntry struct {
	hideName    bool
	showDevices bool
	devices     []Devices
}

type DeviceData struct {
	locations       []conf.Location
	mqttHandler     *MqttHandler
	masterDb        db.MasterDb
	userDb          db.UserDb
	wifiSessionList []WifiSession
	// more to come, e.g. LanSessions
}

func NewDeviceData(locations []conf.Location, mqttHandler *MqttHandler, masterDb db.MasterDb, userDb db.UserDb) *DeviceData {
	dd := DeviceData{locations: locations, mqttHandler: mqttHandler, masterDb: masterDb, userDb: userDb}

	go func() {
		for {
			data := <-mqttHandler.GetNewDataChannel()
			dd.newData(data)
		}
	}()

	return &dd
}

func (d *DeviceData) newData(data []byte) {
	fmt.Println("got data, len ", len(data))
	sessionsList, peopleAndDevices, ok := d.parseWifiSessions(data)
	if ok {
		d.wifiSessionList = sessionsList
		fmt.Printf("peopleAndDevices: %+v\n", peopleAndDevices)
		d.mqttHandler.SendPeopleAndDevices(peopleAndDevices)
	}
}

func (d *DeviceData) GetByIp(ip string) (WifiSession, bool) {
	for _, v := range d.wifiSessionList {
		if v.Ip == ip {
			return v, true
		}
	}

	return WifiSession{}, false
}

func (d *DeviceData) parseWifiSessions(rawData []byte) (sessionsList []WifiSession, peopleAndDevices PeopleAndDevices, success bool) {
	var sessionData map[string]WifiSession
	if err := json.Unmarshal(rawData, &sessionData); err != nil {
		ddLogger.WithFields(logrus.Fields{
			"rawData": string(rawData),
			"error":   err,
		}).Error("Unable to unmarshal wifi session json.")
		return
	}

	username2DevicesMap := make(map[string]*devicesEntry)
SESSION_LOOP:
	for _, wifiSession := range sessionData {
		sessionsList = append(sessionsList, wifiSession)

		peopleAndDevices.DeviceCount++
		var userInfo db.UserDbEntry
		masterDbEntry, ok := d.masterDb.Get(wifiSession.Mac)
		if ok {
			userInfo = masterDbEntry.UserDbEntry
		} else {
			userInfo, ok = d.userDb.Get(wifiSession.Mac)
			if !ok {
				// nothing found for this mac
				peopleAndDevices.UnknownDevicesCount++
				continue
			}
		}
		for _, v := range ignoredVisibility {
			if v == userInfo.Visibility {
				continue SESSION_LOOP
			}
		}

		entry, ok := username2DevicesMap[userInfo.Name]
		if !ok {
			entry = &devicesEntry{}
			username2DevicesMap[userInfo.Name] = entry
		}

		device := Devices{Name: userInfo.DeviceName, Location: d.findLocation(wifiSession.AP)}
		entry.devices = append(entry.devices, device)

		if userInfo.Visibility == db.VisibilityIgnore {
			entry.hideName = true
			continue
		}

		if len(entry.devices) == 1 {
			peopleAndDevices.PeopleCount++
		}

		if userInfo.Visibility == db.VisibilityAnon {
			entry.hideName = true
			continue
		}

		if userInfo.Visibility == db.VisibilityUser {
			entry.showDevices = false
			continue
		}

		if userInfo.Visibility == db.VisibilityAll {
			entry.showDevices = true
			continue
		}
	}

	for username, devicesEntry := range username2DevicesMap {
		if devicesEntry.hideName {
			continue
		}

		var person Person
		if devicesEntry.showDevices {
			person = Person{Name: username, Devices: devicesEntry.devices}
		} else {
			person = Person{Name: username}
		}
		peopleAndDevices.People = append(peopleAndDevices.People, person)
	}

	success = true
	return
}

func (d *DeviceData) findLocation(apID int) string {

	for _, location := range d.locations {
		for _, id := range location.Ids {
			if id == apID {
				return location.Name
			}
		}
	}
	return ""
}

func logParseError(field string, data interface{}) {
	ddLogger.WithFields(logrus.Fields{
		"field": field,
		"data":  data,
	}).Error("Parse error for field.")
}
