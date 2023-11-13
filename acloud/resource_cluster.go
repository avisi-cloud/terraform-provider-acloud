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
		Description:   "Create an Avisi Cloud Kubernetes cluster within an environment",
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
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Cluster",
			},
			"organisation": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug of the Organisation of the Cluster",
			},
			"environment": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug of the Environment of the Cluster",
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
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the Cluster",
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
				Type:        schema.TypeString,
				Required:    true,
				Description: "Region of the Cloud Provider to deploy the Cluster in",
			},
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Avisi Cloud Kubernetes version of the Cluster",
			},
			"cloud_account_identity": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Identity of the Cloud Account used to deploy the Cluster",
			},
			"update_channel": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Avisi Cloud Kubernetes Update Channel that the Cluster follows",
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stopped": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Stops the Cluster if set to true. False by default",
			},
			"cluster_state_wait_seconds": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     600,
				Description: "Time-out for waiting until the cluster reaches the desired state",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)
	nodePools := []acloudapi.NodePools{}

	createCluster := acloudapi.CreateCluster{
		Name:                 d.Get("name").(string),
		Version:              d.Get("version").(string),
		Region:               d.Get("region").(string),
		CloudAccountIdentity: d.Get("cloud_account_identity").(string),
		SLA:                  "none",
		NodePools:            nodePools,
	}

	org := getStringAttributeWithLegacyName(d, "organisation", "organisation_slug")
	env := getStringAttributeWithLegacyName(d, "environment", "environment_slug")

	cluster, err := client.CreateCluster(ctx, org, env, createCluster)

	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create cluster: %w", err))
	}

	if cluster != nil {
		d.SetId(cluster.Identity)
		d.Set("slug", cluster.Slug)
		d.Set("cloud_provider", cluster.CloudProvider)
		err := WaitUntilClusterHasStatus(ctx, d, m, org, *cluster, string(ClusterStateRunning))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error while waiting for cluster: %w", err))
		}
		return nil
	}

	return resourceClusterRead(ctx, d, m)
}

func getStringAttributeWithLegacyName(d *schema.ResourceData, names ...string) string {
	defaultValue := ""
	for _, attributeName := range names {
		value := d.Get(attributeName)
		if value != nil && value != "" {
			defaultValue = value.(string)
		}
	}
	return defaultValue
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)
	org := getStringAttributeWithLegacyName(d, "organisation", "organisation_slug")
	env := getStringAttributeWithLegacyName(d, "environment", "environment_slug")

	slug := d.Get("slug").(string)

	cluster, err := client.GetCluster(ctx, org, env, slug)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to find cluster in org %s and env %s: %w", org, env, err))
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

	return nil
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)

	diags := resourceClusterRead(ctx, d, m)
	if diags != nil && diags.HasError() {
		return diags
	}

	org := getStringAttributeWithLegacyName(d, "organisation", "organisation_slug")
	env := getStringAttributeWithLegacyName(d, "environment", "environment_slug")
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
		return diag.FromErr(fmt.Errorf("failed to update cluster: %w", err))
	}
	if cluster != nil {
		err := WaitUntilClusterHasStatus(ctx, d, m, org, *cluster, desiredStatus)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error while waiting for cluster: %w", err))
		}
	}
	return resourceClusterRead(ctx, d, m)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)

	org := getStringAttributeWithLegacyName(d, "organisation", "organisation_slug")
	env := getStringAttributeWithLegacyName(d, "environment", "environment_slug")
	slug := d.Get("slug").(string)

	updateCluster := acloudapi.UpdateCluster{
		Status: "deleting",
	}

	err := client.DeleteCluster(ctx, org, env, slug, updateCluster)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete cluster: %w", err))
	}

	d.SetId("")

	return nil
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
