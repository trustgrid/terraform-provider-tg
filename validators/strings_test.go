package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsHostname_HappyPath(t *testing.T) {
	valid := []string{"trustgrid.io", "somecluster.whatever.trustgrid.io"}

	for _, v := range valid {
		warnings, errors := IsHostname(v, "hostname")
		assert.Empty(t, warnings)
		assert.Empty(t, errors)
	}
}

func Test_IsHostname_Invalid(t *testing.T) {
	invalid := []string{"trustgrid", "trustgrid.", "trustgrid", "trustgrid..io", "trustgrid.io.", "1.1.1.", "2.2.2.2"}

	for _, v := range invalid {
		_, errors := IsHostname(v, "hostname")
		assert.NotEmpty(t, errors, "%s should not be valid", v)
	}
}
