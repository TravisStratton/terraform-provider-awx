resource "awx_credential_type" "example" {
  name = "Example"
  kind = "cloud"
  inputs = jsonencode(
    {
      "fields" : [
        {
          "type" : "string",
          "id" : "username",
          "label" : "Username"
        },
        {
          "secret" : true,
          "type" : "string",
          "id" : "password",
          "label" : "Password"
        }
      ],
      "required" : ["username", "password"]
    }
  )
  injectors = jsonencode(
    {
      "env" : {
        "THIRD_PARTY_CLOUD_USERNAME" : "{{ username }}",
        "THIRD_PARTY_CLOUD_PASSWORD" : "{{ password }}"
      },
      "extra_vars" : {
        "some_extra_var" : "{{ username }}:{{ password }}"
      }
    }
  )
}

/* Format for inputs json:
{
  "fields": [{
    "id": "api_token",               # required - a unique name used to
                                     # reference the field value

    "label": "API Token",            # required - a unique label for the
                                     # field

    "help_text": "User-facing short text describing the field.",

    "type": ("string" | "boolean")   # defaults to 'string'

    "choices": ["A", "B", "C"]       # (only applicable to `type=string`)

    "format": "ssh_private_key"      # optional, can be used to enforce data
                                     # format validity for SSH private key
                                     # data (only applicable to `type=string`)

    "secret": true,                  # if true, the field value will be encrypted

    "multiline": false               # if true, the field should be rendered
                                     # as multi-line for input entry
                                     # (only applicable to `type=string`)
},{
    # field 2...
},{
    # field 3...
}],

"required": ["api_token"]            # optional; one or more fields can be marked as required
},
*/
