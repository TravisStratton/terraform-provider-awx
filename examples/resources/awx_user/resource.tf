resource "awx_user" "example" {
  username     = "example"
  password     = "test1234"
  first_name   = "test"
  last_name    = "user"
  email        = "test@example.com"
  is_superuser = true
}
