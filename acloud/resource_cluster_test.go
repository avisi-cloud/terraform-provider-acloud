package acloud

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

const (
	testOrg         = "org"
	testEnvironment = "prod"
	testClusterSlug = "example-cluster"
)

// fakeClusterAPI is a minimal platform-api for a single cluster. It records the raw request bodies so tests can check
// which JSON keys the provider sends.
type fakeClusterAPI struct {
	t      *testing.T
	server *httptest.Server

	mu             sync.Mutex
	nodePortAccess acloudapi.NodePortAccess
	requests       int
	createBodies   []map[string]json.RawMessage
	updateBodies   []map[string]json.RawMessage
}

func newFakeClusterAPI(t *testing.T) *fakeClusterAPI {
	t.Helper()
	api := &fakeClusterAPI{t: t}
	clusterPath := "/api/v1/orgs/" + testOrg + "/clusters/" + testEnvironment + "/" + testClusterSlug

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/orgs/"+testOrg+"/clusters/"+testEnvironment, func(w http.ResponseWriter, r *http.Request) {
		body := api.record(r, &api.createBodies)
		if raw, ok := body["nodePortAccess"]; ok {
			api.mu.Lock()
			api.nodePortAccess = api.decodeNodePortAccess(raw)
			api.mu.Unlock()
		}
		api.writeCluster(w)
	})
	mux.HandleFunc("GET "+clusterPath, func(w http.ResponseWriter, r *http.Request) {
		api.record(r, nil)
		api.writeCluster(w)
	})
	mux.HandleFunc("PATCH "+clusterPath, func(w http.ResponseWriter, r *http.Request) {
		body := api.record(r, &api.updateBodies)
		if raw, ok := body["nodePortAccess"]; ok {
			api.mu.Lock()
			api.nodePortAccess = api.decodeNodePortAccess(raw)
			api.mu.Unlock()
		}
		api.writeCluster(w)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		http.NotFound(w, r)
	})

	api.server = httptest.NewServer(mux)
	t.Cleanup(api.server.Close)
	return api
}

func (api *fakeClusterAPI) meta() ConfiguredProvider {
	return ConfiguredProvider{
		Client:       acloudapi.NewClient(nil, acloudapi.ClientOpts{APIUrl: api.server.URL}),
		Organisation: testOrg,
	}
}

func (api *fakeClusterAPI) record(r *http.Request, bodies *[]map[string]json.RawMessage) map[string]json.RawMessage {
	api.mu.Lock()
	defer api.mu.Unlock()
	api.requests++
	if bodies == nil {
		return nil
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		api.t.Errorf("failed to read request body: %v", err)
	}
	body := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &body); err != nil {
		api.t.Errorf("failed to unmarshal request body %q: %v", raw, err)
	}
	*bodies = append(*bodies, body)
	return body
}

func (api *fakeClusterAPI) decodeNodePortAccess(raw json.RawMessage) acloudapi.NodePortAccess {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		api.t.Errorf("nodePortAccess in request body is not a string: %s", raw)
	}
	nodePortAccess, err := acloudapi.ParseNodePortAccess(value)
	if err != nil {
		api.t.Errorf("invalid nodePortAccess in request body: %v", err)
	}
	return nodePortAccess
}

func (api *fakeClusterAPI) writeCluster(w http.ResponseWriter) {
	api.mu.Lock()
	cluster := acloudapi.Cluster{
		Identity:                     "cluster-uuid",
		Name:                         testClusterSlug,
		Slug:                         testClusterSlug,
		EnvironmentSlug:              testEnvironment,
		Region:                       "eu-west-1",
		Version:                      "1.34",
		PodSecurityStandardsProfile:  "PRIVILEGED",
		EnableMultiAvailAbilityZones: true,
		EnableNetworkEncryption:      true,
		NodePortAccess:               api.nodePortAccess,
		Status:                       string(ClusterStateRunning),
		CloudAccount:                 &acloudapi.CloudAccountReference{Identity: "account-1"},
	}
	api.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(cluster); err != nil {
		api.t.Errorf("failed to encode cluster: %v", err)
	}
}

func clusterConfig(extra map[string]interface{}) map[string]interface{} {
	raw := map[string]interface{}{
		"name":                   testClusterSlug,
		"environment":            testEnvironment,
		"region":                 "eu-west-1",
		"version":                "1.34",
		"cloud_account_identity": "account-1",
	}
	for k, v := range extra {
		raw[k] = v
	}
	return raw
}

