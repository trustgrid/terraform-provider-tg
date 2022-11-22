package tg

type KVMImage struct {
	Description string `json:"description"`
	DisplayName string `json:"displayName"`
	ID          string `json:"id,omitempty"`
	Location    string `json:"location"`
	OS          string `json:"os"`
}

type KVMVolume struct {
	Name          string `json:"name"`
	DeviceType    string `json:"deviceType"`
	DeviseBus     string `json:"deviceBus"`
	Size          int    `json:"size,omitempty"`
	ProvisionType string `json:"provisionType"`
	Path          string `json:"path,omitempty"`
	Encrypted     bool   `json:"encrypted"`
}
