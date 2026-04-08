package datasource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func TestIfacesForDevice_ExplicitWAN(t *testing.T) {
	device := tg.Device{
		WAN: "ens4",
		LAN: []string{"ens5", "ens6"},
	}

	ifaces := ifacesForDevice(device)
	require.Len(t, ifaces, 3)

	assert.Equal(t, "ens4", ifaces[0].osName)
	assert.Equal(t, "NIC1", ifaces[0].name)
	assert.Equal(t, "WAN Interface", ifaces[0].description)

	assert.Equal(t, "ens5", ifaces[1].osName)
	assert.Equal(t, "NIC2", ifaces[1].name)
	assert.Equal(t, "LAN Interface", ifaces[1].description)

	assert.Equal(t, "ens6", ifaces[2].osName)
	assert.Equal(t, "NIC3", ifaces[2].name)
	assert.Equal(t, "Interface 3", ifaces[2].description)
}

func TestIfacesForDevice_ExplicitWANOnly(t *testing.T) {
	device := tg.Device{
		WAN: "eno1",
	}

	ifaces := ifacesForDevice(device)
	require.Len(t, ifaces, 1)
	assert.Equal(t, "eno1", ifaces[0].osName)
	assert.Equal(t, "NIC1", ifaces[0].name)
	assert.Equal(t, "WAN Interface", ifaces[0].description)
}

func TestIfacesForDevice_ExplicitLANOnly(t *testing.T) {
	// LAN present but WAN empty — edge case where WAN is temporarily missing.
	// LANs are still numbered from NIC2 (WAN slot reserved as NIC1).
	device := tg.Device{
		LAN: []string{"ens5", "ens6"},
	}

	ifaces := ifacesForDevice(device)
	require.Len(t, ifaces, 2)
	assert.Equal(t, "ens5", ifaces[0].osName)
	assert.Equal(t, "NIC2", ifaces[0].name)
	assert.Equal(t, "LAN Interface", ifaces[0].description)
	assert.Equal(t, "ens6", ifaces[1].osName)
	assert.Equal(t, "NIC3", ifaces[1].name)
	assert.Equal(t, "Interface 3", ifaces[1].description)
}

func TestIfacesForDevice_CatalogVendorModel(t *testing.T) {
	device := tg.Device{
		Vendor: "gcp",
	}

	ifaces := ifacesForDevice(device)
	require.Len(t, ifaces, 2)
	assert.Equal(t, "ens4", ifaces[0].osName)
	assert.Equal(t, "NIC1", ifaces[0].name)
	assert.Equal(t, "WAN Interface", ifaces[0].description)
	assert.Equal(t, "ens5", ifaces[1].osName)
}

func TestIfacesForDevice_CatalogVendorModelSpecific(t *testing.T) {
	device := tg.Device{
		Vendor: "lanner",
		Model:  "nca-1515",
	}

	ifaces := ifacesForDevice(device)
	require.Len(t, ifaces, 6)
	assert.Equal(t, "enp2s0f0", ifaces[0].osName)
	assert.Equal(t, "SFP1", ifaces[0].name)
}

func TestIfacesForDevice_Unknown(t *testing.T) {
	device := tg.Device{
		Vendor: "unknown-vendor",
		Model:  "unknown-model",
	}

	ifaces := ifacesForDevice(device)
	assert.Nil(t, ifaces)
}
