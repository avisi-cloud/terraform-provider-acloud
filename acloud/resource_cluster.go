package acloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

type ClusterState string

const (
	ClusterStateRunning ClusterState = "running"
	ClusterStateStopped ClusterState = "stopped"
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stopped": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"cluster_state_wait_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  600,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
		err := WaitUntilClusterHasStatus(ctx, d, m, org, *cluster, string(ClusterStateRunning))
		if err != nil {
			return diag.FromErr(err)
		}
		return diags
	}

	return resourceClusterRead(ctx, d, m)
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
	if cluster == nil {
		return diag.FromErr(fmt.Errorf("cluster was not found"))
	}

	d.SetId(cluster.Identity)
	d.Set("name", cluster.Name)
	d.Set("description", cluster.Description)
	d.Set("slug", cluster.Slug)
	d.Set("cloud_provider", cluster.CloudProvider)
	d.Set("region", cluster.Region)
	d.Set("version", cluster.Version)
	d.Set("update_channel", cluster.UpdateChannel)
	d.Set("status", cluster.Status)

	return diags
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)

	diags := resourceClusterRead(ctx, d, m)
	if diags != nil && diags.HasError() {
		return diags
	}

	org := d.Get("organisation_slug").(string)
	env := d.Get("environment_slug").(string)
	slug := d.Get("slug").(string)

	stopped := d.Get("stopped").(bool)
	status := d.Get("status").(string)

	updateCluster := acloudapi.UpdateCluster{
		UpdateChannel: d.Get("update_channel").(string),
		Version:       d.Get("version").(string),
	}

	desiredStatus := "running"
	if stopped {
		desiredStatus = "stopped"
	}
	if desiredStatus != status {
		updateCluster.Status = getTransitionStatus(desiredStatus)
	}

	cluster, err := client.UpdateCluster(ctx, org, env, slug, updateCluster)

	if err != nil {
		return diag.FromErr(err)
	}
	if cluster != nil {
		err := WaitUntilClusterHasStatus(ctx, d, m, org, *cluster, desiredStatus)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceClusterRead(ctx, d, m)
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

	d.SetId("")

	return diags
}

func getTransitionStatus(desiredStatus string) string {
	if desiredStatus == string(ClusterStateRunning) {
		return "starting"
	} else if desiredStatus == string(ClusterStateStopped) {
		return "stopping"
	}
	return desiredStatus
}

func WaitUntilClusterHasStatus(ctx context.Context, d *schema.ResourceData, m interface{}, org string, cluster acloudapi.Cluster, desiredStatus string) error {
	client := m.(acloudapi.Client)

	if cluster.Status == desiredStatus {
		return nil
	}

	clusterStateWaitSeconds := d.Get("cluster_state_wait_seconds").(int)

	return eventually(ctx, func(ctx context.Context) error {
		c, err := client.GetCluster(ctx, org, cluster.EnvironmentSlug, cluster.Slug)
		if err != nil {
			return err
		}

		if c.Status != desiredStatus {
			return fmt.Errorf("cluster has not reached status, current status: %s", c.Status)
		}
		return nil
	}, time.Duration(clusterStateWaitSeconds)*time.Second)
}

func eventually(ctx context.Context, f func(ctx context.Context) error, timeout time.Duration) error {
	withTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-withTimeout.Done():
			return withTimeout.Err()
		case <-time.After(10 * time.Second):
			err := f(withTimeout)
			if err != nil {
				// TODO: break on unrecoverable errors, such as 401's
				continue
			}
			withTimeout.Done()
			return nil
		}
	}
}
