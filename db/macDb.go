package db

import (
	"encoding/json"
	"io/ioutil"

	"github.com/ktt-ol/spaceDevices/conf"
	log "github.com/sirupsen/logrus"
)

type MasterDb interface {
	Get(mac string) (MasterDbEntry, bool)
}

type MasterDbEntry struct {
	UserDbEntry
	DeviceType                string `json:"device-type"`
	PoweredWhileClosedWarning bool   `json:"powered-while-closed-warning"`
}

type fileMasterDb struct {
	masterMap map[string]MasterDbEntry
}

func NewMasterDb(config conf.MacDbConf) MasterDb {
	instance := &fileMasterDb{}
	instance.loadDb(config.MasterFile)
	return instance
}

func (db *fileMasterDb) Get(mac string) (MasterDbEntry, bool) {
	value, ok := db.masterMap[mac]
	return value, ok
}

func (db *fileMasterDb) loadDb(masterFile string) {
	file, err := ioutil.ReadFile(masterFile)
	if err != nil {
		log.Fatal("MasterFile error: ", err)
	}

	var parsed map[string]MasterDbEntry
	if err = json.Unmarshal(file, &parsed); err != nil {
		log.Fatal("MasterFile unmarshal err: ", err)
	}

	db.masterMap = parsed
}
