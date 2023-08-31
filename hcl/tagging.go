package hcl

type Tagging struct {
	NodeID      string            `tf:"node_id"`
	ClusterFQDN string            `tf:"cluster_fqdn"`
	Tags        map[string]string `tf:"tags"`
}
