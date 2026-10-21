# tflint configuration. It lives at the repository root because CI runs
# `tflint --init` here before `tflint --recursive`, and plugins are only
# installed from the config in the directory --init runs in.

config {
  call_module_type = "local"
}

plugin "terraform" {
  enabled = true
  preset  = "recommended"
}

plugin "aws" {
  enabled    = true
  deep_check = false
  version    = "0.49.0"
  source     = "github.com/terraform-linters/tflint-ruleset-aws"
}
