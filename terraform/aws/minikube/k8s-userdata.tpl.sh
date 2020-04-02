# shellcheck disable=SC2148
# Add additional DNS since AWS could delay on VPC DNS resolving and cause Cert-Manager delays on cert creation
echo "nameserver 8.8.8.8" >> /etc/resolv.conf

# Install AWS utils
apt-get update
apt-get install -y wget unzip python3-pip
pip3 install awscli
export PATH=\$$PATH:/usr/local/bin

# Copy kubeconfig to S3 bucket
aws s3 cp /home/ubuntu/kubeconfig s3://${cluster_name}/kubeconfig_${cluster_name}

# Copy private ssh key to S3 bucket
cat << EOF > /home/ubuntu/.ssh/id_rsa
${private_key}
EOF
chmod 600 /home/ubuntu/.ssh/id_rsa
aws s3 cp /home/ubuntu/.ssh/id_rsa  s3://${cluster_name}/id_rsa_${cluster_name}.pem
