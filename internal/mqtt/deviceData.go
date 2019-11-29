package mqtt

import (
	"encoding/json"

	"github.com/ktt-ol/spaceDevices/internal/conf"
	"github.com/ktt-ol/spaceDevices/internal/db"

	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/ktt-ol/spaceDevices/pkg/structs"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
)

var ignoredVisibility = [...]db.Visibility{db.VisibilityCriticalInfrastructure, db.VisibilityImportantInfrastructure,
	db.VisibilityInfrastructure, db.VisibilityDeprecatedInfrastructure, db.VisibilityUserInfrastructure}

var ddLogger = logrus.WithField("where", "deviceData")

type devicesEntry struct {
	hideName    bool
	showDevices bool
	devices     []structs.Devices
}

type DeviceData struct {
	locations       []conf.Location
	mqttHandler     *MqttHandler
	masterDb        db.MasterDb
	userDb          db.UserDb
	wifiSessionList []structs.WifiSession

	lastSentHash []byte

	// more to come, e.g. LanSessions
}

func NewDeviceData(locations []conf.Location, mqttHandler *MqttHandler, masterDb db.MasterDb, userDb db.UserDb) *DeviceData {
	dd := DeviceData{locations: locations, mqttHandler: mqttHandler, masterDb: masterDb, userDb: userDb}
	return &dd
}

func (d *DeviceData) ListenAndUpdatePeopleData() {
	go func() {
		for {
			data := <-d.mqttHandler.GetNewDataChannel()
			d.newData(data)
		}
	}()
}

func (d *DeviceData) GetOneEntry() []structs.WifiSession {
	data := <-d.mqttHandler.GetNewDataChannel()
	sessionsList := d.unmarshal(data)
	if sessionsList == nil {
		return nil
	}

	unknownSession := make([]structs.WifiSession, 0, len(sessionsList))
	for _, wifiSession := range sessionsList {
		_, ok := d.masterDb.Get(wifiSession.Mac)
		if ok {
			continue
		}
		_, ok = d.userDb.Get(wifiSession.Mac)
		if ok {
			continue
		}

		unknownSession = append(unknownSession, wifiSession)
	}

	return unknownSession
}

func (d *DeviceData) newData(data []byte) {
	sessionsList, peopleAndDevices, ok := d.parseWifiSessions(data)
	if ok {
		d.wifiSessionList = sessionsList
		if ddLogger.Logger.Level >= logrus.DebugLevel {
			peopleList := make([]string, 0, len(peopleAndDevices.People))
			for _, person := range peopleAndDevices.People {
				if len(person.Name) == 0 {
					continue
				}
				personStr := person.Name + " ["
				for _, device := range person.Devices {
					personStr += device.Name + ","
				}
				personStr += "]"
				peopleList = append(peopleList, personStr)
			}
			sort.Strings(peopleList)
			ddLogger.Debugf("PeopleCount: %d, DeviceCount: %d, UnknownDevicesCount: %d, Persons: %s",
				peopleAndDevices.PeopleCount, peopleAndDevices.DeviceCount, peopleAndDevices.UnknownDevicesCount, strings.Join(peopleList, "; "))
		}
		h := md5.New()
		s := fmt.Sprintf("%v", peopleAndDevices)
		hash := h.Sum([]byte(s))
		if bytes.Equal(hash, d.lastSentHash) {
			ddLogger.Debug("Nothing changed in people count, skipping mqtt")
		} else {
			d.mqttHandler.SendPeopleAndDevices(peopleAndDevices)
			d.lastSentHash = hash
		}

	}
}

// finds the session entry for the given ip v4 or v6 address
func (d *DeviceData) GetByIp(ip string) (structs.WifiSession, bool) {
	if strings.Count(ip, ":") < 2 {
		// v4
		for _, v := range d.wifiSessionList {
			if v.Ip == ip {
				return v, true
			}
		}
	} else {
		// v6
		for _, v := range d.wifiSessionList {
			if len(v.Ipv6) > 0 {
				for _, v6 := range v.Ipv6 {
					if v6 == ip {
						return v, true
					}
				}
			}
		}
	}

	return structs.WifiSession{}, false
}

func (d *DeviceData) unmarshal(rawData []byte) map[string]structs.WifiSession {
	var sessionData map[string]structs.WifiSession
	if err := json.Unmarshal(rawData, &sessionData); err != nil {
		ddLogger.WithFields(logrus.Fields{
			"rawData": string(rawData),
			"error":   err,
		}).Error("Unable to unmarshal wifi session json.")
		return nil
	}

	return sessionData
}

func (d *DeviceData) parseWifiSessions(rawData []byte) (sessionsList []structs.WifiSession, peopleAndDevices structs.PeopleAndDevices, success bool) {
	sessionData := d.unmarshal(rawData)
	if sessionData == nil {
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

		device := structs.Devices{Name: userInfo.DeviceName, Location: d.findLocation(wifiSession.AP)}
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

	peopleAndDevices.People = make([]structs.Person, 0, 10)
	for username, devicesEntry := range username2DevicesMap {
		if devicesEntry.hideName {
			continue
		}

		var person structs.Person
		if devicesEntry.showDevices {
			person = structs.Person{Name: username, Devices: devicesEntry.devices}
			sort.Sort(structs.DevicesSorter(person.Devices))
		} else {
			person = structs.Person{Name: username}
		}
		peopleAndDevices.People = append(peopleAndDevices.People, person)
	}
	sort.Sort(structs.PersonSorter(peopleAndDevices.People))

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
