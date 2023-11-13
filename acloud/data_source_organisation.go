package acloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

func dataSourceOrganisations() *schema.Resource {
	return &schema.Resource{
		Description: "Get an organisation",
		ReadContext: dataSourceOrganisationsRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"slug": {
				Type:     schema.TypeString,
				Required: true,
			},
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceOrganisationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)

	organisations, err := client.GetMemberships(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	slug := d.Get("slug")

	for _, org := range organisations {
		if org.Slug == slug {
			d.Set("id", org.ID)
			d.Set("name", org.Name)
			d.Set("slug", org.Slug)
			d.Set("email", org.Email)
			return nil
		}
	}
	return diag.FromErr(fmt.Errorf("not found"))
}
