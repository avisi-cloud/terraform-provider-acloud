package acloud

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
		CustomizeDiff: customizeNodepoolBootDiskDiff,
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
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Auto healing for nodes within this node pool",
			},
			"upgrade_strategy": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Specify the upgrade strategy for nodes in this pool",
				ValidateFunc: validation.StringInSlice([]string{
					string(acloudapi.NodePoolUpgradeStrategyReplace),
					string(acloudapi.NodePoolUpgradeStrategyInPlace),
					string(acloudapi.NodePoolUpgradeStrategyInPlaceWithoutDrain),
					string(acloudapi.NodePoolUpgradeStrategyReplaceMinorInPlacePatch),
					string(acloudapi.NodePoolUpgradeStrategyReplaceMinorInPlacePatchNoDrain),
				}, false),
			},
			"security_updates_on_join": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     string(acloudapi.NodePoolSecurityUpdatesOnJoinOff),
				Description: "Install OS security updates during node bring-up, before the node joins the cluster. Applies to the first join only, not to upgrades of existing nodes.",
				ValidateFunc: validation.StringInSlice([]string{
					string(acloudapi.NodePoolSecurityUpdatesOnJoinOff),
					string(acloudapi.NodePoolSecurityUpdatesOnJoinInstall),
					string(acloudapi.NodePoolSecurityUpdatesOnJoinInstallAndReboot),
				}, false),
			},
			"boot_disk_source": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  bootDiskSourceDescription + " When not set, the cloud provider's base source is used, and removing the attribute keeps the current value.",
				ValidateFunc: validation.StringInSlice(bootDiskSourceValues(), false),
			},
			"volume_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  volumeSizeDescription + " Leave it out for the default size: the API stores the default size as not set, so configuring the default size itself (50) shows a diff on every plan. Can not be combined with `boot_disk_source = \"node-type-disk\"`.",
				ValidateFunc: validation.IntAtLeast(1),
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

const bootDiskSourceDescription = "What new nodes boot from. " +
	"`volume`: a separate volume of `volume_size` GB. " +
	"`node-type-disk`: the disk that comes with the node type, without a separate volume; the cloud provider must allow it. " +
	"Only supported on OpenStack. Applies only to nodes created after the change."

const volumeSizeDescription = "Size in GB of the volume new nodes boot from, when `boot_disk_source` is `volume`. " +
	"`0` means the default size of the cloud provider. The cloud provider must allow a custom volume size. " +
	"Applies only to nodes created after the change."

func bootDiskSourceValues() []string {
	values := make([]string, 0, len(acloudapi.AllNodePoolBootDiskSources))
	for _, bootDiskSource := range acloudapi.AllNodePoolBootDiskSources {
		values = append(values, string(bootDiskSource))
	}
	return values
}

// customizeNodepoolBootDiskDiff refuses a volume size with the node type disk, because the API would silently drop it.
func customizeNodepoolBootDiskDiff(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	config := d.GetRawConfig()
	if config.IsNull() {
		return nil
	}
	bootDiskSource := config.GetAttr("boot_disk_source")
	volumeSizeConfigured := !config.GetAttr("volume_size").IsNull()
	if volumeSizeConfigured && bootDiskSource.IsKnown() && !bootDiskSource.IsNull() && bootDiskSource.AsString() == string(acloudapi.NodePoolBootDiskSourceNodeTypeDisk) {
		return fmt.Errorf("volume_size can not be used with boot_disk_source %q", acloudapi.NodePoolBootDiskSourceNodeTypeDisk)
	}
	// A new source without a configured size gets its size from the API, e.g. none for the node type disk.
	if d.Id() != "" && d.HasChange("boot_disk_source") && !volumeSizeConfigured {
		return d.SetNewComputed("volume_size")
	}
	return nil
}

