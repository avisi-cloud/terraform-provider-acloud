package acloud

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

func dataSourceNodeJoinConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNodeJoinConfigRead,
		Description: `Provides access to node join configuration for a node pool. Can be used in combination with other terraform providers to provision new Kubernetes Nodes for [Bring Your Own Node](https://docs.avisi.cloud/product/kubernetes/bring-your-own-node/) clusters in Avisi Cloud Kubernetes.

With Bring Your Own Node clusters you can retrieve [join configuration](https://docs.avisi.cloud/docs/how-to/kubernetes-bring-your-own-node/join-nodes-to-cluster/) for new nodes in the form of userdata or install scripts.
This datasource only works for Bring Your Own Node clusters, and requires admin permissions in the organisation.

~> **Sensitive outputs:** ` + "`user_data`, `kubelet_config`, `join_command`, `install_script` and `upgrade_script`" + ` contain credentials that allow a node to join the cluster (base64 encoding does not protect them). They are marked as sensitive, which hides them in plan and apply output, but they are still stored in plain text in the Terraform state. Use a state backend with encryption at rest and restrict access to the state accordingly.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The numeric ID of the node pool, same as `node_pool_id`.",
			},
			"organisation": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Slug of the Organisation. Defaults to the organisation configured in the provider.",
			},
			"environment": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug of the environment of the cluster",
			},
			"cluster": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug of the cluster",
			},
			"node_pool_id": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validateNodeJoinConfigNodePoolID),
				Description:      "Numeric ID of the node pool, for example `acloud_nodepool.example.id`. This is not the UUID `identity` of the node pool.",
			},
			"user_data": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Cloud Init user-data (base64).",
			},
			"kubelet_config": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Kubelet configuration (base64).",
			},
			"join_command": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Command for joining a node to the cluster.",
			},
			"install_script": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Install bash script for joining a node (base64).",
			},
			"upgrade_script": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Install bash script for upgrading a node (base64).",
			},
		},
	}
}

// parseNodeJoinConfigNodePoolID parses the numeric node pool ID that the platform API expects in the join-config path.
func parseNodeJoinConfigNodePoolID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("node_pool_id must be the positive numeric ID of the node pool (for example acloud_nodepool.example.id), not the UUID identity, got %q", value)
	}
	return id, nil
}

func validateNodeJoinConfigNodePoolID(value interface{}, key string) ([]string, []error) {
	s, ok := value.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected %s to be a string, got %T", key, value)}
	}
	if _, err := parseNodeJoinConfigNodePoolID(s); err != nil {
		return nil, []error{err}
	}
	return nil, nil
}

func dataSourceNodeJoinConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}

	environmentSlug := d.Get("environment").(string)
	clusterSlug := d.Get("cluster").(string)
	clusterName := fmt.Sprintf("%s/%s/%s", org, environmentSlug, clusterSlug)

	nodePoolID, err := parseNodeJoinConfigNodePoolID(d.Get("node_pool_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	cluster, err := client.GetCluster(ctx, org, environmentSlug, clusterSlug)
	if err != nil {
		return nodeJoinConfigAPIError(fmt.Sprintf("Failed to read cluster %s", clusterName), err)
	}
	if cluster == nil {
		return diag.Errorf("cluster %s was not found", clusterName)
	}
	// The join-config path is built from these slugs; use the configured ones rather than relying on the response.
	cluster.CustomerSlug = org
	cluster.EnvironmentSlug = environmentSlug
	cluster.Slug = clusterSlug

	// The platform API expects the numeric node pool ID in the join-config path. go-client formats NodePool.Identity
	// into that path, so pass the numeric ID there.
	nodePool := acloudapi.NodePool{Identity: strconv.FormatInt(nodePoolID, 10)}
	if nodePoolID <= math.MaxInt {
		nodePool.ID = int(nodePoolID)
	}
	nodeJoinConfig, err := client.GetNodePoolJoinConfig(ctx, *cluster, nodePool)
	if err != nil {
		return nodeJoinConfigAPIError(fmt.Sprintf("Failed to get join configuration for node pool %d in cluster %s", nodePoolID, clusterName), err)
	}
	if nodeJoinConfig == nil || (nodeJoinConfig.CloudInitUserDataBase64 == "" && nodeJoinConfig.InstallScriptBase64 == "" && nodeJoinConfig.JoinCommand == "") {
		return diag.Errorf("the API returned an empty join configuration for node pool %d in cluster %s", nodePoolID, clusterName)
	}

	d.SetId(strconv.FormatInt(nodePoolID, 10))
	values := map[string]string{
		"user_data":      nodeJoinConfig.CloudInitUserDataBase64,
		"kubelet_config": nodeJoinConfig.KubeletConfigBase64,
		"join_command":   nodeJoinConfig.JoinCommand,
		"install_script": nodeJoinConfig.InstallScriptBase64,
		"upgrade_script": nodeJoinConfig.UpgradeScriptBase64,
	}
	for key, value := range values {
		if err := d.Set(key, value); err != nil {
			// Do not include the value in the error, it contains credentials.
			return diag.Errorf("failed to set %s", key)
		}
	}
	return nil
}

// nodeJoinConfigAPIError turns a go-client error (formatted as "<status>: <message>") into a diagnostic with a hint
// for the common failure causes of the join-config endpoint.
func nodeJoinConfigAPIError(summary string, err error) diag.Diagnostics {
	detail := err.Error()
	var hint string
	switch {
	case strings.HasPrefix(detail, "401:"):
		hint = "Check the credentials configured for the provider."
	case strings.HasPrefix(detail, "403:"):
		hint = "Fetching join configuration requires admin permissions in the organisation."
	case strings.HasPrefix(detail, "404:"):
		hint = "Check that the organisation, environment and cluster exist, and that node_pool_id is the numeric ID of a node pool in this cluster."
	case strings.HasPrefix(detail, "400:"):
		hint = "Join configuration is only available for node pools of Bring Your Own Node clusters."
	}
	if hint != "" {
		detail = detail + "\n\n" + hint
	}
	return diag.Diagnostics{{
		Severity: diag.Error,
		Summary:  summary,
		Detail:   detail,
	}}
}
