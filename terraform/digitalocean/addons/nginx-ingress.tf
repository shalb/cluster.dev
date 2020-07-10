#######################
# Deploy nginx-ingress
#######################

resource "null_resource" "nginx_ingress_install" {
  triggers = {
    config_contents = filemd5(var.config_path)
  }
  provisioner "local-exec" {
    command = "kubectl apply --kubeconfig ${var.config_path} -f 'https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v0.34.0/deploy/static/provider/do/deploy.yaml' && kubectl --kubeconfig ${var.config_path} delete -A ValidatingWebhookConfiguration ingress-nginx-admission"
  }
  provisioner "local-exec" {
    when = destroy
    # destroy ingress object to remove created Load Balancer
    command = "kubectl delete --kubeconfig ${var.config_path} -f 'https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v0.34.0/deploy/static/provider/do/deploy.yaml'"
  }
}
