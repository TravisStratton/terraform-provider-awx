<!-- SPECIAL: This readme should be altered to match each repo to which it is included.  -->
# Terraform Provider AWX (AWX, AAP2.4, AAP2.5)

This is a terraform provider for AWX, AAP2.4 & AAP2.5 built with the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) and based on the [Terraform Provider Scafolding Framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework).

If you find any bugs or have a feature request, please open a GitHub issue.

## Code sharing

This code is used for three different providers:

- The original one: TravisStratton/awx. This one supports awx, aap2.4, and aap2.5
- tfbrew/aap: Supports aap2.5 and greater.
- tfbrew/awx: Supports awx and aap2.4.

**This requires manual intervention & using Go build tags to make the code as simliar as possible accross all three repositories.**

## Build Tags

This repo has 2 build tags: repoAWX and repoAAP. This is so that this code can be used for 2 differently named repositories & Terraform Providers.

The scaffold template's GNUmakefile has been altered to include refencing these tags. Therefore, use the `make` commands to self-compile instead of just using `go` raw. For example, run `make install` instead fo `go install`.

## Special Handling for Each Repo

Search all files in this repository for the phrase `SPECIAL` to find files that may need to be updated to be specfiic to the containing repository & whether this is targetted for an aap named provider or an awx named provider.

## Writing Acceptance Tests

When writing acceptance test, you often have to write Terraform HCL code. Make sure to write your embedded HCL such that it will use the configprefix.Prefix to prefix the resource ID properly.

If you are creating functions to generate HCL, you can wrap the returned string in a function called **configprefix.ReplaceText()** to automatically convert the instances of awx_ or aap_ strings into the one matching your build tag.

For example:

```go
return configprefix.ReplaceText(fmt.Sprintf(`
resource "awx_credential_type" "test-name" {
  name         = "%s"
  description  = "%s"
}
data "awx_credential_type" "test-name" {
  name = awx_credential_type.test-name.name
  kind = awx_credential_type.test-name.kind
}
  `, resource.Name, resource.Description))
}
```
