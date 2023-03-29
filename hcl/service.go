package hcl

import (
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type Service struct {
	NodeID      string `tf:"node_id"`
	ClusterFQDN string `tf:"cluster_fqdn"`
	Name        string `tf:"name"`
	Host        string `tf:"host"`
	Port        int    `tf:"port"`
	Protocol    string `tf:"protocol"`
	Description string `tf:"description"`
}

func (s *Service) UpdateFromTG(svc tg.Service) {
	s.Name = svc.Name
	s.Host = svc.Host
	s.Port = svc.Port
	s.Protocol = svc.Protocol
	s.Description = svc.Description
}

func (s *Service) ToTG(id string) tg.Service {
	return tg.Service{
		ID:          id,
		Name:        s.Name,
		Enabled:     true,
		Host:        s.Host,
		Port:        s.Port,
		Protocol:    s.Protocol,
		Description: s.Description,
	}
}
