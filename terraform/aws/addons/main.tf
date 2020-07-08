# Check for kubeconfig and update resources when cluster gets re-created.
resource "null_resource" "kubeconfig_update" {
  triggers = {
    policy_sha1 = "${sha1(file(var.config_path))}"
  }
}
