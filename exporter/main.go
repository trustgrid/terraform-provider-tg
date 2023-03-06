package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type Resource struct {
	resourceType string
	name         string
	data         map[string]any
}

func (r *Resource) Set(key string, value any) error {
	r.data[key] = value
	return nil
}

func (r *Resource) String() string {
	out := `resource "` + r.resourceType + `" "` + r.name + `" {` + "\n"
	for k, v := range r.data {
		if reflect.TypeOf(v).Kind() == reflect.String {
			out += fmt.Sprintf("\t%s = \"%s\"\n", k, v)
		} else {
			out += "\t" + k + " = " + fmt.Sprintln(v)
		}
	}
	out += "}\n"

	return out
}

func NewResource(resourceType string, name string, data any) *Resource {
	r := &Resource{
		resourceType: resourceType,
		name:         name,
		data:         make(map[string]any),
	}

	must(hcl.EncodeResourceData(data, r))

	return r
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	ctx := context.Background()

	tgc, err := tg.NewClient(ctx, tg.ClientParams{
		APIKey:    os.Getenv("TG_API_KEY_ID"),
		APISecret: os.Getenv("TG_API_KEY_SECRET"),
		APIHost:   os.Getenv("TG_API_HOST"),
		JWT:       os.Getenv("JWT"),
	})

	if err != nil {
		panic(err)
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: ./exporter <nodeID>")
		os.Exit(1)
	}
	nodeID := os.Args[1]

	tgnode := tg.Node{}
	err = tgc.Get(ctx, "/node/"+nodeID, &tgnode)
	switch {
	case errors.Is(err, tg.ErrNotFound):
		fmt.Printf("Node %s not found\n", nodeID)
		return
	case err != nil:
		panic(err)
	}

	if tgnode.Config.SNMP.Enabled {
		tgnode.Config.SNMP.NodeID = nodeID
		fmt.Println(NewResource("tg_snmp", tgnode.Name+"_snmp", tgnode.Config.SNMP).String())
	}

	if tgnode.Config.Gateway.Enabled {
		tgnode.Config.Gateway.NodeID = nodeID
		fmt.Println(NewResource("tg_gateway_config", tgnode.Name+"_gateway", tgnode.Config.Gateway).String())
	}

	if tgnode.Config.ZTNA.Enabled {
		tgnode.Config.ZTNA.NodeID = nodeID
		fmt.Println(NewResource("tg_ztna_gateway_config", tgnode.Name+"_ztna_gateway", tgnode.Config.ZTNA).String())
	}

	if tgnode.Config.Cluster.Enabled {
		tgnode.Config.Cluster.NodeID = nodeID
		fmt.Println(NewResource("tg_node_cluster_config", tgnode.Name+"_cluster_config", tgnode.Config.Cluster).String())
	}
}
