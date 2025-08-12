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

func TestAccScheduleResource(t *testing.T) {
	schedule1 := ScheduleAPIModel{
		Name:               "test-schedule-" + acctest.RandString(5),
		Description:        "Initial test schedule",
		Rrule:              "DTSTART;TZID=UTC:20250301T120000 RRULE:FREQ=DAILY;INTERVAL=1",
		UnifiedJobTemplate: 1,
		Enabled:            true,
	}

	schedule2 := ScheduleAPIModel{
		Name:               "test-schedule-" + acctest.RandString(5),
		Description:        "Updated test schedule",
		Rrule:              "DTSTART;TZID=UTC:20250301T140000 RRULE:FREQ=WEEKLY;INTERVAL=1",
		UnifiedJobTemplate: 1,
		Enabled:            false,
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleResourceConfig(schedule1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(schedule1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(schedule1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("rrule"),
						knownvalue.StringExact(schedule1.Rrule),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("unified_job_template"),
						knownvalue.Int32Exact(int32(schedule1.UnifiedJobTemplate)),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("enabled"),
						knownvalue.Bool(schedule1.Enabled),
					),
				},
			},
			{
				ResourceName:      "awx_schedule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleResourceConfig(schedule2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(schedule2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(schedule2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("rrule"),
						knownvalue.StringExact(schedule2.Rrule),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("unified_job_template"),
						knownvalue.Int32Exact(int32(schedule2.UnifiedJobTemplate)),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("enabled"),
						knownvalue.Bool(schedule2.Enabled),
					),
				},
			},
		},
	})
}

func testAccScheduleResourceConfig(resource ScheduleAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_schedule" "test" {
  name        			= "%s"
  description 			= "%s"
  rrule       			= "%s"
  unified_job_template 	= %d
  enabled     			= %t
}
  `, resource.Name, resource.Description, resource.Rrule, resource.UnifiedJobTemplate, resource.Enabled))
}
