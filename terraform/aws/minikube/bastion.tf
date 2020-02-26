# Generate private key to ssh to instance
resource "tls_private_key" "bastion_key" {
  algorithm = "RSA"
}
