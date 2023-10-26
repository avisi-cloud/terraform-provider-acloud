package acloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

func dataSourceNodeJoinConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNodeJoinConfigRead,
		Description: "Provides access to node join configuration for a node pool. Can be used in combination with other terraform providers to provision new Kubernetes Nodes for Bring Your Own Node clusters in Avisi Cloud Kubernetes.",
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"organisation_slug": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug of the Organisation",
			},
			"environment_slug": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug of the environment of the cluster",
			},
			"cluster_slug": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug of the cluster",
			},
			"node_pool_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the node pool",
			},
			"user_data": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Cloud Init user-data (base64)",
			},
			"kubelet_config": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"join_command": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"install_script": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Install bash script for joining a node (base64).",
			},
			"upgrade_script": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Install bash script for upgrading a node (base64)",
			},
		},
	}
}

func dataSourceNodeJoinConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)

	organisationSlug := d.Get("organisation_slug").(string)
	environmentSlug := d.Get("environment_slug").(string)
	clusterSlug := d.Get("cluster_slug").(string)

	cluster, err := client.GetCluster(ctx, organisationSlug, environmentSlug, clusterSlug)
	if err != nil {
		return diag.FromErr(err)
	}

	if cluster == nil {
		return diag.FromErr(fmt.Errorf("cluster was not found"))
	}

	nodePoolID := d.Get("node_pool_id").(string)

	nodeJoinConfig, err := client.GetNodePoolJoinConfig(ctx, *cluster, acloudapi.NodePool{
		Identity: nodePoolID,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if nodeJoinConfig == nil {
		return diag.FromErr(fmt.Errorf("node join configuration was not found"))
	}

	d.SetId(nodePoolID)
	d.Set("user_data", nodeJoinConfig.CloudInitUserDataBase64)
	d.Set("kubelet_config", nodeJoinConfig.KubeletConfigBase64)
	d.Set("join_command", nodeJoinConfig.JoinCommand)
	d.Set("install_script", nodeJoinConfig.InstallScriptBase64)
	d.Set("upgrade_script", nodeJoinConfig.UpgradeScriptBase64)
	return nil
}
