package acloud

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

const (
	bootstrapTestOrg        = "bootstrap-org"
	bootstrapTestEnv        = "byon-env"
	bootstrapTestCluster    = "byon-cluster"
	bootstrapTestNodePoolID = int64(42)
	// bootstrapTestFakeToken is a fake credential, used to check that join configuration values end up in state and
	// never in diagnostics.
	bootstrapTestFakeToken = "FAKE-TEST-ONLY-bootstrap-token.0000000000000000"
)

var bootstrapSensitiveAttributes = []string{"user_data", "kubelet_config", "join_command", "install_script", "upgrade_script"}

func bootstrapFakeJoinConfig() acloudapi.NodePoolJoinConfig {
	encode := func(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
	return acloudapi.NodePoolJoinConfig{
		Versions: acloudapi.NodeJoinConfigVersions{
			CloudInit:  "24.1",
			Kubernetes: "1.34.1",
			Containerd: "2.1.4",
			Crictl:     "1.34.0",
		},
		CloudInitUserDataBase64: encode("#cloud-config\n# token: " + bootstrapTestFakeToken + "\n"),
		InstallScriptBase64:     encode("#!/bin/bash\n# install, token: " + bootstrapTestFakeToken + "\n"),
		UpgradeScriptBase64:     encode("#!/bin/bash\n# upgrade, token: " + bootstrapTestFakeToken + "\n"),
		JoinCommand:             "kubeadm join 10.0.0.1:6443 --token " + bootstrapTestFakeToken,
		KubeletConfigBase64:     encode("apiVersion: kubelet.config.k8s.io/v1beta1\nkind: KubeletConfiguration\n"),
	}
}

// bootstrapFakeAPI mimics the platform-api endpoints used by acloud_nodepool_join_config:
//
//	GET  /api/v1/orgs/{org}/clusters/{env}/{cluster}
//	POST /api/v1/orgs/{org}/clusters/{env}/{cluster}/pools/{nodePoolId}/join-config
//
// Like the real API, {nodePoolId} is parsed as a Java long (a non-numeric value ends up as a 500 "internal server
// error"), the caller needs organisation admin permissions (403), the node pool must belong to the cluster (404) and
// the cluster must be a Bring Your Own Node cluster (400).
type bootstrapFakeAPI struct {
	t      *testing.T
	server *httptest.Server

	mu       sync.Mutex
	admin    bool
	byon     bool
	pools    map[int64]bool
	response func(w http.ResponseWriter) // overrides the join-config response when set
	requests []string
}

func newBootstrapFakeAPI(t *testing.T) *bootstrapFakeAPI {
	t.Helper()
	api := &bootstrapFakeAPI{t: t, admin: true, byon: true, pools: map[int64]bool{bootstrapTestNodePoolID: true}}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/orgs/{org}/clusters/{env}/{cluster}", func(w http.ResponseWriter, r *http.Request) {
		api.record(r)
		if !api.clusterExists(r) {
			bootstrapWriteJSON(t, w, http.StatusNotFound, acloudapi.Error{Message: "Cluster not found"})
			return
		}
		cloudProvider := "aws"
		if api.byon {
			cloudProvider = "byon"
		}
		bootstrapWriteJSON(t, w, http.StatusOK, map[string]interface{}{
			"name":            bootstrapTestCluster,
			"identity":        "6f1c1e2a-0000-4000-8000-000000000001",
			"environmentSlug": bootstrapTestEnv,
			"customerSlug":    bootstrapTestOrg,
			"slug":            bootstrapTestCluster,
			"cloudProvider":   cloudProvider,
			"status":          "running",
		})
	})
	mux.HandleFunc("POST /api/v1/orgs/{org}/clusters/{env}/{cluster}/pools/{nodePoolId}/join-config", func(w http.ResponseWriter, r *http.Request) {
		api.record(r)
		nodePoolID, err := strconv.ParseInt(r.PathValue("nodePoolId"), 10, 64)
		if err != nil {
			bootstrapWriteJSON(t, w, http.StatusInternalServerError, acloudapi.Error{Message: "internal server error"})
			return
		}
		api.mu.Lock()
		admin, byon, poolExists, response := api.admin, api.byon, api.pools[nodePoolID], api.response
		api.mu.Unlock()
		switch {
		case !admin:
			bootstrapWriteJSON(t, w, http.StatusForbidden, acloudapi.Error{Message: "user is not an admin of the organisation"})
		case !api.clusterExists(r):
			bootstrapWriteJSON(t, w, http.StatusNotFound, acloudapi.Error{Message: "Cluster not found"})
		case !byon:
			bootstrapWriteJSON(t, w, http.StatusBadRequest, acloudapi.Error{Message: "get join config is only available for clusters in a Bring-your-own-Node cloud-provider"})
		case !poolExists:
			bootstrapWriteJSON(t, w, http.StatusNotFound, acloudapi.Error{Message: "Node pool not found"})
		case response != nil:
			response(w)
		default:
			bootstrapWriteJSON(t, w, http.StatusOK, bootstrapFakeJoinConfig())
		}
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		api.record(r)
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		http.NotFound(w, r)
	})

	api.server = httptest.NewServer(mux)
	t.Cleanup(api.server.Close)
	return api
}

