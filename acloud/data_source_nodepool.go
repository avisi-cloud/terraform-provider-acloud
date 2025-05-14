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

func dataSourceNodepool() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataNodepoolRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organisation": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Slug of the Organisation.",
			},
			"environment": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Slug of the Environment.",
			},
			"cluster": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Slug of the Cluster.",
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
				Computed:    true,
				Description: "Name of the Node Pool",
			},
			"availability_zone": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Availability Zone in which the nodes was provisioned.",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func dataNodepoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	return nil
}
