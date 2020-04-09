data "digitalocean_kubernetes_versions" "k8s" {
  version_prefix = "${var.version}."
}