func (api *bootstrapFakeAPI) clusterExists(r *http.Request) bool {
	return r.PathValue("org") == bootstrapTestOrg && r.PathValue("env") == bootstrapTestEnv && r.PathValue("cluster") == bootstrapTestCluster
}

func (api *bootstrapFakeAPI) record(r *http.Request) {
	api.mu.Lock()
	defer api.mu.Unlock()
	api.requests = append(api.requests, r.Method+" "+r.URL.Path)
}

func (api *bootstrapFakeAPI) recorded() []string {
	api.mu.Lock()
	defer api.mu.Unlock()
	return append([]string(nil), api.requests...)
}

func (api *bootstrapFakeAPI) meta(organisation string) ConfiguredProvider {
	return ConfiguredProvider{
		Client:       acloudapi.NewClient(nil, acloudapi.ClientOpts{APIUrl: api.server.URL}),
		Organisation: organisation,
	}
}

func bootstrapWriteJSON(t *testing.T, w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Errorf("failed to encode response: %v", err)
	}
}

func bootstrapConfig(extra map[string]interface{}) map[string]interface{} {
	raw := map[string]interface{}{
		"environment":  bootstrapTestEnv,
		"cluster":      bootstrapTestCluster,
		"node_pool_id": strconv.FormatInt(bootstrapTestNodePoolID, 10),
	}
	for k, v := range extra {
		raw[k] = v
	}
	return raw
}

// bootstrapReadData reads the data source the way the SDK does for Terraform: diff against an empty state, then
// ReadDataApply.
func bootstrapReadData(t *testing.T, config map[string]interface{}, meta ConfiguredProvider) (*terraform.InstanceDiff, *terraform.InstanceState, diag.Diagnostics) {
	t.Helper()
	r := Provider().DataSourcesMap["acloud_nodepool_join_config"]
	diff, err := r.Diff(context.Background(), nil, terraform.NewResourceConfigRaw(config), meta)
	if err != nil {
		t.Fatalf("diff failed: %v", err)
	}
	state, diags := r.ReadDataApply(context.Background(), diff, meta)
	return diff, state, diags
}

func bootstrapAssertNoSecretsInDiags(t *testing.T, diags diag.Diagnostics) {
	t.Helper()
	for _, d := range diags {
		if strings.Contains(d.Summary+d.Detail, bootstrapTestFakeToken) {
			t.Fatalf("diagnostic contains a join credential: %+v", d)
		}
	}
}

func bootstrapDiagsText(diags diag.Diagnostics) string {
	var b strings.Builder
	for _, d := range diags {
		b.WriteString(d.Summary + "\n" + d.Detail + "\n")
	}
	return b.String()
}

func TestNodeJoinConfigSchema(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("provider schema is invalid: %v", err)
	}

	r := Provider().DataSourcesMap["acloud_nodepool_join_config"]
	if r == nil {
		t.Fatal("acloud_nodepool_join_config is not registered")
	}
	// CoreConfigSchema is the schema the SDK sends to Terraform, so this checks the flag reaches Terraform core.
	block := r.CoreConfigSchema()
	for _, name := range bootstrapSensitiveAttributes {
		attr := block.Attributes[name]
		if attr == nil {
			t.Fatalf("%s is missing from the schema", name)
		}
		if !attr.Sensitive || !attr.Computed || attr.Optional || attr.Required {
			t.Errorf("%s must be a sensitive computed attribute, got %+v", name, attr)
		}
	}
	for _, name := range []string{"id", "organisation", "environment", "cluster", "node_pool_id"} {
		if block.Attributes[name].Sensitive {
			t.Errorf("%s must not be sensitive", name)
		}
	}

	if id := r.Schema["id"]; id.Type != schema.TypeString || !id.Computed {
		t.Errorf("id must be a computed string like the ID set by d.SetId, got %+v", id)
	}
	if s := r.Schema["node_pool_id"]; s.Type != schema.TypeString || !s.Required || s.ValidateDiagFunc == nil {
		t.Errorf("node_pool_id must be a validated, required string (compatible with acloud_nodepool.id), got %+v", s)
	}
}

