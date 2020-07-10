data "digitalocean_kubernetes_versions" "k8s" {
  version_prefix = "${var.k8s_version}."
}

resource "digitalocean_kubernetes_cluster" "k8s" {
  name     = var.cluster_name
  region   = var.region
  version  = data.digitalocean_kubernetes_versions.k8s.latest_version
  vpc_uuid = data.terraform_remote_state.vpc.outputs.vpc_id

  node_pool {
    name       = "${var.cluster_name}-worker-pool"
    size       = var.node_type
    auto_scale = true
    min_nodes  = var.min_node_count
    max_nodes  = var.max_node_count
  }
}

resource "local_file" "kubeconfig" {
  count                = var.write_kubeconfig ? 1 : 0
  content              = digitalocean_kubernetes_cluster.k8s.kube_config.0.raw_config
  filename             = substr(var.config_output_path, -1, 1) == "/" ? "${var.config_output_path}kubeconfig_${var.cluster_name}" : var.config_output_path
  file_permission      = "0644"
  directory_permission = "0755"
}
