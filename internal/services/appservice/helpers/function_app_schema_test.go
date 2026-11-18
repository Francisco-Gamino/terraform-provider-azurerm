// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package helpers_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/appservice/helpers"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

func TestFunctionAppPowerShellCoreVersionValidation(t *testing.T) {
	schemas := map[string]*pluginsdk.Schema{
		"linux":        helpers.SiteConfigSchemaLinuxFunctionApp(),
		"linux_slot":   helpers.SiteConfigSchemaLinuxFunctionAppSlot(),
		"windows":      helpers.SiteConfigSchemaWindowsFunctionApp(),
		"windows_slot": helpers.SiteConfigSchemaWindowsFunctionAppSlot(),
	}
	versions := []struct {
		value string
		valid bool
	}{
		{value: "7.6", valid: true},
		{value: "7"},
		{value: "7.0"},
		{value: "7.2"},
		{value: "7.4"},
	}

	for name, siteConfig := range schemas {
		siteConfigResource := siteConfig.Elem.(*pluginsdk.Resource)
		stackResource := siteConfigResource.Schema["application_stack"].Elem.(*pluginsdk.Resource)
		validate := stackResource.Schema["powershell_core_version"].ValidateFunc

		for _, version := range versions {
			t.Run(name+"/"+version.value, func(t *testing.T) {
				_, errors := validate(version.value, "powershell_core_version")
				if gotValid := len(errors) == 0; gotValid != version.valid {
					t.Fatalf("validation for %q: valid = %t, errors = %v", version.value, gotValid, errors)
				}
			})
		}
	}
}

func TestFunctionAppPowerShellLinuxFxVersionRoundTrip(t *testing.T) {
	for _, version := range []string{"7.4", "7.6"} {
		t.Run(version, func(t *testing.T) {
			want := []helpers.ApplicationStackLinuxFunctionApp{{PowerShellCoreVersion: version}}
			encoded := helpers.EncodeFunctionAppLinuxFxVersion(want)
			if encoded == nil || *encoded != "POWERSHELL|"+version {
				t.Fatalf("unexpected Linux FX version: %v", encoded)
			}

			decoded, err := helpers.DecodeFunctionAppLinuxFxVersion(*encoded)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decoded, want) {
				t.Fatalf("expected %#v, got %#v", want, decoded)
			}
		})
	}
}

