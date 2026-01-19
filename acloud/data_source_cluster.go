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
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Cluster UUID Identity as the ID of this Terraform resource",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The internal Terraform identifier.",
			},
			"organisation": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Slug of the organisation of the cluster",
			},
			"environment": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug of the environment that the cluster is part of",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the Cluster",
			},
			"slug": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug of the cluster",
			},
			"cni": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "CNI plugin for Kubernetes",
			},
			"cloud_provider": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Slug of the Cloud Provider",
			},
			"region": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Region of the Cloud Provider to deploy the cluster in",
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Avisi AME version of the cluster",
			},
			"cloud_account_identity": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Identity of the Cloud Account used to deploy the Cluster",
			},
			"update_channel": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Avisi AME update channel that the cluster follows",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Avisi AME Cluster status",
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
			"enable_multi_availability_zones": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Used to configure if the cluster should support multi availability zones for its node pools",
			},
			"maintenance_schedule_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "UUID Identity of the maintenance schedule for the cluster",
			},
			"addons": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Add-ons configured for the cluster",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the add-on",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the add-on is enabled",
						},
						"custom_values": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "Custom values for the add-on. Values are stringified for the API and any keys are allowed.",
						},
					},
				},
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
	d.Set("enable_multi_availability_zones", cluster.EnableMultiAvailAbilityZones)
	d.Set("status", cluster.Status)
	if cluster.MaintenanceSchedule != nil {
		d.Set("maintenance_schedule_id", cluster.MaintenanceSchedule.Identity)
	} else {
		d.Set("maintenance_schedule_id", "")
	}
	flattenedAddons := flattenClusterAddons(cluster.Addons)
	d.Set("addons", flattenedAddons)

	return nil
}
