package acloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-cty/cty/gocty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

const testNodePoolID = 7

// fakeNodePoolAPI is a minimal platform-api for one cluster with one node pool. It records the raw request bodies so
// tests can check which JSON keys the provider sends, and applies them like the API does.
type fakeNodePoolAPI struct {
	t      *testing.T
	server *httptest.Server

	mu             sync.Mutex
	bootDiskSource acloudapi.NodePoolBootDiskSource
	volumeSize     *int
	requests       int
	createBodies   []map[string]json.RawMessage
	updateBodies   []map[string]json.RawMessage
}

func newFakeNodePoolAPI(t *testing.T) *fakeNodePoolAPI {
	t.Helper()
	api := &fakeNodePoolAPI{t: t}
	clusterPath := "/api/v1/orgs/" + testOrg + "/clusters/" + testEnvironment + "/" + testClusterSlug
	poolsPath := clusterPath + "/pools"

	mux := http.NewServeMux()
	mux.HandleFunc("GET "+clusterPath, func(w http.ResponseWriter, r *http.Request) {
		api.record(r, nil)
		api.writeJSON(w, acloudapi.Cluster{Slug: testClusterSlug, EnvironmentSlug: testEnvironment, CustomerSlug: testOrg})
	})
	mux.HandleFunc("GET "+poolsPath, func(w http.ResponseWriter, r *http.Request) {
		api.record(r, nil)
		api.writeJSON(w, []acloudapi.NodePool{api.nodePool()})
	})
	mux.HandleFunc("POST "+poolsPath, func(w http.ResponseWriter, r *http.Request) {
		api.apply(api.record(r, &api.createBodies))
		api.writeJSON(w, api.nodePool())
	})
	mux.HandleFunc("PUT "+poolsPath+"/"+strconv.Itoa(testNodePoolID), func(w http.ResponseWriter, r *http.Request) {
		api.apply(api.record(r, &api.updateBodies))
		api.writeJSON(w, api.nodePool())
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		http.NotFound(w, r)
	})

	api.server = httptest.NewServer(mux)
	t.Cleanup(api.server.Close)
	return api
}

func (api *fakeNodePoolAPI) meta() ConfiguredProvider {
	return ConfiguredProvider{
		Client:       acloudapi.NewClient(nil, acloudapi.ClientOpts{APIUrl: api.server.URL}),
		Organisation: testOrg,
	}
}

func (api *fakeNodePoolAPI) record(r *http.Request, bodies *[]map[string]json.RawMessage) map[string]json.RawMessage {
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

// apply stores the sent boot disk settings; like platform-api, choosing the node type disk clears the size.
func (api *fakeNodePoolAPI) apply(body map[string]json.RawMessage) {
	api.mu.Lock()
	defer api.mu.Unlock()
	if raw, ok := body["bootDiskSource"]; ok {
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			api.t.Errorf("bootDiskSource in request body is not a string: %s", raw)
		}
		bootDiskSource, err := acloudapi.ParseNodePoolBootDiskSource(value)
		if err != nil {
			api.t.Errorf("invalid bootDiskSource in request body: %v", err)
		}
		api.bootDiskSource = bootDiskSource
		if bootDiskSource == acloudapi.NodePoolBootDiskSourceNodeTypeDisk {
			api.volumeSize = nil
		}
	}
	if raw, ok := body["volumeSize"]; ok {
		var value int
		if err := json.Unmarshal(raw, &value); err != nil {
			api.t.Errorf("volumeSize in request body is not a number: %s", raw)
		}
		api.volumeSize = &value
	}
}

func (api *fakeNodePoolAPI) nodePool() acloudapi.NodePool {
	api.mu.Lock()
	defer api.mu.Unlock()
	return acloudapi.NodePool{
		ID:             testNodePoolID,
		Identity:       "nodepool-uuid",
		Name:           "workers",
		NodeSize:       "s5.small",
		MinSize:        1,
		MaxSize:        1,
		BootDiskSource: api.bootDiskSource,
		VolumeSize:     api.volumeSize,
	}
}

