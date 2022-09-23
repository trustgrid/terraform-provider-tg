package tg

type Container struct {
	NodeID      string `tf:"node_id" json:"-"`
	ClusterFQDN string `tf:"cluster_fqdn" json:"-"`
	ID          string `tf:"id" json:"id"`
	Command     string `tf:"command" json:"command,omitempty"`
	Description string `tf:"description" json:"description"`
	Enabled     bool   `tf:"enabled" json:"enabled"`
	ExecType    string `tf:"exec_type" json:"execType"`
	Hostname    string `tf:"hostname" json:"hostname,omitempty"`
	Image       struct {
		Repository string `json:"repository"`
		Tag        string `json:"tag"`
	} `json:"image"`
	ImageRepository     string `tf:"image_repository" json:"-"`
	ImageTag            string `tf:"image_tag" json:"-"`
	Name                string `tf:"name" json:"name"`
	Privileged          bool   `tf:"privileged" json:"privileged"`
	RequireConnectivity bool   `tf:"require_connectivity" json:"requireConnectivity"`
	StopTime            int    `tf:"stop_time" json:"stopTime,omitempty"`
	UseInit             bool   `tf:"use_init" json:"useInit"`
	User                string `tf:"user" json:"user,omitempty"`

	AddCaps        []interface{}          `tf:"add_caps" json:"-"`
	DropCaps       []interface{}          `tf:"drop_caps" json:"-"`
	Variables      map[string]interface{} `tf:"variables" json:"-"`
	LogMaxFileSize int                    `tf:"log_max_file_size" json:"-"`
	LogMaxNumFiles int                    `tf:"log_max_num_files" json:"-"`
}

type containerVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ContainerConfig struct {
	Capabilities struct {
		AddCaps  []string `json:"addCaps,omitempty"`
		DropCaps []string `json:"dropCaps,omitempty"`
	} `json:"capabilities,omitempty"`
	Variables []containerVar `json:"variables,omitempty"`
	Logging   struct {
		MaxFileSize int `json:"maxFileSize,omitempty"`
		NumFiles    int `json:"numFiles,omitempty"`
	} `json:"logging,omitempty"`
}

func (c Container) Config() ContainerConfig {
	cc := ContainerConfig{}
	if len(c.AddCaps) > 0 {
		cc.Capabilities.AddCaps = make([]string, len(c.AddCaps))
		for i, c := range c.AddCaps {
			cc.Capabilities.AddCaps[i] = c.(string)
		}
	}
	if len(c.DropCaps) > 0 {
		cc.Capabilities.DropCaps = make([]string, len(c.DropCaps))
		for i, c := range c.DropCaps {
			cc.Capabilities.DropCaps[i] = c.(string)
		}
	}
	if len(c.Variables) > 0 {
		for k, v := range c.Variables {
			cc.Variables = append(cc.Variables, containerVar{Name: k, Value: v.(string)})
		}
	}
	cc.Logging.MaxFileSize = c.LogMaxFileSize
	cc.Logging.NumFiles = c.LogMaxNumFiles
	return cc
}