func setNodepoolBootDisk(d *schema.ResourceData, nodePool acloudapi.NodePool) {
	d.Set("boot_disk_source", string(nodePool.BootDiskSource))
	volumeSize := 0
	if nodePool.VolumeSize != nil {
		volumeSize = *nodePool.VolumeSize
	}
	d.Set("volume_size", volumeSize)
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

	upgradeStrategy, err := acloudapi.ParseNodePoolUpgradeStrategy(d.Get("upgrade_strategy").(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("cannot parse upgradeStrategy: %w", err))
	}

	securityUpdatesOnJoin, err := acloudapi.ParseNodePoolSecurityUpdatesOnJoin(d.Get("security_updates_on_join").(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("cannot parse securityUpdatesOnJoin: %w", err))
	}

	// Not configured is "" or 0 (the validation refuses both as values), and then left out: the API picks the defaults.
	var bootDiskSource acloudapi.NodePoolBootDiskSource
	if value := d.Get("boot_disk_source").(string); value != "" {
		bootDiskSource, err = acloudapi.ParseNodePoolBootDiskSource(value)
		if err != nil {
			return diag.FromErr(fmt.Errorf("cannot parse bootDiskSource: %w", err))
		}
	}
	var volumeSize *int
	if size := d.Get("volume_size").(int); size != 0 {
		volumeSize = &size
	}

	createNodepool := acloudapi.CreateNodePool{
		Name:     d.Get("name").(string),
		NodeSize: d.Get("node_size").(string),
		// TODO: not yet supported by the API
		// NodeCount: nodeCount,
		AutoScaling:           autoScaling,
		MinSize:               minNodePoolCount,
		MaxSize:               maxNodePoolCount,
		AvailabilityZone:      d.Get("availability_zone").(string),
		Annotations:           castInterfaceMap(d.Get("annotations").(map[string]interface{})),
		Labels:                castInterfaceMap(d.Get("labels").(map[string]interface{})),
		Taints:                castNodeTaints(d.Get("taints").([]interface{})),
		NodeAutoReplacement:   d.Get("node_auto_replacement").(bool),
		UpgradeStrategy:       upgradeStrategy,
		SecurityUpdatesOnJoin: securityUpdatesOnJoin,
		BootDiskSource:        bootDiskSource,
		VolumeSize:            volumeSize,
	}

	nodePool, err := client.CreateNodePool(ctx, *cluster, createNodepool)

	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create node pool: %w", err))
	}

	if nodePool != nil {
		d.SetId(strconv.Itoa(nodePool.ID))
		setNodepoolBootDisk(d, *nodePool)
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

	annotations := nodePool.Annotations
	if annotations == nil {
		annotations = map[string]string{}
	}
	labels := nodePool.Labels
	if labels == nil {
		labels = map[string]string{}
	}

	d.SetId(strconv.Itoa(nodePool.ID))
	d.Set("identity", nodePool.Identity)
	d.Set("name", nodePool.Name)
	d.Set("node_size", nodePool.NodeSize)
	d.Set("auto_scaling", nodePool.AutoScaling)
	d.Set("availability_zone", nodePool.AvailabilityZone)
	d.Set("node_auto_replacement", nodePool.NodeAutoReplacement)
	d.Set("min_size", nodePool.MinSize)
	d.Set("max_size", nodePool.MaxSize)
	d.Set("annotations", annotations)
	d.Set("labels", labels)
	d.Set("taints", nodePool.Taints)
	d.Set("security_updates_on_join", string(nodePool.SecurityUpdatesOnJoin))
	setNodepoolBootDisk(d, nodePool)
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

	securityUpdatesOnJoin, err := acloudapi.ParseNodePoolSecurityUpdatesOnJoin(d.Get("security_updates_on_join").(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("cannot parse securityUpdatesOnJoin: %w", err))
	}

	// Only send a changed boot disk setting; left out, the API keeps the stored value. A size that is "known after apply" reads as 0.
	var bootDiskSource acloudapi.NodePoolBootDiskSource
	if value := d.Get("boot_disk_source").(string); d.HasChange("boot_disk_source") && value != "" {
		bootDiskSource, err = acloudapi.ParseNodePoolBootDiskSource(value)
		if err != nil {
			return diag.FromErr(fmt.Errorf("cannot parse bootDiskSource: %w", err))
		}
	}
	var volumeSize *int
	if size := d.Get("volume_size").(int); d.HasChange("volume_size") && size != 0 {
		volumeSize = &size
	}

	updateNodepool := acloudapi.CreateNodePool{
		NodeSize:              d.Get("node_size").(string),
		MinSize:               d.Get("min_size").(int),
		MaxSize:               d.Get("max_size").(int),
		Annotations:           castInterfaceMap(d.Get("annotations").(map[string]interface{})),
		Labels:                castInterfaceMap(d.Get("labels").(map[string]interface{})),
		Taints:                castNodeTaints(d.Get("taints").([]interface{})),
		SecurityUpdatesOnJoin: securityUpdatesOnJoin,
		BootDiskSource:        bootDiskSource,
		VolumeSize:            volumeSize,
	}

	nodePool, err := client.UpdateNodePool(ctx, *cluster, nodePoolID, updateNodepool)

	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update node pool: %w", err))
	}

	if nodePool != nil {
		d.Set("node_size", nodePool.NodeSize)
		d.Set("auto_scaling", nodePool.AutoScaling)
		d.Set("min_size", nodePool.MinSize)
		d.Set("max_size", nodePool.MaxSize)
		d.Set("annotations", nodePool.Annotations)
		d.Set("labels", nodePool.Labels)
		d.Set("security_updates_on_join", string(nodePool.SecurityUpdatesOnJoin))
		setNodepoolBootDisk(d, *nodePool)
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
