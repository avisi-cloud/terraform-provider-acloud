package acloud

import (
	"context"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMaintenanceSchedule() *schema.Resource {
	return &schema.Resource{
		Description:   "Create a maintenance schedule",
		CreateContext: resourceMaintenanceScheduleCreate,
		ReadContext:   resourceMaintenanceScheduleRead,
		UpdateContext: resourceMaintenanceScheduleUpdate,
		DeleteContext: resourceMaintenanceScheduleDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organisation": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Slug of the Organisation. Can only be set on creation.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the maintenance schedule",
			},
			"windows": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of maintenance windows for the schedule",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"day": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Day of the maintenance window",
						},
						"start_time": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Start time of the maintenance window",
						},
						"duration": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Duration in minutes of the maintenance window",
						},
					},
				},
			},
		},
	}
}

func resourceMaintenanceScheduleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}

	createMaintenanceSchedule, err := client.CreateMaintenanceSchedule(ctx, org, acloudapi.CreateMaintenanceSchedule{
		Name:    d.Get("name").(string),
		Windows: castMaintenanceWindows(d.Get("windows").([]interface{})),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(createMaintenanceSchedule.Identity)
	return resourceMaintenanceScheduleRead(ctx, d, m)
}

func castMaintenanceWindows(windows []interface{}) []acloudapi.MaintenanceWindow {
	var maintenanceWindows []acloudapi.MaintenanceWindow
	for _, window := range windows {
		w := window.(map[string]interface{})
		maintenanceWindows = append(maintenanceWindows, acloudapi.MaintenanceWindow{
			Day:       w["day"].(string),
			StartTime: w["start_time"].(string),
			Duration:  w["duration"].(int),
		})
	}
	return maintenanceWindows
}

func resourceMaintenanceScheduleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}

	maintenanceSchedule, err := client.GetMaintenanceSchedule(ctx, org, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if maintenanceSchedule != nil {
		d.Set("name", maintenanceSchedule.Name)
		d.Set("windows", maintenanceSchedule.MaintenanceWindows)
	}

	return nil
}

func resourceMaintenanceScheduleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}

	updateMaintenanceSchedule, err := client.UpdateMaintenanceSchedule(ctx, org, d.Id(), acloudapi.UpdateMaintenanceSchedule{
		Name:    d.Get("name").(string),
		Windows: castMaintenanceWindows(d.Get("windows").([]interface{})),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if updateMaintenanceSchedule != nil {
		d.Set("name", updateMaintenanceSchedule.Name)
		d.Set("windows", updateMaintenanceSchedule.MaintenanceWindows)
	}

	return resourceMaintenanceScheduleRead(ctx, d, m)
}

func resourceMaintenanceScheduleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := getProvider(m)
	client := provider.Client
	org, err := getOrganisation(provider, d)
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.DeleteMaintenanceSchedule(ctx, org, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
