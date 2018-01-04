package conf_test

import (
	"testing"

	"github.com/ktt-ol/spaceDevices/conf"
	"github.com/stretchr/testify/assert"
)

func Test_exampleConfig(t *testing.T) {
	assert := assert.New(t)

	config := conf.LoadConfig("../config.example.toml")
	assert.Equal(false, config.Misc.DebugLogging)
	assert.Equal("", config.Misc.Logfile)
	assert.Equal(2, len(config.Locations))
	assert.Equal("Bar", config.Locations[0].Name)
	assert.Equal(2, len(config.Locations[0].Ids))
}
