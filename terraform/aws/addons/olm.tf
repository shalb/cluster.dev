# Install Operator Lifecycle Manager

locals {
  olm-version = "0.15.1"
}

resource "null_resource" "olm_install" {
  triggers = {
    config_contents = filemd5(var.config_path)
  }
  provisioner "local-exec" {
    command = "export KUBECONFIG=${var.config_path} && curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/${local.olm-version}/install.sh | bash -s ${local.olm-version}"
  }
}
