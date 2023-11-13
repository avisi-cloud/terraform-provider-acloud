package acloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

func dataSourceCloudAccounts() *schema.Resource {
	return &schema.Resource{
		Description: "List all cloud accounts within an organisation",
		ReadContext: dataSourceCloudAccountsRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organisation": {
				Type:        schema.TypeString,
				Description: "Organisation Slug",
				Required:    true,
			},
			"cloud_provider": {
				Type:        schema.TypeString,
				Description: "Cloud Provider Slug",
				Optional:    true,
			},
			"cloud_account_name": {
				Type:        schema.TypeString,
				Description: "Cloud Account Name",
				Optional:    true,
			},
			"cloud_accounts": {
				Type:        schema.TypeList,
				Description: "List of Cloud Accounts",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"slug": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"available": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceCloudAccountsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(acloudapi.Client)

	organisationSlug := d.Get("organisation").(string)
	providerFilter := d.Get("cloud_provider").(string)
	accountNameFilter := d.Get("cloud_account_name").(string)

	cloudAccounts, err := client.GetCloudAccounts(ctx, organisationSlug)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(organisationSlug)

	accounts := make([]map[string]interface{}, len(cloudAccounts))
	for i, account := range cloudAccounts {
		if providerFilter != "" && account.CloudProfile.CloudProvider != providerFilter {
			continue
		}
		if accountNameFilter != "" && account.DisplayName != accountNameFilter {
			continue
		}
		accounts[i] = getCloudAccountAttributes(account)
	}
	d.Set("cloud_accounts", accounts)
	return nil
}

func getCloudAccountAttributes(cloudAccount acloudapi.CloudAccount) map[string]interface{} {
	return map[string]interface{}{
		"name":     cloudAccount.DisplayName,
		"identity": cloudAccount.Identity,
		"enabled":  cloudAccount.Enabled,
	}
}