func TestFunctionAppPowerShell76SiteConfig(t *testing.T) {
	linuxStack := []helpers.ApplicationStackLinuxFunctionApp{{PowerShellCoreVersion: "7.6"}}
	windowsStack := []helpers.ApplicationStackWindowsFunctionApp{{PowerShellCoreVersion: "7.6"}}
	cases := []struct {
		name        string
		schema      *pluginsdk.Schema
		expand      func(sdk.ResourceMetaData) (*webapps.SiteConfig, error)
		wantLinuxFx string
	}{
		{
			name:   "linux",
			schema: helpers.SiteConfigSchemaLinuxFunctionApp(),
			expand: func(metadata sdk.ResourceMetaData) (*webapps.SiteConfig, error) {
				return helpers.ExpandSiteConfigLinuxFunctionApp([]helpers.SiteConfigLinuxFunctionApp{{ApplicationStack: linuxStack}}, nil, metadata, "~4", "storage", false)
			},
			wantLinuxFx: "POWERSHELL|7.6",
		},
		{
			name:   "linux_slot",
			schema: helpers.SiteConfigSchemaLinuxFunctionAppSlot(),
			expand: func(metadata sdk.ResourceMetaData) (*webapps.SiteConfig, error) {
				return helpers.ExpandSiteConfigLinuxFunctionAppSlot([]helpers.SiteConfigLinuxFunctionAppSlot{{ApplicationStack: linuxStack}}, nil, metadata, "~4", "storage", false)
			},
			wantLinuxFx: "PowerShell|7.6",
		},
		{
			name:   "windows",
			schema: helpers.SiteConfigSchemaWindowsFunctionApp(),
			expand: func(metadata sdk.ResourceMetaData) (*webapps.SiteConfig, error) {
				return helpers.ExpandSiteConfigWindowsFunctionApp([]helpers.SiteConfigWindowsFunctionApp{{ApplicationStack: windowsStack}}, nil, metadata, "~4", "storage", false)
			},
		},
		{
			name:   "windows_slot",
			schema: helpers.SiteConfigSchemaWindowsFunctionAppSlot(),
			expand: func(metadata sdk.ResourceMetaData) (*webapps.SiteConfig, error) {
				return helpers.ExpandSiteConfigWindowsFunctionAppSlot([]helpers.SiteConfigWindowsFunctionAppSlot{{ApplicationStack: windowsStack}}, nil, metadata, "~4", "storage", false)
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, map[string]*schema.Schema{"site_config": test.schema}, map[string]interface{}{
				"site_config": []interface{}{map[string]interface{}{}},
			})
			config, err := test.expand(sdk.ResourceMetaData{ResourceData: data})
			if err != nil {
				t.Fatal(err)
			}
			if test.wantLinuxFx != "" {
				if got := pointer.From(config.LinuxFxVersion); got != test.wantLinuxFx {
					t.Fatalf("expected Linux FX version %q, got %q", test.wantLinuxFx, got)
				}
			} else if got := pointer.From(config.PowerShellVersion); got != "7.6" {
				t.Fatalf("expected PowerShell version 7.6, got %q", got)
			}

			settings := make(map[string]string)
			for _, setting := range pointer.From(config.AppSettings) {
				settings[pointer.From(setting.Name)] = pointer.From(setting.Value)
			}
			if settings["FUNCTIONS_WORKER_RUNTIME"] != "powershell" || settings["FUNCTIONS_EXTENSION_VERSION"] != "~4" {
				t.Fatalf("unexpected runtime settings: %#v", settings)
			}
		})
	}
}

func TestMergeUserAppSettings(t *testing.T) {
	cases := []struct {
		service  []webapps.NameValuePair
		user     map[string]string
		expected []webapps.NameValuePair // Note: The function doesn't preserve the order of the list
	}{
		{
			service: []webapps.NameValuePair{
				{
					Name:  pointer.To("test"),
					Value: pointer.To("ServiceValue"),
				},
			},
			user: map[string]string{
				"test": "UserValue",
			},
			expected: []webapps.NameValuePair{
				{
					Name:  pointer.To("test"),
					Value: pointer.To("UserValue"),
				},
			},
		},
		{
			service: []webapps.NameValuePair{
				{
					Name:  pointer.To("test"),
					Value: pointer.To("ServiceValue"),
				},
				{
					Name:  pointer.To("test2"),
					Value: pointer.To("ServiceValue2"),
				},
				{
					Name:  pointer.To("test3"),
					Value: pointer.To("ServiceValue3"),
				},
				{
					Name:  pointer.To("test4"),
					Value: pointer.To("ServiceValue4"),
				},
			},
			user: map[string]string{
				"test":  "UserValue",
				"test4": "UserValue4",
			},
			expected: []webapps.NameValuePair{
				{
					Name:  pointer.To("test"),
					Value: pointer.To("UserValue"),
				},
				{
					Name:  pointer.To("test2"),
					Value: pointer.To("ServiceValue2"),
				},
				{
					Name:  pointer.To("test3"),
					Value: pointer.To("ServiceValue3"),
				},
				{
					Name:  pointer.To("test4"),
					Value: pointer.To("UserValue4"),
				},
			},
		},
	}

	for _, v := range cases {
		actualRaw := helpers.MergeUserAppSettings(&v.service, v.user)
		actual := *actualRaw
		sort.Slice(actual, func(i, j int) bool {
			return *actual[i].Name < *actual[j].Name
		})
		if !reflect.DeepEqual(actual, v.expected) {
			t.Fatalf("expected %+v, got %+v", v.expected, actual)
		}
	}
}
