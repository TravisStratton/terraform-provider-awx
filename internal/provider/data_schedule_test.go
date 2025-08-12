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

func TestAccScheduleDataSource(t *testing.T) {
	schedule := ScheduleAPIModel{
		Name:               "test-schedule-" + acctest.RandString(5),
		Description:        "Initial test schedule",
		Rrule:              "DTSTART;TZID=UTC:20250301T120000 RRULE:FREQ=DAILY;INTERVAL=1",
		UnifiedJobTemplate: 1,
		Enabled:            true,
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0), // built-in check from tfversion package
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID testing
			{
				Config: testAccScheduleDataSourceConfig(schedule),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(schedule.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(schedule.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("rrule"),
						knownvalue.StringExact(schedule.Rrule),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("unified_job_template"),
						knownvalue.Int32Exact(int32(schedule.UnifiedJobTemplate)),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("enabled"),
						knownvalue.Bool(schedule.Enabled),
					),
				},
			},
		},
	})
}

func testAccScheduleDataSourceConfig(resource ScheduleAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_schedule" "test" {
  name        			= "%s"
  description 			= "%s"
  rrule       			= "%s"
  unified_job_template 	= %d
  enabled     			= %t
}
data "awx_schedule" "test" {
  id = awx_schedule.test.id
}
`, resource.Name, resource.Description, resource.Rrule, resource.UnifiedJobTemplate, resource.Enabled))
}
