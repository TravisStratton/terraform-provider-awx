---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "awx_job_template_instance_group Resource - awx"
subcategory: ""
description: |-
  The /api/v2/job_templates/{id}/instance_groups/ returns all instance_groups objects associated to the template. But, when asked to associate an instance_group or principle. Instead, the terraform schema stores a list of associated instance_groups. And, when creating or deleting or updated, it will make one api call PER list element. This allows the import function to work by only needing to pass in one job template ID to fill out the entire resource. If this was not done this way then when someone tries to to use the terraform plan -generate-config-out=./file.tf functionality it will create the resource block correctly. Otherwise, the -generate-config-out function would have to generate several resource blocks per template id and it's not set up to do that, per my current awareness. As I'm writing this provider specifically so we can use the -generate-config-out option, I felt this was worth the price of breaking this principle. The downside seems to be that this means if one of the list element's api calls succeeds, but a subsequent list element's fails, the success of the first element's call is not magially un-done. So you'll perpas have to use refresh state functions in tf cli to resolve.
---

# awx_job_template_instance_group (Resource)

The /api/v2/job_templates/{id}/instance_groups/ returns all instance_groups objects associated to the template. But, when asked to associate an instance_group or principle. Instead, the terraform schema stores a list of associated instance_groups. And, when creating or deleting or updated, it will make one api call PER list element. This allows the import function to work by only needing to pass in one job template ID to fill out the entire resource. If this was not done this way then when someone tries to to use the terraform plan -generate-config-out=./file.tf functionality it will create the resource block correctly. Otherwise, the -generate-config-out function would have to generate several resource blocks per template id and it's not set up to do that, per my current awareness. As I'm writing this provider specifically so we can use the -generate-config-out option, I felt this was worth the price of breaking this principle. The downside seems to be that this means if one of the list element's api calls succeeds, but a subsequent list element's fails, the success of the first element's call is not magially un-done. So you'll perpas have to use refresh state functions in tf cli to resolve.

## Example Usage

```terraform
resource "awx_job_template_instance_group" "default" {
  instance_groups_ids = [1]
  job_template_id     = 100
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `instance_groups_ids` (List of Number) An ordered list of instance_group IDs associated to a particular Job Template.
- `job_template_id` (String) The ID of the containing Job Template.

## Import

Import is supported using the following syntax:

```shell
terraform import awx_job_template_instance_group.example 100
```