// clusterState returns the state of an existing cluster, with node_port_access set to nodePortAccess, or absent
// when nodePortAccess is empty (state written by a provider version without the attribute).
func clusterState(t *testing.T, nodePortAccess acloudapi.NodePortAccess) *terraform.InstanceState {
	t.Helper()
	d := schema.TestResourceDataRaw(t, resourceCluster().Schema, clusterConfig(nil))
	d.SetId("cluster-uuid")
	d.Set("slug", testClusterSlug)
	d.Set("status", string(ClusterStateRunning))
	state := d.State()
	delete(state.Attributes, "node_port_access")
	if nodePortAccess != "" {
		state.Attributes["node_port_access"] = string(nodePortAccess)
	}
	return state
}

func TestResourceClusterNodePortAccessSchema(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("provider schema is invalid: %v", err)
	}

	s := resourceCluster().Schema["node_port_access"]
	if s == nil {
		t.Fatal("node_port_access is missing from the acloud_cluster resource schema")
	}
	if s.Type != schema.TypeString || !s.Optional || s.Computed || s.ForceNew {
		t.Fatalf("node_port_access must be an optional, non-computed, in-place updatable string, got %+v", s)
	}
	if s.Default != string(acloudapi.NodePortAccessInternal) {
		t.Fatalf("node_port_access default = %v, want %q", s.Default, acloudapi.NodePortAccessInternal)
	}

	ds := dataSourceCluster().Schema["node_port_access"]
	if ds == nil {
		t.Fatal("node_port_access is missing from the acloud_cluster data source schema")
	}
	if ds.Type != schema.TypeString || !ds.Computed || ds.Optional || ds.Required {
		t.Fatalf("data source node_port_access must be a computed string, got %+v", ds)
	}
}

func TestResourceClusterNodePortAccessValidation(t *testing.T) {
	s := resourceCluster().Schema["node_port_access"]
	for _, value := range acloudapi.AllNodePortAccess {
		t.Run(string(value), func(t *testing.T) {
			if _, errs := s.ValidateFunc(string(value), "node_port_access"); len(errs) > 0 {
				t.Fatalf("expected %q to be valid, got %v", value, errs)
			}
		})
	}

	for _, value := range []string{"", "Public", "INTERNAL", "private", "allow-list"} {
		t.Run("invalid "+value, func(t *testing.T) {
			api := newFakeClusterAPI(t)
			diags := Provider().ValidateResource("acloud_cluster", terraform.NewResourceConfigRaw(clusterConfig(map[string]interface{}{
				"node_port_access": value,
			})))
			if !diags.HasError() {
				t.Fatalf("expected node_port_access %q to be rejected", value)
			}
			if api.requests != 0 {
				t.Fatalf("expected no API calls during validation, got %d", api.requests)
			}
		})
	}
}

func TestResourceClusterCreateNodePortAccess(t *testing.T) {
	type createTest struct {
		name   string
		config map[string]interface{}
		want   acloudapi.NodePortAccess
	}
	tests := []createTest{{
		name:   "not configured sends the default",
		config: clusterConfig(nil),
		want:   acloudapi.NodePortAccessInternal,
	}}
	for _, value := range acloudapi.AllNodePortAccess {
		tests = append(tests, createTest{
			name:   string(value) + " is sent",
			config: clusterConfig(map[string]interface{}{"node_port_access": string(value)}),
			want:   value,
		})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := tt.want
			api := newFakeClusterAPI(t)
			d := schema.TestResourceDataRaw(t, resourceCluster().Schema, tt.config)

			if diags := resourceClusterCreate(context.Background(), d, api.meta()); diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}

			if len(api.createBodies) != 1 {
				t.Fatalf("expected 1 create request, got %d", len(api.createBodies))
			}
			raw, ok := api.createBodies[0]["nodePortAccess"]
			if !ok {
				t.Fatal("create request does not contain nodePortAccess")
			}
			if want := `"` + string(value) + `"`; string(raw) != want {
				t.Fatalf("nodePortAccess in create request = %s, want %s", raw, want)
			}
			if got := d.Get("node_port_access").(string); got != string(value) {
				t.Fatalf("node_port_access in state = %q, want %q", got, value)
			}
		})
	}
}

func TestResourceClusterReadNodePortAccess(t *testing.T) {
	for _, value := range acloudapi.AllNodePortAccess {
		t.Run(string(value), func(t *testing.T) {
			api := newFakeClusterAPI(t)
			api.nodePortAccess = value

			d := resourceCluster().Data(clusterState(t, ""))
			if diags := resourceClusterRead(context.Background(), d, api.meta()); diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}
			if got := d.Get("node_port_access").(string); got != string(value) {
				t.Fatalf("resource node_port_access = %q, want %q", got, value)
			}

			ds := schema.TestResourceDataRaw(t, dataSourceCluster().Schema, map[string]interface{}{
				"environment": testEnvironment,
				"slug":        testClusterSlug,
			})
			if diags := dataClusterRead(context.Background(), ds, api.meta()); diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}
			if got := ds.Get("node_port_access").(string); got != string(value) {
				t.Fatalf("data source node_port_access = %q, want %q", got, value)
			}
		})
	}
}

