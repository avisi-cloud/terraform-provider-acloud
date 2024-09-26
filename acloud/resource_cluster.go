package acloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

type ClusterState string

const (
	ClusterStateRunning ClusterState = "running"
	ClusterStateStopped ClusterState = "stopped"
)

func resourceCluster() *schema.Resource {
	return &schema.Resource{
		Description:   "Create an Avisi Cloud Kubernetes cluster within an environment",
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the Cluster",
			},
			"organisation": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Slug of the Organisation of the Cluster. Can only be set on cluster creation.",
			},
			"environment": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Slug of the Environment of the Cluster. Can only be set on cluster creation.",
			},
			"organisation_slug": {
				Type:       schema.TypeString,
				Deprecated: "replaced by organisation",
				Optional:   true,
				Default:    nil,
			},
			"environment_slug": {
				Type:       schema.TypeString,
				Deprecated: "replaced by environment",
				Optional:   true,
				Default:    nil,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the Cluster",
			},
			"slug": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cni": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "CNI plugin for Kubernetes",
			},
			"cloud_provider": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Region of the Cloud Provider to deploy the Cluster in. Can only be set on cluster creation.",
			},
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Avisi Cloud Kubernetes version of the Cluster",
			},
			"cloud_account_identity": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Identity of the Cloud Account used to deploy the Cluster. Can only be set on cluster creation.",
			},
			"update_channel": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Avisi Cloud Kubernetes Update Channel that the Cluster follows",
			},
			"pod_security_standards_profile": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "PRIVILEGED",
				Description: "Pod Security Standards used by default within the cluster",
			},
			"enable_multi_availability_zones": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
				Description: "Enable multi availability zones for the cluster",
			},
			"enable_high_available_control_plane": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable Highly-Availability mode for the cluster's Kubernetes Control Plane",
			},
			"enable_private_cluster": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Enable Private Cluster mode. Can only be set on cluster creation.",
			},
			"enable_network_encryption": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enable Network Encryption at the node level (if supported by the CNI).",
			},
			"enable_auto_upgrade": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable auto-upgrade for the cluster",
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stopped": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Stops the Cluster if set to true. False by default",
			},
			"cluster_state_wait_seconds": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     600,
				Description: "Time-out for waiting until the cluster reaches the desired state",
			},
			"maintenance_schedule_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the maintenance schedule to apply to the cluster",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}

	nodePools := []acloudapi.NodePools{}

	createCluster := acloudapi.CreateCluster{
		Name:                         d.Get("name").(string),
		Version:                      d.Get("version").(string),
		Region:                       d.Get("region").(string),
		CNI:                          d.Get("cni").(string),
		PodSecurityStandardsProfile:  d.Get("pod_security_standards_profile").(string),
		EnableMultiAvailabilityZones: d.Get("enable_multi_availability_zones").(bool),
		EnableHighAvailability:       d.Get("enable_high_available_control_plane").(bool),
		EnableNATGateway:             d.Get("enable_private_cluster").(bool),
		EnableNetworkEncryption:      d.Get("enable_network_encryption").(bool),
		EnableAutoUpgrade:            d.Get("enable_auto_upgrade").(bool),
		CloudAccountIdentity:         d.Get("cloud_account_identity").(string),
		NodePools:                    nodePools,
		MaintenanceScheduleIdentity:  d.Get("maintenance_schedule_id").(string),
		UpdateChannel:                d.Get("update_channel").(string),
	}

	env := getStringAttributeWithLegacyName(d, "environment", "environment_slug")

	cluster, err := client.CreateCluster(ctx, org, env, createCluster)

	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create cluster: %w", err))
	}

	if cluster != nil {
		d.SetId(cluster.Identity)
		d.Set("slug", cluster.Slug)
		d.Set("cloud_provider", cluster.CloudProvider)
		err := WaitUntilClusterHasStatus(ctx, d, m, org, *cluster, string(ClusterStateRunning))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error while waiting for cluster: %w", err))
		}
		return nil
	}

	return resourceClusterRead(ctx, d, m)
}

func getStringAttributeWithLegacyName(d *schema.ResourceData, names ...string) string {
	defaultValue := ""
	for _, attributeName := range names {
		value := d.Get(attributeName)
		if value != nil && value != "" {
			defaultValue = value.(string)
		}
	}
	return defaultValue
}

