package mqtt

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

type Person struct {
	Name    string    `json:"name"`
	Devices []Devices `json:"devices"`
}

type PeopleAndDevices struct {
	People              []Person `json:"people"`
	PeopleCount         uint     `json:"peopleCount"`
	DeviceCount         uint     `json:"deviceCount"`
	UnknownDevicesCount uint     `json:"unknownDevicesCount"`
}
