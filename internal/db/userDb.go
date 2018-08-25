package db

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	"github.com/ktt-ol/spaceDevices/internal/conf"
	log "github.com/sirupsen/logrus"
)

type UserDb interface {
	Get(mac string) (UserDbEntry, bool)
	Set(mac string, info UserDbEntry)
	Delete(mac string)
}

type UserDbEntry struct {
	Name       string     `json:"name"`
	DeviceName string     `json:"device-name"`
	Visibility Visibility `json:"visibility"`
	// last change in ms
	Ts int64 `json:"ts"`
}

type PersistentUserDb struct {
	userMap map[string]UserDbEntry
	lock    sync.RWMutex
	config  conf.MacDbConf
}

func NewUserDb(config conf.MacDbConf) UserDb {
	instance := &PersistentUserDb{config: config}
	instance.loadDb()
	return instance
}

func (db *PersistentUserDb) Get(mac string) (UserDbEntry, bool) {
	db.lock.RLock()
	value, ok := db.userMap[mac]
	db.lock.RUnlock()
	return value, ok
}

func (db *PersistentUserDb) Set(mac string, info UserDbEntry) {
	db.lock.Lock()
	defer db.lock.Unlock()
	db.userMap[mac] = info
	db.saveDb()
}

func (db *PersistentUserDb) Delete(mac string) {
	db.lock.Lock()
	defer db.lock.Unlock()
	delete(db.userMap, mac)
	db.saveDb()
}

func (db *PersistentUserDb) loadDb() {
	db.lock.Lock()
	defer db.lock.Unlock()

	file, err := ioutil.ReadFile(db.config.UserFile)
	if err != nil {
		log.Fatal("UserFile error: ", err)
	}

	var parsed map[string]UserDbEntry
	if err = json.Unmarshal(file, &parsed); err != nil {
		log.Fatal("UserFile unmarshal err: ", err)
	}

	db.userMap = parsed
}

func (db *PersistentUserDb) saveDb() {
	bytes, err := json.MarshalIndent(db.userMap, "", "  ")
	if err != nil {
		log.Fatal("Can't marshal the userDb: ", err)
	}

	if err = ioutil.WriteFile(db.config.UserFile, bytes, 0644); err != nil {
		log.Fatal("Can't save the userDb: ", err)
	}
}
