package hcl

// ClusterActiveMember is the HCL projection of the `tg_cluster_active_member`
// resource. The `tf` tags drive hcl.DecodeResourceData / EncodeResourceData
// for round-tripping between Terraform's *schema.ResourceData and a Go
// struct so the resource Create/Update don't need manual d.Get(...).(string)
// type assertions.
type ClusterActiveMember struct {
	ClusterFQDN string `tf:"cluster_fqdn"`
	NodeID      string `tf:"node_id"`
}
