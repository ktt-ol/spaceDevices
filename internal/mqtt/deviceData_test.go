package mqtt

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/ktt-ol/spaceDevices/internal/conf"

	"github.com/ktt-ol/spaceDevices/internal/db"
	"github.com/stretchr/testify/assert"
)

func Test_parseWifiSessions(t *testing.T) {
	const testData = `
{
  "38126": {
    "last-auth": 1509210709,
    "vlan": "default",
    "stats": {
      "rx-multicast-pkts": 499,
      "rx-unicast-pkts": 1817,
      "tx-unicast-pkts": 734,
      "rx-unicast-bytes": 156208,
      "tx-unicast-bytes": 272461,
      "rx-multicast-bytes": 76808
    },
    "ssid": "mainframe",
    "ip": "192.168.2.127",
    "hostname": "-",
    "last-snr": 40,
    "last-rate-mbits": "24",
    "ap": 2,
    "mac": "2c:0e:3d:aa:aa:aa",
    "radio": 2,
    "userinfo": null,
    "session-start": 1509210709,
    "last-rssi-dbm": -55,
    "last-activity": 1509211581
  },
  "38134": {
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
    "ip": "192.168.2.179",
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
  },
  "38135": {
    "last-auth": 1509211163,
    "vlan": "default",
    "stats": {
      "rx-multicast-pkts": 114,
      "rx-unicast-pkts": 8119,
      "tx-unicast-pkts": 12440,
      "rx-unicast-bytes": 1093407,
      "tx-unicast-bytes": 15083985,
      "rx-multicast-bytes": 20379
    },
    "ssid": "mainframe",
    "ip": "192.168.2.35",
    "hostname": "happle",
    "last-snr": 39,
    "last-rate-mbits": "24",
    "ap": 1,
    "mac": "20:c9:d0:cc:cc:cc",
    "radio": 2,
    "userinfo": {
      "name": "Holger",
      "visibility": "show",
      "ts": 1438474581580
    },
    "session-start": 1509211163,
    "last-rssi-dbm": -56,
    "last-activity": 1509211584
  },
  "38137": {
    "last-auth": 1509211199,
    "vlan": "FreiFunk",
    "stats": {
      "rx-multicast-pkts": 14,
      "rx-unicast-pkts": 931,
      "tx-unicast-pkts": 615,
      "rx-unicast-bytes": 70172,
      "tx-unicast-bytes": 265390,
      "rx-multicast-bytes": 1574
    },
    "ssid": "nordwest.freifunk.net",
    "ip": "10.18.159.6",
    "hostname": "iPhonevineSager",
    "last-snr": 13,
    "last-rate-mbits": "2",
    "ap": 1,
    "mac": "b8:53:ac:dd:dd:dd",
    "radio": 1,
    "userinfo": null,
    "session-start": 1509211199,
    "last-rssi-dbm": -82,
    "last-activity": 1509211584
  }
}
	`
	assert := assert.New(t)

	masterDb := &masterDbTest{}
	userDb := &userDbTest{}
	dd := DeviceData{masterDb: masterDb, userDb: userDb}

	sessions, _, ok := dd.parseWifiSessions([]byte(testData))
	assert.Equal(len(sessions), 4)
	assert.True(ok)

	mustContain := [4]bool{false, false, false, false}
	for _, v := range sessions {
		mustContain[0] = mustContain[0] || (v.Ip == "192.168.2.127" && v.Mac == "2c:0e:3d:aa:aa:aa" && v.Vlan == "default" && v.AP == 2)
		mustContain[1] = mustContain[1] || (v.Ip == "192.168.2.179" && v.Mac == "10:68:3f:bb:bb:bb" && v.Vlan == "default" && v.AP == 1)
		mustContain[2] = mustContain[2] || (v.Ip == "192.168.2.35" && v.Mac == "20:c9:d0:cc:cc:cc" && v.Vlan == "default" && v.AP == 1)
		mustContain[3] = mustContain[3] || (v.Ip == "10.18.159.6" && v.Mac == "b8:53:ac:dd:dd:dd" && v.Vlan == "FreiFunk" && v.AP == 1)
	}
	for _, v := range mustContain {
		assert.True(v)
	}

	// don't fail for garbage
	sessions, _, ok = dd.parseWifiSessions([]byte("{ totally invalid json }"))
	assert.False(ok)
	assert.Equal(len(sessions), 0)
}

