package acloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

func resourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"organisation_slug": {
				Type:     schema.TypeString,
				Required: true,
			},
			"environment_slug": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"slug": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloud_provider": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cloud_account_identity": {
				Type:     schema.TypeString,
				Required: true,
			},
			"update_channel": {
				Type:     schema.TypeString,
				Optional: true,
			},
			//"node_pools": {
			//	Type:     schema.TypeList,
			//	Optional: true,
			//},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)

	var diags diag.Diagnostics

	nodePools := []acloudapi.NodePools{}

	createCluster := acloudapi.CreateCluster{
		Name:                 d.Get("name").(string),
		Version:              d.Get("version").(string),
		Region:               d.Get("region").(string),
		CloudAccountIdentity: d.Get("cloud_account_identity").(string),
		SLA:                  "none",
		NodePools:            nodePools,
	}

	org := d.Get("organisation_slug").(string)
	env := d.Get("environment_slug").(string)

	cluster, err := client.CreateCluster(ctx, org, env, createCluster)

	if err != nil {
		return diag.FromErr(err)
	}
	if cluster != nil {
		d.SetId(cluster.Identity)
		d.Set("slug", cluster.Slug)
		d.Set("cloud_provider", cluster.CloudProvider)
		return diags
	}

	return diag.FromErr(fmt.Errorf("cluster was not created"))
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)
	var diags diag.Diagnostics

	org := d.Get("organisation_slug").(string)
	env := d.Get("environment_slug").(string)
	slug := d.Get("slug").(string)

	cluster, err := client.GetCluster(ctx, org, env, slug)
	if err != nil {
		return diag.FromErr(err)
	}
	if cluster != nil {
		return diags
	}
	return diag.FromErr(fmt.Errorf("cluster was not found"))
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)
	var diags diag.Diagnostics

	org := d.Get("organisation_slug").(string)
	env := d.Get("environment_slug").(string)
	slug := d.Get("slug").(string)

	updateCluster := acloudapi.UpdateCluster{
		UpdateChannel: d.Get("update_channel").(string),
		Version:       d.Get("version").(string),
	}

	cluster, err := client.UpdateCluster(ctx, org, env, slug, updateCluster)

	if err != nil {
		return diag.FromErr(err)
	}
	if cluster != nil {
		return diags
	}

	return diags
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)
	var diags diag.Diagnostics

	org := d.Get("organisation_slug").(string)
	env := d.Get("environment_slug").(string)
	slug := d.Get("slug").(string)

	updateCluster := acloudapi.UpdateCluster{
		Status: "deleting",
	}

	err := client.DeleteCluster(ctx, org, env, slug, updateCluster)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
