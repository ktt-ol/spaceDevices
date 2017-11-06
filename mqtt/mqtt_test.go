package mqtt

import "testing"
import (
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

	sessions := parseWifiSessions([]byte(testData))
	assert.Equal(len(sessions), 4)

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
	sessions = parseWifiSessions([]byte("{ totally invalid json }"))
	assert.Equal(len(sessions), 0)
}