func getOrganisation(provider ConfiguredProvider, d *schema.ResourceData) (string, error) {
	organisation := getStringAttributeWithLegacyName(d, "organisation", "organisation_slug")
	if organisation != "" {
		return organisation, nil
	}
	if provider.Organisation != "" {
		return provider.Organisation, nil
	}
	return "", errors.New("organisation is not set")
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}

	env := getStringAttributeWithLegacyName(d, "environment", "environment_slug")

	slug := d.Get("slug").(string)

	cluster, err := client.GetCluster(ctx, org, env, slug)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to find cluster in org %s and env %s: %w", org, env, err))
	}
	if cluster == nil {
		return diag.FromErr(fmt.Errorf("cluster was not found"))
	}

	d.SetId(cluster.Identity)
	d.Set("name", cluster.Name)
	d.Set("description", cluster.Description)
	d.Set("slug", cluster.Slug)
	d.Set("cloud_provider", cluster.CloudProvider)
	d.Set("cni", cluster.CNI)
	d.Set("region", cluster.Region)
	d.Set("version", cluster.Version)
	d.Set("update_channel", cluster.UpdateChannel)
	d.Set("pod_security_standards_profile", cluster.PodSecurityStandardsProfile)
	d.Set("enable_multi_availability_zones", cluster.EnableMultiAvailAbilityZones)
	d.Set("enable_high_available_control_plane", cluster.HighlyAvailable)
	d.Set("enable_private_cluster", cluster.EnableNATGateway)
	d.Set("enable_network_encryption", cluster.EnableNetworkEncryption)
	d.Set("enable_auto_upgrade", cluster.AutoUpgrade)
	d.Set("status", cluster.Status)
	d.Set("maintenance_schedule_id", cluster.MaintenanceSchedule.Identity)

	return nil
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceClusterRead(ctx, d, m)
	if diags != nil && diags.HasError() {
		return diags
	}

	env := getStringAttributeWithLegacyName(d, "environment", "environment_slug")
	slug := d.Get("slug").(string)

	stopped := d.Get("stopped").(bool)
	status := d.Get("status").(string)

	enableNetworkEncryption := d.Get("enable_network_encryption").(bool)
	if d.HasChange("enable_network_encryption") {
		_, newVal := d.GetChange("enable_network_encryption")
		enableNetworkEncryption = newVal.(bool)
	}

	enableHAControlPlane := d.Get("enable_high_available_control_plane").(bool)
	if d.HasChange("enable_high_available_control_plane") {
		_, newVal := d.GetChange("enable_high_available_control_plane")
		enableHAControlPlane = newVal.(bool)
	}

	enableAutoUpgrade := d.Get("enable_auto_upgrade").(bool)
	if d.HasChange("enable_auto_upgrade") {
		_, newVal := d.GetChange("enable_auto_upgrade")
		enableAutoUpgrade = newVal.(bool)
	}

	pss := d.Get("pod_security_standards_profile").(string)
	if d.HasChange("pod_security_standards_profile") {
		_, newVal := d.GetChange("pod_security_standards_profile")
		pss = newVal.(string)
	}

	maintenanceScheduleIdentity := d.Get("maintenance_schedule_id").(string)
	if d.HasChange("maintenance_schedule_id") {
		_, newVal := d.GetChange("maintenance_schedule_id")
		maintenanceScheduleIdentity = newVal.(string)
	}

	updateCluster := acloudapi.UpdateCluster{
		UpdateChannel:               d.Get("update_channel").(string),
		Version:                     d.Get("version").(string),
		PodSecurityStandardsProfile: &pss,
		EnableNetworkEncryption:     &enableNetworkEncryption,
		EnableHighAvailability:      &enableHAControlPlane,
		EnableAutoUpgrade:           &enableAutoUpgrade,
		MaintenanceScheduleIdentity: &maintenanceScheduleIdentity,
	}

	desiredStatus := "running"
	if stopped {
		desiredStatus = "stopped"
	}
	if desiredStatus != status {
		updateCluster.Status = getTransitionStatus(desiredStatus)
	}

	cluster, err := client.UpdateCluster(ctx, org, env, slug, updateCluster)

	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update cluster: %w", err))
	}
	if cluster != nil {
		err := WaitUntilClusterHasStatus(ctx, d, m, org, *cluster, desiredStatus)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error while waiting for cluster: %w", err))
		}
	}
	return resourceClusterRead(ctx, d, m)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}

	env := getStringAttributeWithLegacyName(d, "environment", "environment_slug")
	slug := d.Get("slug").(string)

	updateCluster := acloudapi.UpdateCluster{
		Status: "deleting",
	}

	error := client.DeleteCluster(ctx, org, env, slug, updateCluster)
	if error != nil {
		return diag.FromErr(fmt.Errorf("failed to delete cluster: %w", err))
	}

	d.SetId("")

	return nil
}

func getProvider(m interface{}) ConfiguredProvider {

	p, ok := m.(ConfiguredProvider)
	if !ok {
		panic("invalid configured provider")
	}

	return p
}

func getTransitionStatus(desiredStatus string) string {
	if desiredStatus == string(ClusterStateRunning) {
		return "starting"
	} else if desiredStatus == string(ClusterStateStopped) {
		return "stopping"
	}
	return desiredStatus
}

func WaitUntilClusterHasStatus(ctx context.Context, d *schema.ResourceData, m interface{}, org string, cluster acloudapi.Cluster, desiredStatus string) error {
	provider := getProvider(m)
	client := provider.Client

	if cluster.Status == desiredStatus {
		return nil
	}

	clusterStateWaitSeconds := d.Get("cluster_state_wait_seconds").(int)

	return eventually(ctx, func(ctx context.Context) error {
		c, err := client.GetCluster(ctx, org, cluster.EnvironmentSlug, cluster.Slug)
		if err != nil {
			return err
		}

		if c.Status != desiredStatus {
			return fmt.Errorf("cluster has not reached status, current status: %s", c.Status)
		}
		return nil
	}, time.Duration(clusterStateWaitSeconds)*time.Second)
}

func eventually(ctx context.Context, f func(ctx context.Context) error, timeout time.Duration) error {
	withTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-withTimeout.Done():
			return withTimeout.Err()
		case <-time.After(10 * time.Second):
			err := f(withTimeout)
			if err != nil {
				// TODO: break on unrecoverable errors, such as 401's
				continue
			}
			withTimeout.Done()
			return nil
		}
	}
}
