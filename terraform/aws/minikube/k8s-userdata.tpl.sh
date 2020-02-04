# Install AWS utils
yum install -y wget unzip python3-pip.noarch
pip3 install awscli
export PATH=\$$PATH:/usr/local/bin

# Copy kubeconfig to s3 bucket
aws s3 cp /home/centos/kubeconfig s3://${cluster_name}/kubeconfig_${cluster_name}

# Copy private ssh key to S3
cat << EOF > /home/centos/.ssh/id_rsa 
${private_key}
EOF
chmod 600 /home/centos/.ssh/id_rsa
aws s3 cp /home/centos/.ssh/id_rsa  s3://${cluster_name}/id_rsa_${cluster_name}.pem