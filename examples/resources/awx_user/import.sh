# The password field contains values that are not returned by the automation controller API and thus not available in state.
# The first plan/apply after import will result in a modification to the password so that the state can be updated.

terraform import awx_user.example 1
