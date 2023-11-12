package acloud

import (
	"context"
	"crypto/sha1"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ACLOUD_PERSONAL_ACCESS_TOKEN", nil),
			},
			"acloud_api": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ACLOUD_API_ENDPOINT", "https://api.avisi.cloud"),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"acloud_environment": resourceEnvironment(),
			"acloud_cluster":     resourceCluster(),
			"acloud_nodepool":    resourceNodepool(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"acloud_cloud_account":                     dataSourceCloudAccount(),
			"acloud_cloud_accounts":                    dataSourceCloudAccounts(),
			"acloud_cloud_provider_availability_zones": dataSourceCloudProviderAvailabilityZones(),
			"acloud_cloud_provider_node_types":         dataSourceCloudProviderNodeTypes(),
			"acloud_cloud_provider_regions":            dataSourceCloudProviderRegions(),
			"acloud_cloud_providers":                   dataSourceCloudProviders(),
			"acloud_cluster":                           dataSourceCluster(),
			"acloud_environment":                       dataSourceEnvironment(),
			"acloud_nodepool_join_config":              dataSourceNodeJoinConfig(),
			"acloud_organisation":                      dataSourceOrganisations(),
			"acloud_update_channel":                    dataSourceUpdateChannel(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	token := d.Get("token").(string)
	acloudApiEndpoint := d.Get("acloud_api").(string)

	authenticator := acloudapi.NewPersonalAccessTokenAuthenticator(token)
	clientOpts := acloudapi.ClientOpts{
		APIUrl: acloudApiEndpoint,
	}

	c := acloudapi.NewClient(authenticator, clientOpts)
	if token != "" {
		c.Resty().OnBeforeRequest(authenticator.Authenticate)
	}
	return c, nil
}

func setAsID(d *schema.ResourceData, customID string) {
	computedId := sha1.Sum([]byte(customID))
	d.SetId(fmt.Sprintf("%x", computedId))
}
