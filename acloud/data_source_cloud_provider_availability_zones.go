package acloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudProviderAvailabilityZones() *schema.Resource {
	return &schema.Resource{
		Description: "List all availability zones for a given cloud region",
		ReadContext: dataSourceCloudProviderAvailabilityZonesRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organisation": {
				Type:        schema.TypeString,
				Description: "Slug of the Organisation. Defaults to the `organisation` of the provider.",
				Optional:    true,
			},
			"cloud_provider": {
				Type:        schema.TypeString,
				Description: "Slug of the cloud provider, for example `aws`.",
				Required:    true,
			},
			"region": {
				Type:        schema.TypeString,
				Description: "Slug of the region, for example `eu-west-1`.",
				Required:    true,
			},
			"availability_zones": {
				Type:        schema.TypeList,
				Description: "Slugs of the availability zones in the region.",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceCloudProviderAvailabilityZonesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}

	cloudProviderSlug := d.Get("cloud_provider").(string)
	regionSlug := d.Get("region").(string)

	availabilityZones, err := client.GetAvailabilityZones(ctx, org, cloudProviderSlug, regionSlug)
	if err != nil {
		return diag.FromErr(err)
	}

	setAsID(d, fmt.Sprintf("%s-%s", org, cloudProviderSlug))

	zones := []string{}
	for _, az := range availabilityZones {
		zones = append(zones, az.Slug)
	}
	d.Set("availability_zones", zones)
	return diags
}
