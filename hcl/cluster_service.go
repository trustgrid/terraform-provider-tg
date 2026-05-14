package hcl

import (
	"github.com/trustgrid/terraform-provider-tg/tg"
)

// ClusterService is the HCL representation of a V2 cluster L4 service.
// Writes go through /v2/cluster/{cluster_fqdn}/config/services[/{service_id}].
type ClusterService struct {
	ServiceID           string `tf:"service_id"`
	ClusterFQDN         string `tf:"cluster_fqdn"`
	Name                string `tf:"name"`
	Protocol            string `tf:"protocol"`
	Host                string `tf:"host"`
	Port                int    `tf:"port"`
	Description         string `tf:"description"`
	Enabled             bool   `tf:"enabled"`
	SourceInterface     string `tf:"source_interface"`
	SourceFromClusterIP bool   `tf:"source_from_cluster_ip"`
}

func (s ClusterService) ToTG() tg.Service {
	return tg.Service{
		ID:                  s.ServiceID,
		Name:                s.Name,
		Enabled:             s.Enabled,
		Host:                s.Host,
		Port:                s.Port,
		Protocol:            s.Protocol,
		Description:         s.Description,
		SourceInterface:     s.SourceInterface,
		SourceFromClusterIP: s.SourceFromClusterIP,
	}
}

func (s ClusterService) UpdateFromTG(svc tg.Service) HCL[tg.Service] {
	return ClusterService{
		ServiceID:           svc.ID,
		ClusterFQDN:         s.ClusterFQDN,
		Name:                svc.Name,
		Protocol:            svc.Protocol,
		Host:                svc.Host,
		Port:                svc.Port,
		Description:         svc.Description,
		Enabled:             svc.Enabled,
		SourceInterface:     svc.SourceInterface,
		SourceFromClusterIP: svc.SourceFromClusterIP,
	}
}