func TestResourceClusterUpdateNodePortAccess(t *testing.T) {
	tests := []struct {
		name string
		// server is the value on the API; state is the value in state, empty meaning absent.
		server acloudapi.NodePortAccess
		state  acloudapi.NodePortAccess
		config map[string]interface{}
		// want is the expected nodePortAccess in the PATCH body; empty means the key must be left out.
		want acloudapi.NodePortAccess
	}{
		{
			name:   "changed value is sent",
			server: acloudapi.NodePortAccessInternal,
			state:  acloudapi.NodePortAccessInternal,
			config: clusterConfig(map[string]interface{}{"node_port_access": "public"}),
			want:   acloudapi.NodePortAccessPublic,
		},
		{
			name:   "changed value is sent from public",
			server: acloudapi.NodePortAccessPublic,
			state:  acloudapi.NodePortAccessPublic,
			config: clusterConfig(map[string]interface{}{"node_port_access": "allowlist"}),
			want:   acloudapi.NodePortAccessAllowlist,
		},
		{
			name:   "unchanged configured value is left out",
			server: acloudapi.NodePortAccessAllowlist,
			state:  acloudapi.NodePortAccessAllowlist,
			config: clusterConfig(map[string]interface{}{"node_port_access": "allowlist", "description": "changed"}),
		},
		{
			name:   "unchanged default is left out",
			server: acloudapi.NodePortAccessInternal,
			state:  acloudapi.NodePortAccessInternal,
			config: clusterConfig(map[string]interface{}{"description": "changed"}),
		},
		{
			name:   "removing the attribute resets to the default",
			server: acloudapi.NodePortAccessPublic,
			state:  acloudapi.NodePortAccessPublic,
			config: clusterConfig(nil),
			want:   acloudapi.NodePortAccessInternal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			r := resourceCluster()
			api := newFakeClusterAPI(t)
			api.nodePortAccess = tt.server
			meta := api.meta()

			state := clusterState(t, tt.state)
			diff, err := r.Diff(ctx, state, terraform.NewResourceConfigRaw(tt.config), meta)
			if err != nil {
				t.Fatalf("unexpected diff error: %v", err)
			}
			if diff == nil || diff.Empty() {
				t.Fatal("expected a diff")
			}
			if diff.RequiresNew() {
				t.Fatalf("expected an in-place update, got a replacement: %v", diff)
			}

			newState, diags := r.Apply(ctx, state, diff, meta)
			if diags.HasError() {
				t.Fatalf("unexpected apply error: %v", diags)
			}

			if len(api.updateBodies) != 1 {
				t.Fatalf("expected 1 update request, got %d", len(api.updateBodies))
			}
			raw, sent := api.updateBodies[0]["nodePortAccess"]
			switch {
			case tt.want == "" && sent:
				t.Fatalf("update request contains nodePortAccess %s, want it left out", raw)
			case tt.want != "" && !sent:
				t.Fatalf("update request does not contain nodePortAccess, want %q", tt.want)
			case tt.want != "" && string(raw) != `"`+string(tt.want)+`"`:
				t.Fatalf("nodePortAccess in update request = %s, want %q", raw, tt.want)
			}

			wantState := tt.server
			if tt.want != "" {
				wantState = tt.want
			}
			if got := newState.Attributes["node_port_access"]; got != string(wantState) {
				t.Fatalf("node_port_access in state after update = %q, want %q", got, wantState)
			}
		})
	}
}

func TestResourceClusterNodePortAccessNotConfiguredDiff(t *testing.T) {
	r := resourceCluster()
	for _, value := range acloudapi.AllNodePortAccess {
		t.Run(string(value), func(t *testing.T) {
			diff, err := r.Diff(context.Background(), clusterState(t, value), terraform.NewResourceConfigRaw(clusterConfig(nil)), nil)
			if err != nil {
				t.Fatalf("unexpected diff error: %v", err)
			}
			var attr *terraform.ResourceAttrDiff
			if diff != nil {
				attr = diff.Attributes["node_port_access"]
			}
			if value == acloudapi.NodePortAccessInternal {
				if attr != nil {
					t.Fatalf("expected no diff for node_port_access at the default, got %#v", attr)
				}
				return
			}
			if attr == nil || attr.Old != string(value) || attr.New != string(acloudapi.NodePortAccessInternal) || attr.RequiresNew {
				t.Fatalf("expected an in-place diff %q -> %q, got %#v", value, acloudapi.NodePortAccessInternal, attr)
			}
		})
	}
}
