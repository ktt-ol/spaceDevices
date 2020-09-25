package mqtt

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ktt-ol/spaceDevices/internal/conf"

	"github.com/ktt-ol/spaceDevices/internal/db"
	"github.com/ktt-ol/spaceDevices/pkg/structs"
	"github.com/stretchr/testify/assert"
)

func Test_parseWifiSessions(t *testing.T) {
	const testData = `
	[
	{"ipv4": "192.168.2.127", "ipv6": [], "mac": "2c:0e:3d:aa:aa:aa", "ap": 2, "location": "Space"},
	{"ipv4": "192.168.2.179", "ipv6": ["6e7b:c7c6:9517:a9d0:958c:3939:c93e:9864", "e759:68b6:4c7d:8483:81b7:be87:119b:7ee1"], "mac": "10:68:3f:bb:bb:bb", "ap": 1, "location": "Radstelle"},
	{"ipv4": "192.168.2.35", "ipv6": [ "325c:7fa7:cc79:bcb7:a2b1:26f6:a4ef:2501" ], "mac": "20:c9:d0:cc:cc:cc", "ap": 1, "location": "Space"},
	{"ipv4": "10.18.159.6", "ipv6": [ ], "mac": "b8:53:ac:dd:dd:dd", "ap": 1, "location": ""}
	]
	`

	assert := assert.New(t)

	masterDb := &masterDbTest{}
	userDb := &userDbTest{}
	dd := DeviceData{masterDb: masterDb, userDb: userDb}

	sessions, _, ok := dd.parseWifiSessions([]byte(testData))
	assert.Equal(4, len(sessions))
	assert.True(ok)

	v := findByIp(assert, sessions, "192.168.2.127")
	assert.True(v.Mac == "2c:0e:3d:aa:aa:aa" && v.AP == 2 && v.Location == "Space")
	assert.Equal(0, len(v.Ipv6))

	v = findByIp(assert, sessions, "192.168.2.179")
	assert.True(v.Mac == "10:68:3f:bb:bb:bb"  && v.AP == 1 && v.Location == "Radstelle")
	assert.Equal(2, len(v.Ipv6))
	assert.Equal("6e7b:c7c6:9517:a9d0:958c:3939:c93e:9864", v.Ipv6[0])
	assert.Equal("e759:68b6:4c7d:8483:81b7:be87:119b:7ee1", v.Ipv6[1])

	v = findByIp(assert, sessions, "192.168.2.35")
	assert.True(v.Mac == "20:c9:d0:cc:cc:cc"  && v.AP == 1)
	assert.Equal(1, len(v.Ipv6))
	assert.Equal("325c:7fa7:cc79:bcb7:a2b1:26f6:a4ef:2501", v.Ipv6[0])

	v = findByIp(assert, sessions, "10.18.159.6")
	assert.True(v.Mac == "b8:53:ac:dd:dd:dd"  && v.AP == 1 && v.Location == "")
	assert.Equal(0, len(v.Ipv6))

	// don't fail for garbage
	sessions, _, ok = dd.parseWifiSessions([]byte("{ totally invalid json }"))
	assert.False(ok)
	assert.Equal(len(sessions), 0)
}

func findByIp(assert *assert.Assertions, sessions []structs.WifiSession, ip string) *structs.WifiSession {
	for _, v := range sessions {
		if v.Ipv4 == ip {
			return &v
		}
	}

	assert.Fail("IP not found", ip)
	return nil
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

func assertPeopleAndDevices(assert *assert.Assertions, peopleArrayCount int, peopleCount uint16, deviceCount uint16, unknownDevicesCount uint16, test structs.PeopleAndDevices) {
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
	Location string `Json:"location"`
	Ipv4   string  `json:"ipv4"`
	Ipv6 []string `json:"ipv6"`
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
	return sessionTestType{"Space", "10.1.1." + lastIp, make([]string, 0, 0),1, "00:00:00:00:00:" + lastMac}
}

func newSessionTestData(testData ...sessionTestType) []byte {
	sessionData := make([]sessionTestType, 0)

	for _, val := range testData {
		sessionData = append(sessionData, val)
	}

	bytes, err := json.Marshal(sessionData)
	if err != nil {
		panic(err)
	}

	return bytes
}
