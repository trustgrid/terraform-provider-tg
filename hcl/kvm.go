package hcl

import "github.com/trustgrid/terraform-provider-tg/tg"

type KVMImage struct {
	NodeID string `tf:"node_id"`
	UID    string `tf:"uid"`

	Description string `tf:"description"`
	DisplayName string `tf:"display_name"`
	Location    string `tf:"location"`
	OS          string `tf:"os"`
}

func (h *KVMImage) ToTG() *tg.KVMImage {
	return &tg.KVMImage{
		DisplayName: h.DisplayName,
		Description: h.Description,
		Location:    h.Location,
		OS:          h.OS,
	}
}

func (h *KVMImage) UpdateFromTG(r tg.KVMImage) {
	h.DisplayName = r.DisplayName
	h.Description = r.Description
	h.Location = r.Location
	h.OS = r.OS
}

func (h *KVMImage) ResourceURL(ID string) string {
	return h.URL() + "/" + ID
}

func (h *KVMImage) URL() string {
	return "/v2/node/" + h.NodeID + "/kvm/image"
}

type KVMVolume struct {
	NodeID string `tf:"node_id"`

	Name          string `tf:"name"`
	DeviceType    string `tf:"device_type"`
	DeviseBus     string `tf:"device_bus"`
	Size          int    `tf:"size"`
	ProvisionType string `tf:"provision_type"`
	Encrypted     bool   `tf:"encrypted"`
	Path          string `tf:"path"`
}

func (h *KVMVolume) ToTG() *tg.KVMVolume {
	return &tg.KVMVolume{
		Name:          h.Name,
		DeviceType:    h.DeviceType,
		DeviseBus:     h.DeviseBus,
		Size:          h.Size,
		ProvisionType: h.ProvisionType,
		Path:          h.Path,
		Encrypted:     h.Encrypted,
	}
}

func (h *KVMVolume) URL() string {
	return "/v2/node/" + h.NodeID + "/kvm/volume"
}

func (h *KVMVolume) ResourceURL() string {
	return h.URL() + "/" + h.Name
}

func (h *KVMVolume) UpdateFromTG(r tg.KVMVolume) {
	h.DeviceType = r.DeviceType
	h.DeviseBus = r.DeviseBus
	h.Size = r.Size
	h.ProvisionType = r.ProvisionType
	h.Path = r.Path
	h.Encrypted = r.Encrypted
}
