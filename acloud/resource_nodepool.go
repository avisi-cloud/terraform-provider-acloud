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
		CreateContext: resourceNodepoolCreate,
		ReadContext:   resourceNodepoolRead,
		UpdateContext: resourceNodepoolUpdate,
		DeleteContext: resourceNodepoolDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organisation_slug": {
				Type:     schema.TypeString,
				Required: true,
			},
			"environment_slug": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cluster_slug": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"node_size": {
				Type:     schema.TypeString,
				Required: true,
			},
			"auto_scaling": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"min_size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"max_size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"annotations": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"taints": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
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
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceNodepoolCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)
	var diags diag.Diagnostics

	cluster := getCluster(ctx, d, m)

	if cluster == nil {
		return diag.FromErr(fmt.Errorf("cluster was not found"))
	}

	createNodepool := acloudapi.CreateNodePool{
		Name:        d.Get("name").(string),
		NodeSize:    d.Get("node_size").(string),
		MinSize:     d.Get("min_size").(int),
		MaxSize:     d.Get("max_size").(int),
		Annotations: castInterfaceMap(d.Get("annotations").(map[string]interface{})),
		Labels:      castInterfaceMap(d.Get("labels").(map[string]interface{})),
		Taints:      castNodeTaints(d.Get("taints").([]interface{})),
	}

	nodePool, err := client.CreateNodePool(ctx, *cluster, createNodepool)

	if err != nil {
		return diag.FromErr(err)
	}

	if nodePool != nil {
		d.SetId(strconv.Itoa(nodePool.ID))
		return diags
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

func getCluster(ctx context.Context, d *schema.ResourceData, m interface{}) *acloudapi.Cluster {
	client := m.(acloudapi.Client)

	org := d.Get("organisation_slug").(string)
	env := d.Get("environment_slug").(string)
	cls := d.Get("cluster_slug").(string)

	cluster, _ := client.GetCluster(ctx, org, env, cls)

	if cluster != nil {
		return cluster
	}

	return nil
}

func resourceNodepoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)
	var diags diag.Diagnostics

	cluster := getCluster(ctx, d, m)

	if cluster == nil {
		return diag.FromErr(fmt.Errorf("cluster was not found"))
	}

	nodePools, err := client.GetNodePoolsByCluster(ctx, *cluster)

	if err != nil {
		return diag.FromErr(err)
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
	d.Set("min_size", nodePool.MinSize)
	d.Set("max_size", nodePool.MaxSize)
	d.Set("annotations", nodePool.Annotations)
	d.Set("labels", nodePool.Labels)
	d.Set("taints", nodePool.Taints)

	return diags
}

func resourceNodepoolUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)
	var diags diag.Diagnostics

	cluster := getCluster(ctx, d, m)

	if cluster == nil {
		return diag.FromErr(fmt.Errorf("cluster was not found"))
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
		return diag.FromErr(err)
	}

	if nodePool != nil {
		d.Set("node_size", nodePool.NodeSize)
		d.Set("min_size", nodePool.MinSize)
		d.Set("max_size", nodePool.MaxSize)
		d.Set("annotations", nodePool.Annotations)
		d.Set("labels", nodePool.Labels)
		return diags
	}

	return resourceNodepoolRead(ctx, d, m)
}

func resourceNodepoolDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)
	var diags diag.Diagnostics

	cluster := getCluster(ctx, d, m)

	if cluster == nil {
		return diag.FromErr(fmt.Errorf("cluster was not found"))
	}

	nodePoolID, _ := strconv.Atoi(d.Get("id").(string))

	err := client.DeleteNodePool(ctx, *cluster, nodePoolID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
