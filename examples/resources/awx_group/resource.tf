resource "awx_organization" "example" {
  name        = "example"
  description = "example"
}

resource "awx_inventory" "example" {
  name         = "example"
  description  = "example"
  organization = awx_organization.example.id
}

resource "awx_group" "example-variables" {
  name        = "group-name-example"
  description = "Example with jsonencoded variables."
  inventory   = awx_inventory.example.id
  variables = jsonencode(
    {
      foo = "bar"
      baz = "qux"
    }
  )
}