func (api *fakeNodePoolAPI) writeJSON(w http.ResponseWriter, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		api.t.Errorf("failed to encode response: %v", err)
	}
}

func nodepoolConfig(extra map[string]interface{}) map[string]interface{} {
	raw := map[string]interface{}{
		"environment": testEnvironment,
		"cluster":     testClusterSlug,
		"name":        "workers",
		"node_size":   "s5.small",
		// Create fails without an upgrade strategy, see resourceNodepoolCreate.
		"upgrade_strategy": string(acloudapi.NodePoolUpgradeStrategyReplace),
	}
	for k, v := range extra {
		raw[k] = v
	}
	return raw
}

// nodepoolRawConfig builds the config as the cty value Terraform sends, which the SDK puts in RawConfig during a plan.
func nodepoolRawConfig(t *testing.T, config map[string]interface{}) cty.Value {
	t.Helper()
	attributes := map[string]cty.Value{}
	for name, attributeType := range resourceNodepool().CoreConfigSchema().ImpliedType().AttributeTypes() {
		value, ok := config[name]
		if !ok {
			attributes[name] = cty.NullVal(attributeType)
			continue
		}
		ctyValue, err := gocty.ToCtyValue(value, attributeType)
		if err != nil {
			t.Fatalf("cannot convert %s=%v: %v", name, value, err)
		}
		attributes[name] = ctyValue
	}
	return cty.ObjectVal(attributes)
}

// nodepoolState returns the state of an existing node pool. An empty bootDiskSource leaves both boot disk attributes
// out, like state written by a provider version without them.
func nodepoolState(t *testing.T, bootDiskSource acloudapi.NodePoolBootDiskSource, volumeSize int) *terraform.InstanceState {
	t.Helper()
	d := schema.TestResourceDataRaw(t, resourceNodepool().Schema, nodepoolConfig(nil))
	d.SetId(strconv.Itoa(testNodePoolID))
	d.Set("identity", "nodepool-uuid")
	state := d.State()
	delete(state.Attributes, "boot_disk_source")
	delete(state.Attributes, "volume_size")
	if bootDiskSource != "" {
		state.Attributes["boot_disk_source"] = string(bootDiskSource)
		state.Attributes["volume_size"] = strconv.Itoa(volumeSize)
	}
	return state
}

// planNodepool plans config against state the way Terraform does, including RawConfig.
func planNodepool(t *testing.T, state *terraform.InstanceState, config map[string]interface{}, meta interface{}) (*terraform.InstanceDiff, error) {
	t.Helper()
	state.RawConfig = nodepoolRawConfig(t, config)
	return resourceNodepool().Diff(context.Background(), state, terraform.NewResourceConfigRaw(config), meta)
}

func intPointer(value int) *int {
	return &value
}

func TestResourceNodepoolBootDiskSchema(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("provider schema is invalid: %v", err)
	}

	for name, wantType := range map[string]schema.ValueType{"boot_disk_source": schema.TypeString, "volume_size": schema.TypeInt} {
		s := resourceNodepool().Schema[name]
		if s == nil {
			t.Fatalf("%s is missing from the acloud_nodepool resource schema", name)
		}
		if s.Type != wantType || !s.Optional || !s.Computed || s.ForceNew || s.Default != nil {
			t.Fatalf("%s must be an optional, computed, in-place updatable attribute without default, got %+v", name, s)
		}

		ds := dataSourceNodepool().Schema[name]
		if ds == nil {
			t.Fatalf("%s is missing from the acloud_nodepool data source schema", name)
		}
		if ds.Type != wantType || !ds.Computed || ds.Optional || ds.Required {
			t.Fatalf("data source %s must be computed, got %+v", name, ds)
		}
	}
}

func TestResourceNodepoolBootDiskValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr string
	}{
		{name: "unknown boot_disk_source", config: map[string]interface{}{"boot_disk_source": "NETWORK_VOLUME"}, wantErr: "boot_disk_source"},
		{name: "uppercase boot_disk_source", config: map[string]interface{}{"boot_disk_source": "VOLUME"}, wantErr: "boot_disk_source"},
		{name: "empty boot_disk_source", config: map[string]interface{}{"boot_disk_source": ""}, wantErr: "boot_disk_source"},
		{name: "zero volume_size", config: map[string]interface{}{"volume_size": 0}, wantErr: "volume_size"},
		{name: "negative volume_size", config: map[string]interface{}{"volume_size": -5}, wantErr: "volume_size"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := newFakeNodePoolAPI(t)
			diags := Provider().ValidateResource("acloud_nodepool", terraform.NewResourceConfigRaw(nodepoolConfig(tt.config)))
			if !diags.HasError() {
				t.Fatalf("expected %v to be rejected", tt.config)
			}
			if got := fmt.Sprint(diags); !strings.Contains(got, tt.wantErr) {
				t.Fatalf("expected an error about %s, got %v", tt.wantErr, got)
			}
			if api.requests != 0 {
				t.Fatalf("expected no API calls during validation, got %d", api.requests)
			}
		})
	}

	for _, value := range acloudapi.AllNodePoolBootDiskSources {
		t.Run("valid "+string(value), func(t *testing.T) {
			diags := Provider().ValidateResource("acloud_nodepool", terraform.NewResourceConfigRaw(nodepoolConfig(map[string]interface{}{"boot_disk_source": string(value)})))
			if diags.HasError() {
				t.Fatalf("expected %q to be valid, got %v", value, diags)
			}
		})
	}
}

func TestResourceNodepoolVolumeSizeWithNodeTypeDiskIsRefused(t *testing.T) {
	config := nodepoolConfig(map[string]interface{}{"boot_disk_source": "node-type-disk", "volume_size": 80})
	for name, state := range map[string]*terraform.InstanceState{
		"new node pool":      {},
		"existing node pool": nodepoolState(t, acloudapi.NodePoolBootDiskSourceVolume, 80),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := planNodepool(t, state, config, nil)
			if err == nil || !strings.Contains(err.Error(), `volume_size can not be used with boot_disk_source "node-type-disk"`) {
				t.Fatalf("expected the combination to be refused, got %v", err)
			}
		})
	}
}

func TestResourceNodepoolCreateBootDisk(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]interface{}
		// wantBootDiskSource and wantVolumeSize are the expected JSON values; empty means the key must be left out.
		wantBootDiskSource string
		wantVolumeSize     string
	}{
		{name: "not configured leaves both out", config: nodepoolConfig(nil)},
		{name: "node-type-disk is sent without a size", config: nodepoolConfig(map[string]interface{}{"boot_disk_source": "node-type-disk"}), wantBootDiskSource: `"node-type-disk"`},
		{name: "volume with a size sends both", config: nodepoolConfig(map[string]interface{}{"boot_disk_source": "volume", "volume_size": 80}), wantBootDiskSource: `"volume"`, wantVolumeSize: "80"},
		{name: "only a size sends only the size", config: nodepoolConfig(map[string]interface{}{"volume_size": 120}), wantVolumeSize: "120"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := newFakeNodePoolAPI(t)
			d := schema.TestResourceDataRaw(t, resourceNodepool().Schema, tt.config)

			if diags := resourceNodepoolCreate(context.Background(), d, api.meta()); diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}
			if len(api.createBodies) != 1 {
				t.Fatalf("expected 1 create request, got %d", len(api.createBodies))
			}
			assertJSONKey(t, api.createBodies[0], "bootDiskSource", tt.wantBootDiskSource)
			assertJSONKey(t, api.createBodies[0], "volumeSize", tt.wantVolumeSize)
		})
	}
}