func TestNodeJoinConfigNodePoolIDValidation(t *testing.T) {
	p := Provider()
	for _, value := range []string{"1", "42", "9223372036854775807"} {
		t.Run("valid "+value, func(t *testing.T) {
			diags := p.ValidateDataSource("acloud_nodepool_join_config", terraform.NewResourceConfigRaw(bootstrapConfig(map[string]interface{}{"node_pool_id": value})))
			if diags.HasError() {
				t.Fatalf("expected %q to be valid, got %v", value, diags)
			}
		})
	}
	for _, value := range []string{"", "0", "-1", "abc", "1.5", " 42", "0x2a", "9223372036854775808", "6f1c1e2a-0000-4000-8000-000000000002"} {
		t.Run("invalid "+value, func(t *testing.T) {
			diags := p.ValidateDataSource("acloud_nodepool_join_config", terraform.NewResourceConfigRaw(bootstrapConfig(map[string]interface{}{"node_pool_id": value})))
			if !diags.HasError() {
				t.Fatalf("expected %q to be rejected", value)
			}
			if text := bootstrapDiagsText(diags); !strings.Contains(text, "numeric ID") {
				t.Fatalf("expected a diagnostic explaining node_pool_id is numeric, got %q", text)
			}
		})
	}
}

func TestNodeJoinConfigRead(t *testing.T) {
	tests := []struct {
		name             string
		providerOrg      string
		config           map[string]interface{}
		wantOrganisation string
	}{
		{name: "organisation from provider", providerOrg: bootstrapTestOrg, config: bootstrapConfig(nil)},
		{
			name:             "organisation from data source",
			providerOrg:      "other-org",
			config:           bootstrapConfig(map[string]interface{}{"organisation": bootstrapTestOrg}),
			wantOrganisation: bootstrapTestOrg,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := newBootstrapFakeAPI(t)
			meta := api.meta(tt.providerOrg)

			diff, state, diags := bootstrapReadData(t, tt.config, meta)
			bootstrapAssertNoSecretsInDiags(t, diags)
			if diags.HasError() {
				t.Fatalf("read failed: %v", diags)
			}

			clusterPath := "/api/v1/orgs/" + bootstrapTestOrg + "/clusters/" + bootstrapTestEnv + "/" + bootstrapTestCluster
			wantRequests := []string{"GET " + clusterPath, "POST " + clusterPath + "/pools/42/join-config"}
			if got := api.recorded(); !reflect.DeepEqual(got, wantRequests) {
				t.Fatalf("requests = %q, want %q", got, wantRequests)
			}

			for _, name := range bootstrapSensitiveAttributes {
				if attr := diff.Attributes[name]; attr == nil || !attr.NewComputed || !attr.Sensitive {
					t.Errorf("diff for %s must be computed and sensitive, got %+v", name, attr)
				}
			}

			joinConfig := bootstrapFakeJoinConfig()
			want := map[string]string{
				"id":             "42",
				"organisation":   tt.wantOrganisation,
				"environment":    bootstrapTestEnv,
				"cluster":        bootstrapTestCluster,
				"node_pool_id":   "42",
				"user_data":      joinConfig.CloudInitUserDataBase64,
				"kubelet_config": joinConfig.KubeletConfigBase64,
				"join_command":   joinConfig.JoinCommand,
				"install_script": joinConfig.InstallScriptBase64,
				"upgrade_script": joinConfig.UpgradeScriptBase64,
			}
			for key, value := range want {
				if got := state.Attributes[key]; got != value {
					t.Errorf("state %s = %q, want %q", key, got, value)
				}
			}
			if state.ID != "42" {
				t.Errorf("state ID = %q, want %q", state.ID, "42")
			}

			// The SDK converts the state to the schema type before sending it to Terraform; this failed for a
			// non-numeric id while id was declared as a number.
			r := Provider().DataSourcesMap["acloud_nodepool_join_config"]
			value, err := state.AttrsAsObjectValue(r.CoreConfigSchema().ImpliedType())
			if err != nil {
				t.Fatalf("state does not match the schema: %v", err)
			}
			if id := value.GetAttr("id"); id.Type().FriendlyName() != "string" || id.AsString() != "42" {
				t.Fatalf("id in state = %#v, want string \"42\"", id)
			}

			// A second read (the next plan) must produce the same state.
			_, secondState, diags := bootstrapReadData(t, tt.config, meta)
			if diags.HasError() {
				t.Fatalf("second read failed: %v", diags)
			}
			if !reflect.DeepEqual(state.Attributes, secondState.Attributes) {
				t.Fatalf("second read changed the state:\nfirst:  %v\nsecond: %v", state.Attributes, secondState.Attributes)
			}
		})
	}
}

