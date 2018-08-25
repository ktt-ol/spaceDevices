package structs

import (
	"strings"
)

type WifiSession struct {
	Ip   string
	Mac  string
	Vlan string
	AP   int
}

type Devices struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type DevicesSorter []Devices

func (s DevicesSorter) Len() int {
	return len(s)
}
func (s DevicesSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s DevicesSorter) Less(i, j int) bool {
	return strings.Compare(s[i].Name, s[j].Name) < 0
}

type Person struct {
	Name    string    `json:"name"`
	Devices []Devices `json:"devices"`
}

type PersonSorter []Person

func (s PersonSorter) Len() int {
	return len(s)
}
func (s PersonSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s PersonSorter) Less(i, j int) bool {
	return strings.Compare(s[i].Name, s[j].Name) < 0
}

type PeopleAndDevices struct {
	People              []Person `json:"people"`
	PeopleCount         uint16   `json:"peopleCount"`
	DeviceCount         uint16   `json:"deviceCount"`
	UnknownDevicesCount uint16   `json:"unknownDevicesCount"`
}