func Test_peopleCalculation(t *testing.T) {
	assert := assert.New(t)
	masterMap := make(map[string]db.MasterDbEntry)
	masterDb := &masterDbTest{masterMap: masterMap}
	userMap := make(map[string]db.UserDbEntry)
	userDb := &userDbTest{userMap}
	locations := []conf.Location{conf.Location{Name: "Bar", Ids: []int{1, 3}}}
	dd := DeviceData{locations: locations, masterDb: masterDb, userDb: userDb}

	testData := newSessionTestData(stt("1", "01"), stt("2", "02"), stt("3", "03"), stt("4", "04"), stt("5", "05"))
	_, peopleAndDevices, _ := dd.parseWifiSessions(testData)
	assertPeopleAndDevices(assert, 0, 0, 5, 5, peopleAndDevices)

	userMap["00:00:00:00:00:01"] = db.UserDbEntry{Name: "holger", DeviceName: "handy", Visibility: db.VisibilityUser}
	_, peopleAndDevices, _ = dd.parseWifiSessions(testData)
	assertPeopleAndDevices(assert, 1, 1, 5, 4, peopleAndDevices)

	userMap["00:00:00:00:00:02"] = db.UserDbEntry{Name: "hans", DeviceName: "", Visibility: db.VisibilityAnon}
	_, peopleAndDevices, _ = dd.parseWifiSessions(testData)
	assertPeopleAndDevices(assert, 1, 2, 5, 3, peopleAndDevices)

	userMap["00:00:00:00:00:03"] = db.UserDbEntry{Name: "herman", DeviceName: "", Visibility: db.VisibilityIgnore}
	_, peopleAndDevices, _ = dd.parseWifiSessions(testData)
	assertPeopleAndDevices(assert, 1, 2, 5, 2, peopleAndDevices)

	userMap["00:00:00:00:00:04"] = db.UserDbEntry{Name: "olaf", DeviceName: "iphone", Visibility: db.VisibilityAll}
	_, peopleAndDevices, _ = dd.parseWifiSessions(testData)
	assertPeopleAndDevices(assert, 2, 3, 5, 1, peopleAndDevices)
	for _, p := range peopleAndDevices.People {
		if p.Name == "olaf" {
			assert.Equal("iphone", p.Devices[0].Name)
			assert.Equal("Bar", p.Devices[0].Location)
		} else {
			assert.Equal(0, len(p.Devices))
		}
	}
	entry := db.MasterDbEntry{}
	entry.Name = "pc1"
	entry.Visibility = db.VisibilityCriticalInfrastructure
	masterMap["00:00:00:00:00:05"] = entry
	_, peopleAndDevices, _ = dd.parseWifiSessions(testData)
	assertPeopleAndDevices(assert, 2, 3, 5, 0, peopleAndDevices)

	// add a second device for olaf
	testData = newSessionTestData(stt("1", "01"), stt("2", "02"), stt("3", "03"), stt("4", "04"), stt("5", "05"), stt("6", "06"))
	userMap["00:00:00:00:00:06"] = db.UserDbEntry{Name: "olaf", DeviceName: "mac", Visibility: db.VisibilityAll}
	_, peopleAndDevices, _ = dd.parseWifiSessions(testData)
	fmt.Printf("peopleAndDevices: %+v\n", peopleAndDevices)
	assertPeopleAndDevices(assert, 2, 3, 6, 0, peopleAndDevices)
	for _, p := range peopleAndDevices.People {
		if p.Name == "olaf" {
			assert.Equal(2, len(p.Devices))
		} else {
			assert.Equal(0, len(p.Devices))
		}
	}
}

func assertPeopleAndDevices(assert *assert.Assertions, peopleArrayCount int, peopleCount uint, deviceCount uint, unknownDevicesCount uint, test PeopleAndDevices) {
	assert.Equal(peopleArrayCount, len(test.People), "len(People)")
	assert.Equal(peopleCount, test.PeopleCount, "peopleCount")
	assert.Equal(deviceCount, test.DeviceCount, "deviceCount")
	assert.Equal(unknownDevicesCount, test.UnknownDevicesCount, "unknownDevicesCount")
}

func Test_peopleNeverNil(t *testing.T) {
	assert := assert.New(t)
	dd := DeviceData{}

	testData := newSessionTestData()
	_, peopleAndDevices, success := dd.parseWifiSessions(testData)
	assert.True(success)
	assert.NotNil(peopleAndDevices.People)
}

/****************************************/
/* helper to create test data and mocks */
/****************************************/

type sessionTestType struct {
	Vlan string  `json:"vlan"`
	IP   string  `json:"ip"`
	Ap   float64 `json:"ap"`
	Mac  string  `json:"mac"`
}

type userDbTest struct {
	userMap map[string]db.UserDbEntry
}

func (db *userDbTest) Get(mac string) (db.UserDbEntry, bool) {
	value, ok := db.userMap[mac]
	return value, ok
}

func (db *userDbTest) Set(mac string, info db.UserDbEntry) {
	db.userMap[mac] = info
}

func (db *userDbTest) Delete(mac string) {
	delete(db.userMap, mac)
}

type masterDbTest struct {
	masterMap map[string]db.MasterDbEntry
}

func (db *masterDbTest) Get(mac string) (db.MasterDbEntry, bool) {
	value, ok := db.masterMap[mac]
	return value, ok
}

func stt(lastIp string, lastMac string) sessionTestType {
	return sessionTestType{"vlan", "10.1.1." + lastIp, 1, "00:00:00:00:00:" + lastMac}
}

func newSessionTestData(testData ...sessionTestType) []byte {
	sessionData := make(map[string]sessionTestType)

	for index, val := range testData {
		sessionData[strconv.Itoa(index)] = val
	}

	bytes, err := json.Marshal(sessionData)
	if err != nil {
		panic(err)
	}

	return bytes
}
