resource "awx_organization" "example" {
  name        = "example"
  description = "example"
}

data "awx_credential_type" "hashi_vault_secret_lookup" {
  name = "HashiCorp Vault Secret Lookup"
  kind = "external"
}

data "awx_credential_type" "source_control" {
  name = "Source Control"
  kind = "scm"
}

resource "awx_credential" "example_hashi_source_cred" {
  credential_type = data.awx_credential_type.hashi_vault_secret_lookup.id
  name            = "example_hashi_source_cred"
  organization    = awx_organization.example.id
  inputs = jsonencode({
    "api_version" : "v2",
    "cacert" : "",
    "client_cert_public" : "", # not available in AAP 2.4
    "client_cert_role" : "",   # not available in AAP 2.4
    "default_auth_path" : "approle",
    "kubernetes_role" : "",
    "namespace" : "fake",
    "role_id" : "fakse",
    "secret_id" : "fakevalue",
    "url" : "https://hcp-vault-private-vault-fake.z1.hashicorp.cloud:8200",
  "username" : "" })
}

resource "awx_credential" "example_hashi_target_cred" {
  credential_type = data.awx_credential_type.source_control.id
  name            = "example_hashi_target_cred"
  organization    = awx_organization.example.id
}

resource "awx_credential_input_sources" "example_hashi_cred_input_src" {
  description      = "Example 1"
  input_field_name = "ssh_key_data"
  metadata = {
    "auth_path" : ""
    "secret_key" : "testing_ssh_key"
    "secret_path" : "secret/deploy-keys"
    "secret_backend" : ""
    "secret_version" : ""
  }
  target_credential = awx_credential.example_hashi_target_cred.id
  source_credential = awx_credential.example_hashi_source_cred.id
}

resource "awx_credential_input_sources" "example_hashi_cred_input_src_2" {
  description      = "Example 2"
  input_field_name = "username"
  metadata = {
    "auth_path" : ""
    "secret_key" : "testing_username"
    "secret_path" : "secret/dev/deployinfo"
    "secret_backend" : ""
    "secret_version" : ""
  }
  target_credential = awx_credential.example_hashi_target_cred.id
  source_credential = awx_credential.example_hashi_source_cred.id
}
