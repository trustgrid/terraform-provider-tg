package tg

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServicesConfig_UnmarshalJSON_V1ArrayShape(t *testing.T) {
	body := []byte(`{"services":[
		{"id":"svc-1","name":"foo","host":"10.0.0.1","port":80,"protocol":"tcp","enabled":true},
		{"id":"svc-2","name":"bar","host":"10.0.0.2","port":443,"protocol":"tcp","enabled":false}
	]}`)

	var cfg ServicesConfig
	require.NoError(t, json.Unmarshal(body, &cfg))

	require.Len(t, cfg.Services, 2)
	assert.Equal(t, "svc-1", cfg.Services[0].ID)
	assert.Equal(t, "foo", cfg.Services[0].Name)
	assert.True(t, cfg.Services[0].Enabled)
	assert.Equal(t, "svc-2", cfg.Services[1].ID)
	assert.False(t, cfg.Services[1].Enabled)
}

func TestServicesConfig_UnmarshalJSON_V2ItemsShape(t *testing.T) {
	body := []byte(`{"items":{
		"svc-1":{"id":"svc-1","name":"foo","host":"10.0.0.1","port":80,"protocol":"tcp","enabled":true},
		"svc-2":{"id":"svc-2","name":"bar","host":"10.0.0.2","port":443,"protocol":"tcp","enabled":false}
	}}`)

	var cfg ServicesConfig
	require.NoError(t, json.Unmarshal(body, &cfg))

	require.Len(t, cfg.Services, 2)
	// Sorted by ID for stable iteration order.
	assert.Equal(t, "svc-1", cfg.Services[0].ID)
	assert.Equal(t, "foo", cfg.Services[0].Name)
	assert.Equal(t, "svc-2", cfg.Services[1].ID)
}

func TestServicesConfig_UnmarshalJSON_V2InheritsKeyAsID(t *testing.T) {
	// V2 server may omit the inner id field when it matches the map key.
	body := []byte(`{"items":{
		"svc-key-1":{"name":"foo","host":"10.0.0.1","port":80,"protocol":"tcp"}
	}}`)

	var cfg ServicesConfig
	require.NoError(t, json.Unmarshal(body, &cfg))

	require.Len(t, cfg.Services, 1)
	assert.Equal(t, "svc-key-1", cfg.Services[0].ID)
	assert.Equal(t, "foo", cfg.Services[0].Name)
}

func TestServicesConfig_UnmarshalJSON_Empty(t *testing.T) {
	cases := map[string]string{
		"empty object":  `{}`,
		"v1 empty":      `{"services":[]}`,
		"v1 null":       `{"services":null}`,
		"v2 empty":      `{"items":{}}`,
		"v2 null":       `{"items":null}`,
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			var cfg ServicesConfig
			require.NoError(t, json.Unmarshal([]byte(body), &cfg))
			assert.Empty(t, cfg.Services)
		})
	}
}

func TestServicesConfig_UnmarshalJSON_BothShapesPresentPrefersV2(t *testing.T) {
	// If a buggy backend returns both shapes, prefer V2 (items map) as the
	// post-migration source of truth.
	body := []byte(`{
		"items":{"svc-new":{"id":"svc-new","name":"new"}},
		"services":[{"id":"svc-old","name":"old"}]
	}`)

	var cfg ServicesConfig
	require.NoError(t, json.Unmarshal(body, &cfg))

	require.Len(t, cfg.Services, 1)
	assert.Equal(t, "svc-new", cfg.Services[0].ID)
}

func TestServicesConfig_UnmarshalJSON_Invalid(t *testing.T) {
	var cfg ServicesConfig
	err := json.Unmarshal([]byte(`{"services":"not-an-array"}`), &cfg)
	assert.Error(t, err)
}

// TestServicesConfig_PR242_RegressionV2DoesNotEmptyV1Resources is the safety-net
// regression test for the destructive scenario that closed PR #242. Before the
// dual-shape decoder, parsing a V2 items-shape response into the V1-only struct
// silently left Services empty, which would make legacy tg_service resources
// see "service deleted out of band" and plan destructive recreates. This test
// pins the fix: a V2-shape response MUST yield the real list of services.
func TestServicesConfig_PR242_RegressionV2DoesNotEmptyV1Resources(t *testing.T) {
	// Realistic V2 cluster GET /cluster/{fqdn} response excerpt.
	clusterBody := []byte(`{
		"name":"hq",
		"fqdn":"hq.example.test",
		"config":{
			"services":{
				"items":{
					"abc-123":{"id":"abc-123","name":"existing-svc","host":"10.0.0.1","port":8080,"protocol":"tcp","enabled":true},
					"def-456":{"id":"def-456","name":"other-svc","host":"10.0.0.2","port":443,"protocol":"tcp","enabled":true}
				}
			}
		}
	}`)

	var cluster Cluster
	require.NoError(t, json.Unmarshal(clusterBody, &cluster))

	require.NotNil(t, cluster.Config.Services, "Services must not be nil — would cause tg_service Read to drop the resource from state")
	require.Len(t, cluster.Config.Services.Services, 2, "BOTH services must be visible — empty list triggers destructive recreate plan")

	ids := []string{cluster.Config.Services.Services[0].ID, cluster.Config.Services.Services[1].ID}
	assert.Contains(t, ids, "abc-123")
	assert.Contains(t, ids, "def-456")
}

