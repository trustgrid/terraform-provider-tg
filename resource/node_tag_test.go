package resource

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func TestCreateNodeTag_UsesConditionalMutation(t *testing.T) {
	var requestPath string
	var mutation nodeTagMutation
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path
		require.NoError(t, json.NewDecoder(r.Body).Decode(&mutation))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	previousClient := http.DefaultClient
	http.DefaultClient = server.Client()
	defer func() { http.DefaultClient = previousClient }()

	data := schema.TestResourceDataRaw(t, NodeTag().Schema, map[string]interface{}{
		"node_id": "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc",
		"name":    "purpose",
		"value":   "agent-first-demo",
	})
	client := &tg.Client{APIHost: strings.TrimPrefix(server.URL, "https://")}

	require.Empty(t, createNodeTag(context.Background(), data, client))
	require.Equal(t, "/node/35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc/tag/purpose/compare-and-set", requestPath)
	require.Nil(t, mutation.ExpectedValue)
	require.Equal(t, "agent-first-demo", *mutation.Value)
	require.Len(t, mutation.RequestID, len("node-tag-")+32)
	require.Equal(t, "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc/purpose", data.Id())
}

func TestImportNodeTag_ParsesNodeAndTag(t *testing.T) {
	data := schema.TestResourceDataRaw(t, NodeTag().Schema, map[string]interface{}{})
	data.SetId("35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc/purpose")

	resources, err := importNodeTag(context.Background(), data, nil)
	require.NoError(t, err)
	require.Len(t, resources, 1)
	require.Equal(t, "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc", data.Get("node_id"))
	require.Equal(t, "purpose", data.Get("name"))
}

func TestDeleteNodeTag_UsesExpectedValue(t *testing.T) {
	var mutation nodeTagMutation
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/node/35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc/tag/purpose/compare-and-set", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&mutation))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	previousClient := http.DefaultClient
	http.DefaultClient = server.Client()
	defer func() { http.DefaultClient = previousClient }()

	data := schema.TestResourceDataRaw(t, NodeTag().Schema, map[string]interface{}{
		"node_id": "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc",
		"name":    "purpose",
		"value":   "agent-first-demo",
	})
	data.SetId("35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc/purpose")
	client := &tg.Client{APIHost: strings.TrimPrefix(server.URL, "https://")}

	require.Empty(t, deleteNodeTag(context.Background(), data, client))
	require.NotNil(t, mutation.ExpectedValue)
	require.Equal(t, "agent-first-demo", *mutation.ExpectedValue)
	require.Nil(t, mutation.Value)
	require.Empty(t, data.Id())
}

func TestReadNodeTag_EvictsStateWhenTagWasRemoved(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/node/35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{
			"uid":  "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc",
			"tags": map[string]string{"owner": "platform"},
		}))
	}))
	defer server.Close()
	previousClient := http.DefaultClient
	http.DefaultClient = server.Client()
	defer func() { http.DefaultClient = previousClient }()

	data := schema.TestResourceDataRaw(t, NodeTag().Schema, map[string]interface{}{})
	data.SetId("35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc/purpose")
	client := &tg.Client{APIHost: strings.TrimPrefix(server.URL, "https://")}

	require.Empty(t, readNodeTag(context.Background(), data, client))
	require.Empty(t, data.Id())
}

func TestUpdateNodeTag_RejectsStalePriorValue(t *testing.T) {
	var mutation nodeTagMutation
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&mutation))
		http.Error(w, "tag changed", http.StatusConflict)
	}))
	defer server.Close()
	previousClient := http.DefaultClient
	http.DefaultClient = server.Client()
	defer func() { http.DefaultClient = previousClient }()

	data := schema.TestResourceDataRaw(t, NodeTag().Schema, map[string]interface{}{
		"node_id": "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc",
		"name":    "purpose",
		"value":   "new-value",
	})
	data.SetId("35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc/purpose")
	client := &tg.Client{APIHost: strings.TrimPrefix(server.URL, "https://")}

	diagnostics := writeNodeTag(
		context.Background(),
		data,
		client,
		stringPointer("stale-value"),
		stringPointer("new-value"),
	)
	require.NotEmpty(t, diagnostics)
	require.NotNil(t, mutation.ExpectedValue)
	require.Equal(t, "stale-value", *mutation.ExpectedValue)
	require.NotNil(t, mutation.Value)
	require.Equal(t, "new-value", *mutation.Value)
	require.Equal(t, "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc/purpose", data.Id())
}

func TestTaggingWriteTags_ChangesOnlyChangedNodeKeys(t *testing.T) {
	var mu sync.Mutex
	mutations := make(map[string]nodeTagMutation)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var mutation nodeTagMutation
		if err := json.NewDecoder(r.Body).Decode(&mutation); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		mu.Lock()
		mutations[r.URL.Path] = mutation
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	previousClient := http.DefaultClient
	http.DefaultClient = server.Client()
	defer func() { http.DefaultClient = previousClient }()
	client := &tg.Client{APIHost: strings.TrimPrefix(server.URL, "https://")}
	r := tagging{}
	before := map[string]string{"owner": "platform", "purpose": "old"}
	after := map[string]string{"owner": "platform", "purpose": "new"}

	_, err := r.writeTags(context.Background(), client, hcl.Tagging{
		NodeID: "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc",
		Tags:   after,
	}, before)
	require.NoError(t, err)
	require.Len(t, mutations, 1)
	mutation := mutations["/node/35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc/tag/purpose/compare-and-set"]
	require.NotNil(t, mutation.ExpectedValue)
	require.Equal(t, "old", *mutation.ExpectedValue)
	require.NotNil(t, mutation.Value)
	require.Equal(t, "new", *mutation.Value)
}

func TestNodeTagFromID_RejectsMalformedImportID(t *testing.T) {
	for _, id := range []string{"", "node/purpose", "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc/", "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc/a/b"} {
		t.Run(id, func(t *testing.T) {
			_, _, err := nodeTagFromID(id)
			require.Error(t, err)
		})
	}
}
