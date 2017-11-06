package db

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"sync"

	"github.com/ktt-ol/spaceDevices/conf"
	log "github.com/sirupsen/logrus"
)

type UserMacSettings struct {
	userMap map[string]UserMacInfo
	lock    sync.RWMutex
	config  conf.MacDbConf
}

type Visibility uint8

const (
	// Not shown at all
	VisibilityIgnore Visibility = iota
	// don't show the name, but increments the anonymous user count
	VisibilityAnon
	// show the user, but not the device name(s)
	VisibilityUser
	// show user and the device names
	VisibilityAll
)

func ParseVisibility(visibility uint8) (Visibility, bool) {
	if visibility > 3 {
		return 0, false
	}

	return Visibility(visibility), true
}

func NewUserMacSettings(config conf.MacDbConf) *UserMacSettings {
	instance := &UserMacSettings{config: config}
	instance.loadDb()
	return instance
}

func (db *UserMacSettings) Get(mac string) (UserMacInfo, bool) {
	db.lock.RLock()
	value, ok := db.userMap[mac]
	db.lock.RUnlock()
	return value, ok
}

func (db *UserMacSettings) Set(mac string, info UserMacInfo) {
	db.lock.Lock()
	defer db.lock.Unlock()
	db.userMap[mac] = info
	db.saveDb()
}

func (db *UserMacSettings) Delete(mac string) {
	db.lock.Lock()
	defer db.lock.Unlock()
	delete(db.userMap, mac)
}

func (db *UserMacSettings) loadDb() {
	db.lock.Lock()
	defer db.lock.Unlock()

	file, err := ioutil.ReadFile(db.config.UserFile)
	if err != nil {
		log.Fatal("UserFile error: ", err)
	}

	var parsed entryMap
	if err = json.Unmarshal(file, &parsed); err != nil {
		log.Fatal("Unmarshal err: ", err)
	}

	db.userMap = parsed
}

func (db *UserMacSettings) saveDb() {
	bytes, err := json.MarshalIndent(db.userMap, "", "  ")
	if err != nil {
		log.Fatal("Can't marshal the userDb: ", err)
	}

	if err = ioutil.WriteFile(db.config.UserFile, bytes, 0644); err != nil {
		log.Fatal("Can't save the userDb: ", err)
	}
}

type entryMap map[string]UserMacInfo

type UserMacInfo struct {
	Name       string
	DeviceName string
	Visibility Visibility
	// last change in ms
	Ts int64
}

// IsMacLocallyAdministered expects the mac in the format e.g. "20:c9:d0:7a:fa:31"
// https://en.wikipedia.org/wiki/MAC_address
func IsMacLocallyAdministered(mac string) bool {
	// 00000010
	const mask = 1 << 1

	first2chars := mac[:2]
	decimal, _ := strconv.ParseInt(first2chars, 16, 8)
	return (decimal & mask) == mask
}
