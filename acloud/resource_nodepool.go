package acloud

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/exp/slices"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

func resourceNodepool() *schema.Resource {
	return &schema.Resource{
		Description:   "Create a node pool for a cluster",
		CreateContext: resourceNodepoolCreate,
		ReadContext:   resourceNodepoolRead,
		UpdateContext: resourceNodepoolUpdate,
		DeleteContext: resourceNodepoolDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organisation": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Slug of the Organisation. Can only be set on creation.",
			},
			"environment": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Slug of the Environment. Can only be set on creation.",
			},
			"cluster": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Slug of the Cluster. Can only be set on creation.",
			},
			"organisation_slug": {
				Type:       schema.TypeString,
				Deprecated: "replaced by organisation",
				Optional:   true,
				Default:    nil,
			},
			"environment_slug": {
				Type:       schema.TypeString,
				Deprecated: "replaced by environment",
				Optional:   true,
				Default:    nil,
			},
			"cluster_slug": {
				Type:       schema.TypeString,
				Deprecated: "replaced by cluster",
				Optional:   true,
				Default:    nil,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Node Pool",
			},
			"availability_zone": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Availability Zone in which the nodes will be provisioned. Can only be set on creation.",
			},
			"node_size": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Type of machines in the Node Pool",
			},
			"node_count": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Number of nodes in the Node Pool. Used when auto_scaling is set to `false`.",
			},
			"auto_scaling": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enables auto scaling of the Node Pool when set to `true`",
			},
			"min_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Minimum amount of nodes in the Node Pool. Used when auto_scaling is set to `true`.",
			},
			"max_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Maximum amount of nodes in the Node Pool. Used when auto_scaling is set to `true`.",
			},
			"node_auto_replacement": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"annotations": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Annotations to put on the nodes in the Node Pool",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"labels": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Labels to put on the nodes in the Node Pool",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"taints": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Taints to put on the nodes in the Node Pool",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"effect": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceNodepoolCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client

	cluster, err := getClusterForNodePool(ctx, d, m)
	if err != nil {
		return diag.FromErr(fmt.Errorf("cluster was not found: %w", err))
	}

	nodeCount := d.Get("node_count").(int)
	minNodePoolCount := d.Get("min_size").(int)
	maxNodePoolCount := d.Get("max_size").(int)

	autoScaling := d.Get("auto_scaling").(bool)
	if !autoScaling {
		minNodePoolCount = nodeCount
		maxNodePoolCount = nodeCount
	}

	createNodepool := acloudapi.CreateNodePool{
		Name:     d.Get("name").(string),
		NodeSize: d.Get("node_size").(string),
		// TODO: not yet supported by the API
		// NodeCount: nodeCount,
		MinSize:             minNodePoolCount,
		MaxSize:             maxNodePoolCount,
		AvailabilityZone:    d.Get("availability_zone").(string),
		Annotations:         castInterfaceMap(d.Get("annotations").(map[string]interface{})),
		Labels:              castInterfaceMap(d.Get("labels").(map[string]interface{})),
		Taints:              castNodeTaints(d.Get("taints").([]interface{})),
		NodeAutoReplacement: d.Get("node_auto_replacement").(bool),
	}

	nodePool, err := client.CreateNodePool(ctx, *cluster, createNodepool)

	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create node pool: %w", err))
	}

	if nodePool != nil {
		d.SetId(strconv.Itoa(nodePool.ID))
		return nil
	}

	return resourceNodepoolRead(ctx, d, m)

}

func castNodeTaints(taints []interface{}) []acloudapi.NodeTaint {

	result := []acloudapi.NodeTaint{}

	for _, taint := range taints {
		t := taint.(map[string]interface{})

		newTaint := acloudapi.NodeTaint{
			Key:    t["key"].(string),
			Value:  t["value"].(string),
			Effect: t["effect"].(string),
		}
		result = append(result, newTaint)
	}

	return result
}

func castInterfaceMap(original map[string]interface{}) map[string]string {
	result := make(map[string]string)

	for key, value := range original {
		if str, ok := value.(string); ok {
			result[key] = str
		}
	}

	return result
}

func getClusterForNodePool(ctx context.Context, d *schema.ResourceData, m interface{}) (*acloudapi.Cluster, error) {
	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return nil, err
	}

	env := getStringAttributeWithLegacyName(d, "environment", "environment_slug")
	cls := getStringAttributeWithLegacyName(d, "cluster", "cluster_slug")

	return client.GetCluster(ctx, org, env, cls)
}

func resourceNodepoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client

	cluster, err := getClusterForNodePool(ctx, d, m)
	if err != nil {
		return diag.FromErr(fmt.Errorf("cluster was not found: %w", err))
	}

	nodePools, err := client.GetNodePoolsByCluster(ctx, *cluster)

	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to find node pool: %w", err))
	}

	nodePoolID, _ := strconv.Atoi(d.Get("id").(string))

	idx := slices.IndexFunc(nodePools, func(pool acloudapi.NodePool) bool {
		return pool.ID == nodePoolID
	})

	if idx == -1 {
		return diag.FromErr(fmt.Errorf("nodepool was not found"))
	}

	nodePool := nodePools[idx]

	d.SetId(strconv.Itoa(nodePool.ID))
	d.Set("name", nodePool.Name)
	d.Set("node_size", nodePool.NodeSize)
	d.Set("auto_scaling", nodePool.AutoScaling)
	d.Set("availability_zone", nodePool.AvailabilityZone)
	d.Set("node_auto_replacement", nodePool.NodeAutoReplacement)
	d.Set("min_size", nodePool.MinSize)
	d.Set("max_size", nodePool.MaxSize)
	d.Set("annotations", nodePool.Annotations)
	d.Set("labels", nodePool.Labels)
	d.Set("taints", nodePool.Taints)
	return nil
}

func resourceNodepoolUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client

	cluster, err := getClusterForNodePool(ctx, d, m)
	if err != nil {
		return diag.FromErr(fmt.Errorf("cluster was not found: %w", err))
	}

	nodePoolID, _ := strconv.Atoi(d.Get("id").(string))

	updateNodepool := acloudapi.CreateNodePool{
		NodeSize:    d.Get("node_size").(string),
		MinSize:     d.Get("min_size").(int),
		MaxSize:     d.Get("max_size").(int),
		Annotations: castInterfaceMap(d.Get("annotations").(map[string]interface{})),
		Labels:      castInterfaceMap(d.Get("labels").(map[string]interface{})),
		Taints:      castNodeTaints(d.Get("taints").([]interface{})),
	}

	nodePool, err := client.UpdateNodePool(ctx, *cluster, nodePoolID, updateNodepool)

	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update node pool: %w", err))
	}

	if nodePool != nil {
		d.Set("node_size", nodePool.NodeSize)
		d.Set("min_size", nodePool.MinSize)
		d.Set("max_size", nodePool.MaxSize)
		d.Set("annotations", nodePool.Annotations)
		d.Set("labels", nodePool.Labels)
		return nil
	}

	return resourceNodepoolRead(ctx, d, m)
}

func resourceNodepoolDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client

	cluster, err := getClusterForNodePool(ctx, d, m)
	if err != nil {
		return diag.FromErr(fmt.Errorf("cluster was not found: %w", err))
	}

	nodePoolID, _ := strconv.Atoi(d.Get("id").(string))

	err = client.DeleteNodePool(ctx, *cluster, nodePoolID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete node pool: %w", err))
	}

	d.SetId("")

	return nil
}
