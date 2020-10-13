#######################
# Deploy nginx-ingress
#######################

resource "null_resource" "nginx_ingress_install" {
  triggers = {
    config_contents = filemd5(var.config_path)
  }
  provisioner "local-exec" {
    command = "kubectl apply --kubeconfig ${var.config_path} -f 'https://raw.githubusercontent.com/shalb/terraform-aws-minikube/master/addons/ingress.yaml'"
  }
  provisioner "local-exec" {
    when = destroy
    # destroy ingress object to remove created Load Balancer
    command = "exit 0"
  }
}
