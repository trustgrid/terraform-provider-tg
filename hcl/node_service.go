package hcl

import (
	"github.com/trustgrid/terraform-provider-tg/tg"
)

// NodeService is the HCL representation of a V2 node L4 service.
// Writes go through /v2/node/{node_id}/config/services[/{service_id}].
//
// SourceFromClusterIP is intentionally omitted — there is no cluster VIP on a
// single node. SourceInterface is included for parity with the cluster
// resource (the underlying V2 service body carries the same field).
type NodeService struct {
	ServiceID       string `tf:"service_id"`
	NodeID          string `tf:"node_id"`
	Name            string `tf:"name"`
	Protocol        string `tf:"protocol"`
	Host            string `tf:"host"`
	Port            int    `tf:"port"`
	Description     string `tf:"description"`
	Enabled         bool   `tf:"enabled"`
	SourceInterface string `tf:"source_interface"`
}

func (s NodeService) ToTG() tg.Service {
	return tg.Service{
		ID:              s.ServiceID,
		Name:            s.Name,
		Enabled:         s.Enabled,
		Host:            s.Host,
		Port:            s.Port,
		Protocol:        s.Protocol,
		Description:     s.Description,
		SourceInterface: s.SourceInterface,
	}
}

func (s NodeService) UpdateFromTG(svc tg.Service) HCL[tg.Service] {
	return NodeService{
		ServiceID:       svc.ID,
		NodeID:          s.NodeID,
		Name:            svc.Name,
		Protocol:        svc.Protocol,
		Host:            svc.Host,
		Port:            svc.Port,
		Description:     svc.Description,
		Enabled:         svc.Enabled,
		SourceInterface: svc.SourceInterface,
	}
}
