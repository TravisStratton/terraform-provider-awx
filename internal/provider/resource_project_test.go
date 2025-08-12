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

func TestAccProjectResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	project1 := ProjectAPIModel{
		Name:           "test-project-" + acctest.RandString(5),
		Description:    "Initial test git project",
		ScmType:        "git",
		ScmUrl:         "https://github.com/example/repo.git",
		Organization:   1,
		ScmUpdOnLaunch: false,
		Timeout:        1,
	}

	project2 := ProjectAPIModel{
		Name:           "test-project-" + acctest.RandString(5),
		Description:    "Updated test git project",
		ScmType:        "git",
		ScmUrl:         "https://github.com/example/updated-repo.git",
		Organization:   1,
		ScmUpdOnLaunch: true,
		Timeout:        1,
	}

	project3 := ProjectAPIModel{
		Name:           "test-project-" + acctest.RandString(5),
		Description:    "svn project",
		ScmType:        "svn",
		ScmUrl:         "svn://bad_ip/test_repo",
		Organization:   1,
		ScmUpdOnLaunch: false,
		Timeout:        1,
	}

	project4 := ProjectAPIModel{
		Name:           "test-project-" + acctest.RandString(5),
		Description:    "archive project",
		ScmType:        "archive",
		ScmUrl:         "https://github.com/user/repo",
		Organization:   1,
		ScmUpdOnLaunch: true,
		Timeout:        1,
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectResourceConfig(project1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(project1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(project1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test", configprefix.Prefix),
						tfjsonpath.New("scm_type"),
						knownvalue.StringExact(project1.ScmType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test", configprefix.Prefix),
						tfjsonpath.New("scm_url"),
						knownvalue.StringExact(project1.ScmUrl),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test", configprefix.Prefix),
						tfjsonpath.New("scm_update_on_launch"),
						knownvalue.Bool(project1.ScmUpdOnLaunch),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test", configprefix.Prefix),
						tfjsonpath.New("timeout"),
						knownvalue.Int32Exact(int32(project1.Timeout)),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_project.test", configprefix.Prefix),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      "awx_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectResourceConfig(project2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(project2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(project2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test", configprefix.Prefix),
						tfjsonpath.New("scm_type"),
						knownvalue.StringExact(project2.ScmType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test", configprefix.Prefix),
						tfjsonpath.New("scm_url"),
						knownvalue.StringExact(project2.ScmUrl),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test", configprefix.Prefix),
						tfjsonpath.New("scm_update_on_launch"),
						knownvalue.Bool(project2.ScmUpdOnLaunch),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test", configprefix.Prefix),
						tfjsonpath.New("timeout"),
						knownvalue.Int32Exact(int32(project2.Timeout)),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_project.test", configprefix.Prefix),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
			{
				Config: testAccProjectResource3Config(project3),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test-svn", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(project3.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test-svn", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(project3.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test-svn", configprefix.Prefix),
						tfjsonpath.New("scm_type"),
						knownvalue.StringExact(project3.ScmType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test-svn", configprefix.Prefix),
						tfjsonpath.New("scm_url"),
						knownvalue.StringExact(project3.ScmUrl),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test-svn", configprefix.Prefix),
						tfjsonpath.New("scm_update_on_launch"),
						knownvalue.Bool(project3.ScmUpdOnLaunch),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test-svn", configprefix.Prefix),
						tfjsonpath.New("timeout"),
						knownvalue.Int32Exact(int32(project3.Timeout)),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test-svn", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_project.test-svn", configprefix.Prefix),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
			{
				Config: testAccProjectResource4Config(project4),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test-archive", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(project4.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test-archive", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(project4.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test-archive", configprefix.Prefix),
						tfjsonpath.New("scm_type"),
						knownvalue.StringExact(project4.ScmType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test-archive", configprefix.Prefix),
						tfjsonpath.New("scm_url"),
						knownvalue.StringExact(project4.ScmUrl),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test-archive", configprefix.Prefix),
						tfjsonpath.New("scm_update_on_launch"),
						knownvalue.Bool(project4.ScmUpdOnLaunch),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.test-archive", configprefix.Prefix),
						tfjsonpath.New("timeout"),
						knownvalue.Int32Exact(int32(project4.Timeout)),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test-archive", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_project.test-archive", configprefix.Prefix),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccProjectResourceConfig(resource ProjectAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "test" {
  name        			= "%s"
}
resource "awx_project" "test" {
  name         			= "%s"
  description  			= "%s"
  scm_type     			= "%s"
  scm_url      			= "%s"
  organization 			= awx_organization.test.id
  scm_update_on_launch 	= %v
  timeout				= %d
}
  `, acctest.RandString(5), resource.Name, resource.Description, resource.ScmType, resource.ScmUrl, resource.ScmUpdOnLaunch, resource.Timeout))
}

func testAccProjectResource3Config(resource ProjectAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "test-svn" {
  name        			= "%s"
}
resource "awx_project" "test-svn" {
  name         			= "%s"
  description  			= "%s"
  scm_type     			= "%s"
  scm_url      			= "%s"
  organization 			= awx_organization.test-svn.id
  scm_update_on_launch 	= %v
  timeout				= %d
}
  `, acctest.RandString(5), resource.Name, resource.Description, resource.ScmType, resource.ScmUrl, resource.ScmUpdOnLaunch, resource.Timeout))
}

func testAccProjectResource4Config(resource ProjectAPIModel) string {
	return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_organization" "test-archive" {
  name        			= "%s"
}
resource "awx_project" "test-archive" {
  name         			= "%s"
  description  			= "%s"
  scm_type     			= "%s"
  scm_url      			= "%s"
  organization 			= awx_organization.test-archive.id
  scm_update_on_launch 	= %v
  timeout				= %d
}
  `, acctest.RandString(5), resource.Name, resource.Description, resource.ScmType, resource.ScmUrl, resource.ScmUpdOnLaunch, resource.Timeout))
}