func TestResourceNodepoolReadBootDisk(t *testing.T) {
	tests := []struct {
		name               string
		bootDiskSource     acloudapi.NodePoolBootDiskSource
		volumeSize         *int
		wantBootDiskSource string
		wantVolumeSize     int
	}{
		{name: "node type disk", bootDiskSource: acloudapi.NodePoolBootDiskSourceNodeTypeDisk, wantBootDiskSource: "node-type-disk"},
		{name: "volume with a size", bootDiskSource: acloudapi.NodePoolBootDiskSourceVolume, volumeSize: intPointer(80), wantBootDiskSource: "volume", wantVolumeSize: 80},
		{name: "volume with the default size", bootDiskSource: acloudapi.NodePoolBootDiskSourceVolume, wantBootDiskSource: "volume"},
		{name: "provider without boot disk choices"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := newFakeNodePoolAPI(t)
			api.bootDiskSource = tt.bootDiskSource
			api.volumeSize = tt.volumeSize

			d := resourceNodepool().Data(nodepoolState(t, "", 0))
			if diags := resourceNodepoolRead(context.Background(), d, api.meta()); diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}
			ds := schema.TestResourceDataRaw(t, dataSourceNodepool().Schema, map[string]interface{}{
				"id":          strconv.Itoa(testNodePoolID),
				"environment": testEnvironment,
				"cluster":     testClusterSlug,
			})
			if diags := dataNodepoolRead(context.Background(), ds, api.meta()); diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}

			for kind, data := range map[string]*schema.ResourceData{"resource": d, "data source": ds} {
				if got := data.Get("boot_disk_source").(string); got != tt.wantBootDiskSource {
					t.Fatalf("%s boot_disk_source = %q, want %q", kind, got, tt.wantBootDiskSource)
				}
				if got := data.Get("volume_size").(int); got != tt.wantVolumeSize {
					t.Fatalf("%s volume_size = %d, want %d", kind, got, tt.wantVolumeSize)
				}
			}
		})
	}
}

func TestResourceNodepoolUpdateBootDisk(t *testing.T) {
	tests := []struct {
		name string
		// stateSource and stateSize are on the API and in state.
		stateSource acloudapi.NodePoolBootDiskSource
		stateSize   int
		config      map[string]interface{}
		// wantBootDiskSource and wantVolumeSize are the expected JSON values; empty means the key must be left out.
		wantBootDiskSource string
		wantVolumeSize     string
		wantStateSource    string
		wantStateSize      int
	}{
		{
			name:               "switch to the node type disk sends only the source",
			stateSource:        acloudapi.NodePoolBootDiskSourceVolume,
			stateSize:          80,
			config:             nodepoolConfig(map[string]interface{}{"boot_disk_source": "node-type-disk"}),
			wantBootDiskSource: `"node-type-disk"`,
			wantStateSource:    "node-type-disk",
			wantStateSize:      0,
		},
		{
			name:            "changed size is sent alone",
			stateSource:     acloudapi.NodePoolBootDiskSourceVolume,
			stateSize:       80,
			config:          nodepoolConfig(map[string]interface{}{"boot_disk_source": "volume", "volume_size": 100}),
			wantVolumeSize:  "100",
			wantStateSource: "volume",
			wantStateSize:   100,
		},
		{
			name:               "switch to a volume with a size sends both",
			stateSource:        acloudapi.NodePoolBootDiskSourceNodeTypeDisk,
			config:             nodepoolConfig(map[string]interface{}{"boot_disk_source": "volume", "volume_size": 120}),
			wantBootDiskSource: `"volume"`,
			wantVolumeSize:     "120",
			wantStateSource:    "volume",
			wantStateSize:      120,
		},
		{
			name:            "unchanged configured values are left out",
			stateSource:     acloudapi.NodePoolBootDiskSourceVolume,
			stateSize:       80,
			config:          nodepoolConfig(map[string]interface{}{"boot_disk_source": "volume", "volume_size": 80, "min_size": 2, "max_size": 2}),
			wantStateSource: "volume",
			wantStateSize:   80,
		},
		{
			name:            "not configured values are left out",
			stateSource:     acloudapi.NodePoolBootDiskSourceNodeTypeDisk,
			config:          nodepoolConfig(map[string]interface{}{"min_size": 2, "max_size": 2}),
			wantStateSource: "node-type-disk",
			wantStateSize:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := resourceNodepool()
			api := newFakeNodePoolAPI(t)
			api.bootDiskSource = tt.stateSource
			if tt.stateSize != 0 {
				api.volumeSize = intPointer(tt.stateSize)
			}
			meta := api.meta()

			state := nodepoolState(t, tt.stateSource, tt.stateSize)
			diff, err := planNodepool(t, state, tt.config, meta)
			if err != nil {
				t.Fatalf("unexpected plan error: %v", err)
			}
			if diff == nil || diff.Empty() {
				t.Fatal("expected a diff")
			}
			if diff.RequiresNew() {
				t.Fatalf("expected an in-place update, got a replacement: %v", diff)
			}

			newState, diags := r.Apply(context.Background(), state, diff, meta)
			if diags.HasError() {
				t.Fatalf("unexpected apply error: %v", diags)
			}
			if len(api.updateBodies) != 1 {
				t.Fatalf("expected 1 update request, got %d", len(api.updateBodies))
			}
			assertJSONKey(t, api.updateBodies[0], "bootDiskSource", tt.wantBootDiskSource)
			assertJSONKey(t, api.updateBodies[0], "volumeSize", tt.wantVolumeSize)

			if got := newState.Attributes["boot_disk_source"]; got != tt.wantStateSource {
				t.Fatalf("boot_disk_source in state after update = %q, want %q", got, tt.wantStateSource)
			}
			if got := newState.Attributes["volume_size"]; got != strconv.Itoa(tt.wantStateSize) {
				t.Fatalf("volume_size in state after update = %q, want %d", got, tt.wantStateSize)
			}
		})
	}
}

