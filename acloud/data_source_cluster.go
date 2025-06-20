package acloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataClusterRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the Cluster",
			},
			"organisation": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Slug of the Organisation of the Cluster",
			},
			"environment": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug of the Environment of the Cluster",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the Cluster",
			},
			"slug": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug of the Cluster",
			},
			"cni": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "CNI plugin for Kubernetes",
			},
			"cloud_provider": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Region of the Cloud Provider to deploy the Cluster in",
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Avisi Cloud Kubernetes version of the Cluster",
			},
			"cloud_account_identity": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Identity of the Cloud Account used to deploy the Cluster",
			},
			"update_channel": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Avisi Cloud Kubernetes Update Channel that the Cluster follows",
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pod_security_standards_profile": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Pod Security Standards used by default within the cluster",
			},
			"delete_protection": {
				Description: "Is delete protection enabled on the cluster",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"maintenance_schedule_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the maintenance schedule for the cluster",
			},
		},
	}
}

func dataClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}

	env := d.Get("environment").(string)
	slug := d.Get("slug").(string)

	cluster, err := client.GetCluster(ctx, org, env, slug)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get cluster: %w", err))
	}
	if cluster == nil {
		return diag.FromErr(fmt.Errorf("cluster was not found"))
	}

	d.SetId(cluster.Identity)
	d.Set("name", cluster.Name)
	d.Set("description", cluster.Description)
	d.Set("slug", cluster.Slug)
	d.Set("cloud_account_identity", cluster.CloudAccount.Identity)
	d.Set("delete_protection", cluster.DeleteProtection)
	d.Set("cloud_provider", cluster.CloudProvider)
	d.Set("region", cluster.Region)
	d.Set("version", cluster.Version)
	d.Set("update_channel", cluster.UpdateChannel)
	d.Set("status", cluster.Status)
	d.Set("maintenance_schedule_id", cluster.MaintenanceSchedule.Identity)

	return nil
}
