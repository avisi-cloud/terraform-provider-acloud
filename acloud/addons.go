package acloud

import (
	"fmt"
	"maps"

	"github.com/avisi-cloud/go-client/pkg/acloudapi"
)

func defaultClusterAddons() map[string]acloudapi.APIAddon {
	return map[string]acloudapi.APIAddon{
		"fluxOperator": {
			Enabled:      false,
			CustomValues: map[string]string{},
		},
		"ingressController": {
			Enabled: false,
			CustomValues: map[string]string{
				"type": "ingress-nginx",
			},
		},
		"defaultNetworkPolicies": {
			Enabled:      true,
			CustomValues: map[string]string{},
		},
		"sealedSecrets": {
			Enabled:      false,
			CustomValues: map[string]string{},
		},
		"amePool": {
			Enabled:      false,
			CustomValues: map[string]string{},
		},
		"certManager": {
			Enabled:      false,
			CustomValues: map[string]string{},
		},
		"cloudNativePG": {
			Enabled:      false,
			CustomValues: map[string]string{},
		},
		"logging": {
			Enabled:      false,
			CustomValues: map[string]string{},
		},
		"nfs": {
			Enabled:      false,
			CustomValues: map[string]string{},
		},
		"monitoring": {
			Enabled:      false,
			CustomValues: map[string]string{},
		},
		"gpu": {
			Enabled:      false,
			CustomValues: map[string]string{},
		},
		"kured": {
			Enabled: true,
			CustomValues: map[string]string{
				"endTime":     "06:00",
				"timeZone":    "UTC",
				"startTime":   "0:00",
				"rebootDays":  "sun,mon,tue,wed,thu,fri,sat",
				"forceReboot": "false",
			},
		},
	}
}

func mergeClusterAddons(base, overrides map[string]acloudapi.APIAddon) map[string]acloudapi.APIAddon {
	if len(base) == 0 && len(overrides) == 0 {
		return nil
	}

	merged := make(map[string]acloudapi.APIAddon, len(base)+len(overrides))
	maps.Copy(merged, base)
	maps.Copy(merged, overrides)
	return merged
}

func expandClusterAddons(raw []any) map[string]acloudapi.APIAddon {
	if len(raw) == 0 {
		return nil
	}

	addons := make(map[string]acloudapi.APIAddon, len(raw))

	for _, value := range raw {
		addonValues, ok := value.(map[string]any)
		if !ok {
			continue
		}

		name, _ := addonValues["name"].(string)
		if name == "" {
			continue
		}

		enabled, _ := addonValues["enabled"].(bool)

		var customValues map[string]string
		if rawCustomValues, ok := addonValues["custom_values"]; ok && rawCustomValues != nil {
			if customValuesMap, ok := rawCustomValues.(map[string]any); ok {
				customValues = castInterfaceMapToString(customValuesMap)
			}
		}

		addons[name] = acloudapi.APIAddon{
			Enabled:      enabled,
			CustomValues: customValues,
		}
	}

	if len(addons) == 0 {
		return nil
	}

	return addons
}

func flattenClusterAddons(addons map[string]acloudapi.APIAddon) []any {
	if len(addons) == 0 {
		return nil
	}

	result := make([]interface{}, 0, len(addons))

	for name, addon := range addons {
		flattened := map[string]any{
			"name":    name,
			"enabled": addon.Enabled,
		}

		if len(addon.CustomValues) > 0 {
			flattened["custom_values"] = addon.CustomValues
		}

		result = append(result, flattened)
	}

	return result
}

func castInterfaceMapToString(original map[string]interface{}) map[string]string {
	if len(original) == 0 {
		return nil
	}

	result := make(map[string]string, len(original))

	for key, value := range original {
		if value == nil {
			continue
		}
		result[key] = fmt.Sprint(value)
	}

	if len(result) == 0 {
		return nil
	}

	return result
}
