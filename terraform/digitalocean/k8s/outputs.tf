output "cluster_status" {
  value = digitalocean_kubernetes_cluster.k8s.status
}

output "cluster_endpoint" {
  value = digitalocean_kubernetes_cluster.k8s.endpoint
}

output "kubernetes_config" {
  # If first symbol defines absolute path (/) - form fullname to config, if not - use provided
  value = substr(var.config_output_path, -1, 1) == "/" ? "${var.config_output_path}kubeconfig_${var.cluster_name}" : var.config_output_path
}