func TestNodeJoinConfigReadRejectsInvalidNodePoolIDBeforeHTTP(t *testing.T) {
	for _, value := range []string{"0", "-7", "6f1c1e2a-0000-4000-8000-000000000002", ""} {
		t.Run(value, func(t *testing.T) {
			api := newBootstrapFakeAPI(t)
			// Call the read function directly, bypassing the schema validation.
			d := schema.TestResourceDataRaw(t, dataSourceNodeJoinConfig().Schema, bootstrapConfig(map[string]interface{}{"node_pool_id": value}))
			diags := dataSourceNodeJoinConfigRead(context.Background(), d, api.meta(bootstrapTestOrg))
			if !diags.HasError() || !strings.Contains(bootstrapDiagsText(diags), "numeric ID") {
				t.Fatalf("expected a node_pool_id error, got %v", diags)
			}
			if got := api.recorded(); len(got) != 0 {
				t.Fatalf("expected no HTTP requests, got %q", got)
			}
			if d.Id() != "" {
				t.Fatalf("id must not be set on error, got %q", d.Id())
			}
		})
	}
}

func TestNodeJoinConfigReadErrors(t *testing.T) {
	tests := []struct {
		name        string
		providerOrg string
		config      map[string]interface{}
		setup       func(api *bootstrapFakeAPI)
		wantText    []string
		wantHTTP    bool
	}{
		{
			name:     "organisation not set",
			config:   bootstrapConfig(nil),
			wantText: []string{"organisation is not set"},
		},
		{
			name:        "cluster not found",
			providerOrg: bootstrapTestOrg,
			config:      bootstrapConfig(map[string]interface{}{"cluster": "missing"}),
			wantText:    []string{"Failed to read cluster bootstrap-org/byon-env/missing", "404: Cluster not found", "Check that the organisation"},
			wantHTTP:    true,
		},
		{
			name:        "not an organisation admin",
			providerOrg: bootstrapTestOrg,
			config:      bootstrapConfig(nil),
			setup:       func(api *bootstrapFakeAPI) { api.admin = false },
			wantText:    []string{"node pool 42 in cluster bootstrap-org/byon-env/byon-cluster", "403:", "requires admin permissions"},
			wantHTTP:    true,
		},
		{
			name:        "node pool not in cluster",
			providerOrg: bootstrapTestOrg,
			config:      bootstrapConfig(map[string]interface{}{"node_pool_id": "43"}),
			wantText:    []string{"node pool 43", "404: Node pool not found", "numeric ID of a node pool in this cluster"},
			wantHTTP:    true,
		},
		{
			name:        "not a bring your own node cluster",
			providerOrg: bootstrapTestOrg,
			config:      bootstrapConfig(nil),
			setup:       func(api *bootstrapFakeAPI) { api.byon = false },
			wantText:    []string{"400:", "Bring Your Own Node"},
			wantHTTP:    true,
		},
		{
			name:        "empty response body",
			providerOrg: bootstrapTestOrg,
			config:      bootstrapConfig(nil),
			setup: func(api *bootstrapFakeAPI) {
				api.response = func(w http.ResponseWriter) { w.WriteHeader(http.StatusOK) }
			},
			wantText: []string{"empty join configuration for node pool 42"},
			wantHTTP: true,
		},
		{
			name:        "response without join configuration",
			providerOrg: bootstrapTestOrg,
			config:      bootstrapConfig(nil),
			setup: func(api *bootstrapFakeAPI) {
				api.response = func(w http.ResponseWriter) {
					bootstrapWriteJSON(api.t, w, http.StatusOK, map[string]interface{}{"versions": map[string]string{"kubernetes": "1.34.1"}})
				}
			},
			wantText: []string{"empty join configuration"},
			wantHTTP: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := newBootstrapFakeAPI(t)
			if tt.setup != nil {
				tt.setup(api)
			}
			_, state, diags := bootstrapReadData(t, tt.config, api.meta(tt.providerOrg))
			if !diags.HasError() {
				t.Fatalf("expected an error, got state %v", state)
			}
			bootstrapAssertNoSecretsInDiags(t, diags)
			text := bootstrapDiagsText(diags)
			for _, want := range tt.wantText {
				if !strings.Contains(text, want) {
					t.Errorf("diagnostics %q do not contain %q", text, want)
				}
			}
			if got := api.recorded(); (len(got) > 0) != tt.wantHTTP {
				t.Errorf("requests = %q, want HTTP requests: %v", got, tt.wantHTTP)
			}
			if state != nil && state.ID != "" {
				t.Errorf("no state must be stored on error, got ID %q", state.ID)
			}
		})
	}
}