func TestServicesConfig_V1Cluster_StillReadsCleanly(t *testing.T) {
	// The other half of the safety net: V1 clusters must still decode after
	// the change. A customer on V1 upgrading the provider should see no
	// difference in tg_service behavior.
	clusterBody := []byte(`{
		"name":"hq",
		"fqdn":"hq.example.test",
		"config":{
			"services":{
				"services":[
					{"id":"abc-123","name":"existing-svc","host":"10.0.0.1","port":8080,"protocol":"tcp","enabled":true}
				]
			}
		}
	}`)

	var cluster Cluster
	require.NoError(t, json.Unmarshal(clusterBody, &cluster))

	require.NotNil(t, cluster.Config.Services)
	require.Len(t, cluster.Config.Services.Services, 1)
	assert.Equal(t, "abc-123", cluster.Config.Services.Services[0].ID)
}

func TestConnectorsConfig_UnmarshalJSON_V1ArrayShape(t *testing.T) {
	body := []byte(`{"connectors":[
		{"id":"conn-1","node":"local","service":"127.0.0.1:8080","port":8081,"protocol":"tcp","enabled":true}
	]}`)

	var cfg ConnectorsConfig
	require.NoError(t, json.Unmarshal(body, &cfg))

	require.Len(t, cfg.Connectors, 1)
	assert.Equal(t, "conn-1", cfg.Connectors[0].ID)
	assert.Equal(t, "local", cfg.Connectors[0].Node)
}

func TestConnectorsConfig_UnmarshalJSON_V2ItemsShape(t *testing.T) {
	body := []byte(`{"items":{
		"conn-1":{"id":"conn-1","node":"local","service":"127.0.0.1:8080","port":8081,"protocol":"tcp","enabled":true},
		"conn-2":{"id":"conn-2","node":"local","service":"127.0.0.1:9090","port":9091,"protocol":"tcp","enabled":false}
	}}`)

	var cfg ConnectorsConfig
	require.NoError(t, json.Unmarshal(body, &cfg))

	require.Len(t, cfg.Connectors, 2)
	assert.Equal(t, "conn-1", cfg.Connectors[0].ID)
	assert.Equal(t, "conn-2", cfg.Connectors[1].ID)
}

func TestConnectorsConfig_UnmarshalJSON_Empty(t *testing.T) {
	cases := map[string]string{
		"empty object": `{}`,
		"v1 empty":     `{"connectors":[]}`,
		"v1 null":      `{"connectors":null}`,
		"v2 empty":     `{"items":{}}`,
		"v2 null":      `{"items":null}`,
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			var cfg ConnectorsConfig
			require.NoError(t, json.Unmarshal([]byte(body), &cfg))
			assert.Empty(t, cfg.Connectors)
		})
	}
}

func TestConnectorsConfig_PR242_RegressionV2DoesNotEmptyV1Resources(t *testing.T) {
	// Same safety-net as the services version: V2 connectors response must
	// yield the real list, not an empty one.
	clusterBody := []byte(`{
		"name":"hq",
		"fqdn":"hq.example.test",
		"config":{
			"connectors":{
				"items":{
					"abc-123":{"id":"abc-123","node":"local","service":"127.0.0.1:8080","port":8081,"protocol":"tcp","enabled":true}
				}
			}
		}
	}`)

	var cluster Cluster
	require.NoError(t, json.Unmarshal(clusterBody, &cluster))

	require.NotNil(t, cluster.Config.Connectors)
	require.Len(t, cluster.Config.Connectors.Connectors, 1)
	assert.Equal(t, "abc-123", cluster.Config.Connectors.Connectors[0].ID)
}

func TestServicesConfig_MarshalJSON_ProducesV1Shape(t *testing.T) {
	// Marshaling stays as V1 array shape — the legacy tg_service write path
	// PUTs {services:[...]} to the V1 endpoint and that must continue working
	// for V1 clusters during the migration window.
	cfg := ServicesConfig{
		Services: []Service{
			{ID: "svc-1", Name: "foo", Host: "10.0.0.1", Port: 80, Protocol: "tcp"},
		},
	}
	body, err := json.Marshal(cfg)
	require.NoError(t, err)
	assert.JSONEq(t, `{"services":[{"id":"svc-1","name":"foo","enabled":false,"host":"10.0.0.1","port":80,"protocol":"tcp","description":""}]}`, string(body))
}

func TestService_V2Fields_OmitemptyOnV1Payload(t *testing.T) {
	// V1 customers must continue PUTting payloads that don't include the
	// V2-only fields. Empty values must be omitted to keep the wire shape
	// identical to what existed before.
	svc := Service{ID: "svc-1", Name: "foo", Host: "10.0.0.1", Port: 80, Protocol: "tcp"}
	body, err := json.Marshal(svc)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "sourceInterface")
	assert.NotContains(t, string(body), "sourceFromClusterIP")
}

func TestService_V2Fields_PresentWhenSet(t *testing.T) {
	svc := Service{ID: "svc-1", Name: "foo", Host: "10.0.0.1", Port: 80, Protocol: "tcp",
		SourceInterface: "ens192", SourceFromClusterIP: true}
	body, err := json.Marshal(svc)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"sourceInterface":"ens192"`)
	assert.Contains(t, string(body), `"sourceFromClusterIP":true`)
}