func TestResourceNodepoolSwitchToNodeTypeDiskPlansVolumeSizeAsUnknown(t *testing.T) {
	state := nodepoolState(t, acloudapi.NodePoolBootDiskSourceVolume, 80)
	diff, err := planNodepool(t, state, nodepoolConfig(map[string]interface{}{"boot_disk_source": "node-type-disk"}), nil)
	if err != nil {
		t.Fatalf("unexpected plan error: %v", err)
	}
	attr := diff.Attributes["volume_size"]
	if attr == nil || !attr.NewComputed {
		t.Fatalf("expected volume_size to be known after apply, got %#v", attr)
	}
}

func TestResourceNodepoolBootDiskNotConfiguredDiff(t *testing.T) {
	tests := map[string]struct {
		source acloudapi.NodePoolBootDiskSource
		size   int
	}{
		"node type disk":            {source: acloudapi.NodePoolBootDiskSourceNodeTypeDisk},
		"volume with a size":        {source: acloudapi.NodePoolBootDiskSourceVolume, size: 80},
		"volume with the default":   {source: acloudapi.NodePoolBootDiskSourceVolume},
		"state without the setting": {},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			diff, err := planNodepool(t, nodepoolState(t, tt.source, tt.size), nodepoolConfig(nil), nil)
			if err != nil {
				t.Fatalf("unexpected plan error: %v", err)
			}
			if diff == nil {
				return
			}
			for _, attribute := range []string{"boot_disk_source", "volume_size"} {
				if attr := diff.Attributes[attribute]; attr != nil && (attr.Old != attr.New || attr.NewComputed) {
					t.Fatalf("expected no diff for %s when it is not configured, got %#v", attribute, attr)
				}
			}
		})
	}
}

func assertJSONKey(t *testing.T, body map[string]json.RawMessage, key, want string) {
	t.Helper()
	raw, sent := body[key]
	switch {
	case want == "" && sent:
		t.Fatalf("request contains %s %s, want it left out", key, raw)
	case want != "" && !sent:
		t.Fatalf("request does not contain %s, want %s", key, want)
	case want != "" && string(raw) != want:
		t.Fatalf("%s in request = %s, want %s", key, raw, want)
	}
}
