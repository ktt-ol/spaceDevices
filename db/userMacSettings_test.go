package db

import "testing"
import (
	"github.com/stretchr/testify/assert"
)

func Test_IsMacLocallyAdministered(t *testing.T) {
	assert.True(t, IsMacLocallyAdministered("06:00:00:00:00:00"))
	assert.True(t, IsMacLocallyAdministered("62:01:0f:b5:f2:d9"))
	assert.False(t, IsMacLocallyAdministered("20:c9:d0:7a:fa:31"))
}

