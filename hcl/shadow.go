package hcl

import (
	"sort"
	"strings"
)

type Shadow struct {
	PackageVersion string            `tf:"package_version,omitempty"`
	ClusterMaster  bool              `tf:"cluster_master,omitempty"`
	Nics           []string          `tf:"nics,omitempty"`
	Reported       map[string]string `tf:"reported,omitempty"`
}

func (s *Shadow) UpdateFromTG(shadow map[string]any) {
	pv, ok := shadow["package.version"].(string)
	if ok {
		s.PackageVersion = pv
	}

	cm, ok := shadow["cluster.master"].(string)
	if ok {
		s.ClusterMaster = cm == "true"
	}

	if s.Reported == nil {
		s.Reported = make(map[string]string)
	}

	nics := make([]string, 0)

	for k, v := range shadow {
		vs, ok := v.(string)
		if ok {
			s.Reported[k] = vs
		}
		if strings.HasPrefix(k, "nic.") && strings.HasSuffix(k, ".ip") {
			name := strings.Replace(strings.Replace(k, "nic.", "", 1), ".ip", "", 1)
			nics = append(nics, name)
		}
	}

	sort.SliceStable(nics, func(i, j int) bool {
		return nics[i] < nics[j]
	})

	s.Nics = nics
}
