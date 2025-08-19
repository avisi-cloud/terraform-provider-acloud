package acloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudProfile() *schema.Resource {
	return &schema.Resource{
		Description: "Get a cloud profile",
		ReadContext: dataCloudProfileRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity": {
				Type:        schema.TypeString,
				Description: "Identity of the Cloud Profile",
				Optional:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the Cloud Account",
				Optional:    true,
			},
			"public": {
				Type:        schema.TypeBool,
				Description: "Returns if the Cloud Profile is publicly available",
				Computed:    true,
			},
			"cloud_provider": {
				Type:        schema.TypeString,
				Description: "Slug of the Cloud Provider of the Cloud Account",
				Optional:    true,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Returns if the Cloud Account is enabled",
				Computed:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Type of the Cloud Profile",
				Computed:    true,
			},
			"regions": {
				Type:        schema.TypeList,
				Description: "Regions available for the Cloud Profile",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataCloudProfileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}
	displayName := d.Get("name").(string)
	cloudProvider := d.Get("cloud_provider").(string)

	cloudProfiles, err := client.GetCloudProfiles(ctx, org)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get cloud profiles for organisation %q: %w", org, err))
	}

	var cloudProfile acloudapi.CloudProfile

	for _, profile := range cloudProfiles {
		if displayName != "" && profile.CloudProvider == cloudProvider && strings.EqualFold(profile.DisplayName, displayName) {
			cloudProfile = profile
			break
		}
		if profile.Identity == d.Get("identity").(string) {
			cloudProfile = profile
			break
		}
	}
	if cloudProfile.Identity == "" {
		return diag.FromErr(fmt.Errorf("cloud profile %q not found", displayName))
	}

	d.SetId(cloudProfile.Identity)
	d.Set("identity", cloudProfile.Identity)
	d.Set("name", cloudProfile.DisplayName)
	d.Set("cloud_provider", cloudProfile.CloudProvider)
	d.Set("enabled", cloudProfile.Enabled)
	d.Set("public", cloudProfile.Public)
	d.Set("type", cloudProfile.Type)
	d.Set("regions", cloudProfile.Regions)
	return nil
}
