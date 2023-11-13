package acloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

func dataSourceUpdateChannel() *schema.Resource {
	return &schema.Resource{
		Description: "Get a Kubernetes update channel, including current Avisi Cloud Kubernetes version",
		ReadContext: dataSourceUpdateChannelRead,
		Schema: map[string]*schema.Schema{
			"organisation": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug of the Organisation",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the update channel",
			},
			"available": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Returns if the update channel is available",
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Avisi Cloud Kubernetes Version associated with the Update Channel",
			},
		},
	}
}

func dataSourceUpdateChannelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)
	orgSlug := d.Get("organisation").(string)
	channelName := d.Get("name").(string)
	updateChannels, err := client.GetUpdateChannels(ctx, orgSlug)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get update channel: %w", err))
	}

	for _, updateChannel := range updateChannels {
		if updateChannel.Name == channelName {
			d.SetId(updateChannel.Name)
			d.Set("available", updateChannel.Available)
			d.Set("version", updateChannel.KubernetesClusterVersion)
			return nil
		}
	}
	return diag.FromErr(fmt.Errorf("update channel was not found"))
}
