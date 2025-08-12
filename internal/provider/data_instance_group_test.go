package provider

import (
	"fmt"
	"testing"

	"github.com/TravisStratton/terraform-provider-awx/internal/configprefix"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccInstanceGroupDataSource(t *testing.T) {
	resource1 := InstanceGroupAPIModel{
		Name:                     "test-instance-group-" + acctest.RandString(5),
		PolicyInstancePercentage: 100,
	}

	resource2 := InstanceGroupAPIModel{
		Name:             "test-container-group-" + acctest.RandString(5),
		IsContainerGroup: true,
		PodSpecOverride:  "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"namespace\":\"awx\"},\"spec\":{\"automountServiceAccountToken\":false,\"containers\":[{\"args\":[\"ansible-runner\",\"worker\",\"--private-data-dir=/runner\"],\"image\":\"quay.io/ansible/awx-ee:latest\",\"name\":\"worker\",\"resources\":{\"requests\":{\"cpu\":\"250m\",\"memory\":\"100Mi\"}}}],\"serviceAccountName\":\"default\"}}",
	}
	resource2podspecoverride, ok := resource2.PodSpecOverride.(string)
	if !ok {
		return
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0), // built-in check from tfversion package
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID testing instance group
			{
				Config: testAccInstanceGroupDataSource1Config(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_instance_group.test-instance", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_instance_group.test-instance", configprefix.Prefix),
						tfjsonpath.New("is_container_group"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_instance_group.test-instance", configprefix.Prefix),
						tfjsonpath.New("max_concurrent_jobs"),
						knownvalue.Int32Exact(0),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_instance_group.test-instance", configprefix.Prefix),
						tfjsonpath.New("max_forks"),
						knownvalue.Int32Exact(0),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_instance_group.test-instance", configprefix.Prefix),
						tfjsonpath.New("policy_instance_minimum"),
						knownvalue.Int32Exact(0),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_instance_group.test-instance", configprefix.Prefix),
						tfjsonpath.New("policy_instance_percentage"),
						knownvalue.Int32Exact(int32(resource1.PolicyInstancePercentage)),
					),
				},
			},
			// Ready by Name testing container group
			{
				Config: testAccInstanceGroupDataSource2Config(resource2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_instance_group.test-container", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_instance_group.test-container", configprefix.Prefix),
						tfjsonpath.New("is_container_group"),
						knownvalue.Bool(resource2.IsContainerGroup),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_instance_group.test-container", configprefix.Prefix),
						tfjsonpath.New("pod_spec_override"),
						knownvalue.StringExact(resource2podspecoverride),
					),
				},
			},
		},
	})
}

func testAccInstanceGroupDataSource1Config(resource InstanceGroupAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_instance_group" "test-instance" {
  name                       = "%s"
  policy_instance_percentage = %d
}
data "awx_instance_group" "test-instance" {
  id = awx_instance_group.test-instance.id
}
`, resource.Name, resource.PolicyInstancePercentage))
}

func testAccInstanceGroupDataSource2Config(resource InstanceGroupAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_instance_group" "test-container" {
  name       = "%s"
  is_container_group = %v
  pod_spec_override = jsonencode(%s)
}
data "awx_instance_group" "test-container" {
  id = awx_instance_group.test-container.id
}
`, resource.Name, resource.IsContainerGroup, resource.PodSpecOverride))
}
