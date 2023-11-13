package acloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

func dataSourceEnvironment() *schema.Resource {
	return &schema.Resource{
		Description: "Get an environment",
		ReadContext: dataSourceEnvironmentRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"organisation": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug of the organisation",
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the environment",
				Computed:    true,
			},
			"slug": {
				Type:        schema.TypeString,
				Description: "Slug of the environment",
				Required:    true,
			},
		},
	}
}

func dataSourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)

	orgSlug := d.Get("organisation").(string)
	slug := d.Get("slug").(string)
	environment, err := client.GetEnvironment(ctx, orgSlug, slug)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get environment: %w", err))
	}
	d.Set("id", environment.ID)
	d.Set("name", environment.Name)
	d.Set("organisation", orgSlug)
	d.Set("slug", slug)
	return nil
}
